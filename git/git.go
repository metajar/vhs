package git

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"vhs/devices"
)

type Git struct {
	RepoDir string
	Branch  string
	log     *zap.Logger
}

// NewGit creates a new Git object.
func NewGit(repoDir string, branch string) Git {
	err := os.MkdirAll(repoDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create git repository directory: %v", err)
	}

	_, err = os.Stat(filepath.Join(repoDir, ".git"))
	if os.IsNotExist(err) {
		cmd := exec.Command("git", "init")
		cmd.Dir = repoDir
		if err := cmd.Run(); err != nil {
			log.Fatalf("Failed to initialize git repository: %v", err)
		}
	}
	l, err := zap.NewProduction()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return Git{
		RepoDir: repoDir,
		Branch:  branch,
		log:     l,
	}
}

func (g *Git) SaveDeviceConfiguration(device devices.Device) error {
	deviceDir := filepath.Join(g.RepoDir, device.GetDeviceType())
	os.MkdirAll(deviceDir, os.ModePerm)

	deviceFile := filepath.Join(deviceDir, device.Name)
	timestamp := time.Now().Format(time.RFC3339)
	content := fmt.Sprintf("%s\n%s", timestamp, string(device.Payload))

	if err := ioutil.WriteFile(deviceFile, []byte(content), 0644); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond) // Add sleep before git add
	_, err := g.runGitCommand("add", deviceFile)
	if err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}
	time.Sleep(50 * time.Millisecond) // Add sleep before git commit
	output, err := g.runGitCommand("commit", "-m", fmt.Sprintf("Updated configuration for device %s", device.Name))
	if err != nil {
		// If the commit failed because there were no changes, ignore the error.
		if bytes.Contains(output, []byte("nothing to commit, working tree clean")) {
			return nil
		}
		return fmt.Errorf("git commit failed: %w, output: %s", err, output)
	}
	return nil

}

// commit commits changes in the Git repo.
func (g *Git) commit(filename string) error {
	_, err := g.runGitCommand("add", filename)
	if err != nil {
		return err
	}
	_, err = g.runGitCommand("commit", "-m", "Updated "+filename)
	if err != nil {
		return err
	}
	return nil
}

func (g *Git) Push() error {
	_, err := g.runGitCommand("push", "origin", g.Branch)
	return err
}

func (g *Git) Pull() error {
	_, err := g.runGitCommand("pull", "origin", g.Branch)
	return err
}

// Clone clones the remote Git repository.
func (g *Git) Clone(repoURL string) error {
	_, err := g.runGitCommand("clone", repoURL, g.RepoDir)
	if err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}
	return nil
}

func (g *Git) SetUpstreamBranch() error {
	_, err := g.runGitCommand("branch", "--set-upstream-to", "origin/"+g.Branch, g.Branch)
	if err != nil {
		return fmt.Errorf("failed to set upstream branch: %s, err: %s", g.Branch, err)
	}
	return nil
}

func (g *Git) StartPeriodicPush(ctx context.Context, duration time.Duration, maxAge time.Duration) {
	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := g.deprecateOldFiles(g.RepoDir, maxAge)
			if err != nil {
				g.log.Error("Failed to deprecate old files", zap.Error(err))
			}
			output, err := g.runGitCommand("push", "origin", g.Branch)
			if err != nil {
				g.log.Error("Failed to push changes", zap.Error(err))
			} else {
				if strings.Contains(string(output), "Everything up-to-date") {
					continue
				}
				g.log.Info("Pushed Changes Successfully!", zap.String("output", string(output)))
			}
		}
	}
}

func (g *Git) deprecateOldFiles(rootPath string, maxAge time.Duration) error {
	deprecatedFolderPath := filepath.Join(rootPath, "deprecated")
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip the deprecated folder and any files/folders within it.
		if strings.HasPrefix(path, deprecatedFolderPath) {
			if path == deprecatedFolderPath {
				return filepath.SkipDir
			}
			return nil
		}
		// Skip the .git folder and any files/folders within it.
		gitFolderPath := filepath.Join(rootPath, ".git")
		if strings.HasPrefix(path, gitFolderPath) {
			if path == gitFolderPath {
				return filepath.SkipDir
			}
			return nil
		}
		if !info.IsDir() {
			g.log.Info("Scanning file", zap.String("file", path))
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			scanner := bufio.NewScanner(file)
			if scanner.Scan() {
				timestampStr := scanner.Text()
				timestamp, err := time.Parse(time.RFC3339, timestampStr)
				if err != nil {
					g.log.Warn("Failed to parse timestamp for file", zap.String("file", path))
					return nil
				}
				if time.Since(timestamp) > maxAge {
					deprecatedPath := filepath.Join(rootPath, "deprecated", strings.TrimPrefix(path, rootPath))
					err := os.MkdirAll(filepath.Dir(deprecatedPath), os.ModePerm)
					if err != nil {
						return err
					}
					err = os.Rename(path, deprecatedPath)
					if err != nil {
						return err
					}
					g.log.Info("Deprecated file moved", zap.String("old", path), zap.String("new", deprecatedPath))
					_, err = g.runGitCommand("add", deprecatedPath)
					if err != nil {
						g.log.Error("Failed to add file", zap.String("file", deprecatedPath), zap.Error(err))
					}
					relativePath, err := filepath.Rel(rootPath, path)
					if err != nil {
						return err
					}
					_, err = g.runGitCommand("rm", relativePath)
					if err != nil {
						g.log.Error("Failed to remove deprecated file", zap.String("file", path), zap.Error(err))
					}
					_, err = g.runGitCommand("commit", "-m", fmt.Sprintf("Deprecating of file  %s", path))
					if err != nil {
						g.log.Error("Failed to ADD for deprecated file", zap.String("file", path), zap.Error(err))
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (g *Git) runGitCommand(args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.RepoDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("git command failed: %w, output: %s", err, output)
	}
	return output, nil
}
