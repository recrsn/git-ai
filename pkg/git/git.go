package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/recrsn/git-ai/pkg/logger"
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
		logger.Error("Error getting staged changes: %v", err)
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
		logger.Error("Error getting recent commits: %v", err)
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
		logger.Error("Error getting changed files: %v", err)
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
		logger.Error("Error creating commit: %v: %s", err, stderr.String())
		return fmt.Errorf("error creating commit: %v: %s", err, stderr.String())
	}
	return nil
}

// UsesConventionalCommits checks if the repository already uses conventional commits
func UsesConventionalCommits() bool {
	// Get the last 30 commits to check the pattern
	cmd := exec.Command("git", "log", "-n", "30", "--pretty=format:%s")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		// If command fails (maybe new repo), just return false
		logger.Debug("Error checking conventional commits: %v", err)
		return false
	}

	// Define regex pattern for conventional commits
	conventionalPattern := regexp.MustCompile(`^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-zA-Z0-9_-]+\))?: .+`)

	// Count commits following the convention
	commits := strings.Split(out.String(), "\n")
	conventionalCount := 0
	totalCount := 0

	for _, commit := range commits {
		if commit == "" {
			continue
		}
		totalCount++
		if conventionalPattern.MatchString(commit) {
			conventionalCount++
		}
	}

	// If more than 50% of commits follow convention, return true
	if totalCount > 0 && conventionalCount*100/totalCount >= 50 {
		return true
	}

	return false
}

// GetConfig gets the value of a git config entry
func GetConfig(key string) (string, error) {
	cmd := exec.Command("git", "config", "--get", key)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		// Not found or error
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

// SetConfig sets a git config value
func SetConfig(key, value string) error {
	cmd := exec.Command("git", "config", key, value)
	err := cmd.Run()
	if err != nil {
		logger.Error("Error setting git config %s: %v", key, err)
		return fmt.Errorf("error setting git config %s: %v", key, err)
	}
	return nil
}
