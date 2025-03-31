package llm

import (
	"strings"

	"github.com/recrsn/git-ai/pkg/git"
)

// GenerateCommitMessage generates a commit message based on staged changes and commit history
// Note: This is a placeholder implementation. In a real implementation,
// this would call an LLM API to generate the commit message.
func GenerateCommitMessage(diff, recentCommits string) string {
	changedFiles := git.GetChangedFiles()
	files := strings.Split(changedFiles, "\n")
	var fileCount int
	for _, file := range files {
		if file != "" {
			fileCount++
		}
	}
	
	// Build a list of non-empty files for the title
	var nonEmptyFiles []string
	for _, file := range files {
		if file != "" {
			nonEmptyFiles = append(nonEmptyFiles, file)
		}
	}
	
	message := "Update"
	if len(nonEmptyFiles) > 0 {
		message = "Update " + strings.Join(nonEmptyFiles[:min(3, len(nonEmptyFiles))], ", ")
	}
	message += "\n\n"
	message += "Changed files:\n"
	for _, file := range files {
		if file != "" {
			message += "- " + file + "\n"
		}
	}
	
	// Note: In production, you would replace this with actual LLM call
	return message
}

// Helper function for min of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}