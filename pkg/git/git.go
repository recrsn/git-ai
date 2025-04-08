package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
func CreateCommit(message string, amend bool) error {
	args := []string{"commit", "-m", message}
	if amend {
		args = append(args, "--amend")
	}
	cmd := exec.Command("git", args...)

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

// GetPreferredEditor returns the editor to use based on git-ai config, git configuration, or environment variables
func GetPreferredEditor() string {
	// First try git-ai specific config
	editor, err := GetConfig("git-ai.editor")
	if err == nil && editor != "" {
		return editor
	}

	// Then try git's core.editor config
	editor, err = GetConfig("core.editor")
	if err == nil && editor != "" {
		return editor
	}

	// Then try environment variables
	if envEditor := strings.TrimSpace(os.Getenv("GIT_EDITOR")); envEditor != "" {
		return envEditor
	}

	if envEditor := strings.TrimSpace(os.Getenv("EDITOR")); envEditor != "" {
		return envEditor
	}

	if envEditor := strings.TrimSpace(os.Getenv("VISUAL")); envEditor != "" {
		return envEditor
	}

	// Default editors based on platform
	return "vi" // Fallback to vi as it's commonly available
}

// EditWithExternalEditor opens the user's preferred editor to edit the provided content
// and returns the edited content
func EditWithExternalEditor(initialContent string) (string, error) {
	editor := GetPreferredEditor()

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "git-ai-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}

	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // Clean up the file when done

	// Write the initial content to the file
	if _, err := tmpFile.WriteString(initialContent); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write to temporary file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Prepare the editor command
	var cmd *exec.Cmd
	editorParts := strings.Fields(editor)
	if len(editorParts) > 1 {
		cmd = exec.Command(editorParts[0], append(editorParts[1:], tmpPath)...)
	} else {
		cmd = exec.Command(editor, tmpPath)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Open the editor
	logger.Debug("Opening editor: %s %s", editor, tmpPath)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error running editor: %w", err)
	}

	// Read back the edited content
	editedContent, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", fmt.Errorf("failed to read edited file: %w", err)
	}

	return string(editedContent), nil
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
