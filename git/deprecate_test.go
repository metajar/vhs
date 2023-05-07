package git

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDeprecateOldFiles(t *testing.T) {
	// Create a temporary directory to act as a Git repository
	tempDir, err := ioutil.TempDir("", "git_test_repo")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize a new Git object
	gitObj := NewGit(tempDir, "master")

	// Create test files with different timestamps
	file1 := filepath.Join(tempDir, "file1")
	file2 := filepath.Join(tempDir, "file2")

	// Create file1 with a timestamp older than 24 hours
	err = createTestFile(file1, time.Now().Add(-25*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create file2 with a timestamp within the past 24 hours
	err = createTestFile(file2, time.Now().Add(-23*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := gitObj.commit(file1); err != nil {
		t.Fatalf("Failed to add and commit file1: %v", err)
	}
	if err := gitObj.commit(file2); err != nil {
		t.Fatalf("Failed to add and commit file2: %v", err)
	}

	// Run the deprecateOldFiles function with a max age of 24 hours
	err = gitObj.deprecateOldFiles(tempDir, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to deprecate old files: %v", err)
	}

	// Check if file1 was moved to the deprecated folder
	deprecatedFile1 := filepath.Join(tempDir, "deprecated", "file1")
	if _, err := os.Stat(deprecatedFile1); os.IsNotExist(err) {
		t.Errorf("File1 should have been deprecated, but it was not: %v", err)
	}

	// Check if file2 is still present in the original location
	if _, err := os.Stat(file2); os.IsNotExist(err) {
		t.Errorf("File2 should not have been deprecated, but it was: %v", err)
	}
}

func createTestFile(path string, timestamp time.Time) error {
	content := timestamp.Format(time.RFC3339) + "\nTest content"
	return ioutil.WriteFile(path, []byte(content), 0644)
}
