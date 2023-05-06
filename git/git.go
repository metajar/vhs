package git

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"vhs/devices"
)

type Git struct {
	RepoDir string
	Branch  string
}

// NewGit creates a new Git object.
func NewGit(repoDir string, branch string) Git {
	return Git{
		RepoDir: repoDir,
		Branch:  branch,
	}
}

func (g *Git) SaveDeviceConfiguration(device devices.Device) error {
	deviceDir := filepath.Join(g.RepoDir, device.GetDeviceType())
	os.MkdirAll(deviceDir, os.ModePerm)

	deviceFile := filepath.Join(deviceDir, device.Name)
	if err := ioutil.WriteFile(deviceFile, device.Payload, 0644); err != nil {
		return err
	}

	time.Sleep(50 * time.Millisecond) // Add sleep before git add

	cmd := exec.Command("git", "add", deviceFile)
	cmd.Dir = g.RepoDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git add failed: %w, output: %s", err, output)
	}

	time.Sleep(50 * time.Millisecond) // Add sleep before git commit

	cmd = exec.Command("git", "commit", "-m", fmt.Sprintf("Update configuration for device %s", device.Name))
	cmd.Dir = g.RepoDir
	output, err = cmd.CombinedOutput()
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
	cmd := exec.Command("git", "add", filename)
	cmd.Dir = g.RepoDir
	if _, err := cmd.Output(); err != nil {
		return err
	}
	cmd = exec.Command("git", "commit", "-m", "Update "+filename)
	cmd.Dir = g.RepoDir
	if _, err := cmd.Output(); err != nil {
		return err
	}
	return nil
}

// push pushes changes to the remote Git repo.
func (g *Git) Push() error {
	cmd := exec.Command("git", "push", "origin", g.Branch)
	cmd.Dir = g.RepoDir
	if _, err := cmd.Output(); err != nil {
		return err
	}
	return nil
}

// pull pulls the latest changes from the remote Git repo.
func (g *Git) Pull() error {
	cmd := exec.Command("git", "pull", "origin", g.Branch)
	cmd.Dir = g.RepoDir
	if _, err := cmd.Output(); err != nil {
		return err
	}
	return nil
}

// Clone clones the remote Git repository.
func (g *Git) Clone(repoURL string) error {
	cmd := exec.Command("git", "clone", repoURL, g.RepoDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w, output: %s", err, output)
	}
	return nil
}

func (g *Git) SetUpstreamBranch() error {
	cmd := exec.Command("git", "branch", "--set-upstream-to", "origin/"+g.Branch, g.Branch)
	cmd.Dir = g.RepoDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set upstream branch: %w, output: %s", err, output)
	}
	return nil
}

func (g *Git) StartPeriodicPush(ctx context.Context, duration time.Duration) {
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cmd := exec.Command("git", "status")
			cmd.Dir = g.RepoDir
			output, _ := cmd.CombinedOutput()
			log.Printf("Git status before pushing:\n%s", output)

			cmd = exec.Command("git", "log", "--oneline", "-n", "5") // Show the last 5 commits
			cmd.Dir = g.RepoDir
			output, _ = cmd.CombinedOutput()
			log.Printf("Commit history before pushing:\n%s", output)

			cmd = exec.Command("git", "push", "origin", g.Branch)
			cmd.Dir = g.RepoDir
			output, err := cmd.CombinedOutput() // Modify this line
			if err != nil {
				log.Printf("Failed to push changes: %v, output: %s", err, output)
			} else {
				log.Printf("Pushed changes successfully, output: %s", output) // Add this line
			}
			//if err := g.Push(); err != nil {
			//	log.Printf("Failed to push changes: %v", err)
			//}
		}
	}
}
