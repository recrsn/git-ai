package git

import (
	"bytes"
	"fmt"
	"os/exec"
)

// HasStagedChanges checks if there are any staged changes in the git repository
func HasStagedChanges() bool {
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	err := cmd.Run()
	// Exit status 1 means there are changes
	return err != nil
}

// GetStagedDiff returns the diff of staged changes
func GetStagedDiff() string {
	cmd := exec.Command("git", "diff", "--cached")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error getting staged changes: %v\n", err)
		return ""
	}
	return out.String()
}

// GetRecentCommits returns the recent commit messages
func GetRecentCommits() string {
	cmd := exec.Command("git", "log", "-n", "5", "--pretty=format:%s")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error getting recent commits: %v\n", err)
		return ""
	}
	return out.String()
}

// GetChangedFiles returns a list of staged files
func GetChangedFiles() string {
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error getting changed files: %v\n", err)
		return ""
	}
	return out.String()
}

// CreateCommit creates a git commit with the given message
func CreateCommit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error creating commit: %v: %s", err, stderr.String())
	}
	return nil
}