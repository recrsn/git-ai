package llm

import (
	"strings"
	"testing"
)

func TestGenerateCommitMessage(t *testing.T) {
	// Sample diff and commit history for testing
	diff := "diff --git a/main.go b/main.go\nindex abc123..def456 100644\n--- a/main.go\n+++ b/main.go\n@@ -1,3 +1,4 @@\n package main\n \n+// This is a test comment"
	recentCommits := "Fix bug in user authentication\nUpdate README with installation instructions\nAdd new feature for notifications"

	// Test the generate commit message function
	message := GenerateCommitMessage(diff, recentCommits)

	// Verify it produces a non-empty message
	if message == "" {
		t.Errorf("Expected non-empty commit message, got empty string")
	}
	
	// Check that it contains "Changed files:"
	if !strings.Contains(message, "Changed files:") {
		t.Errorf("Expected commit message to contain 'Changed files:', got: %s", message)
	}
}