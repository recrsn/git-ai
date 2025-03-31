package llm

import (
	"strings"
	"testing"
)

func TestGenerateSimpleMessage(t *testing.T) {
	// Test the fallback simple message generation function
	message := generateSimpleMessage()

	// Verify it produces a non-empty message
	if message == "" {
		t.Errorf("Expected non-empty commit message, got empty string")
	}
	
	// Check that it contains "Changed files:"
	if !strings.Contains(message, "Changed files:") {
		t.Errorf("Expected commit message to contain 'Changed files:', got: %s", message)
	}
}

func TestGetUserPrompt(t *testing.T) {
	// Sample data for testing
	diff := "diff --git a/main.go b/main.go\nindex abc123..def456 100644\n--- a/main.go\n+++ b/main.go\n@@ -1,3 +1,4 @@\n package main\n \n+// This is a test comment"
	recentCommits := "Fix bug in user authentication\nUpdate README with installation instructions"
	changedFiles := "main.go\nREADME.md"
	
	// Test the prompt generation function
	prompt, err := GetUserPrompt(diff, changedFiles, recentCommits)
	if err != nil {
		t.Errorf("Failed to generate user prompt: %v", err)
	}
	
	// Verify the prompt contains all necessary parts
	if !strings.Contains(prompt, "# Changes (diff):") {
		t.Errorf("Expected prompt to contain '# Changes (diff):' section")
	}
	if !strings.Contains(prompt, "main.go") {
		t.Errorf("Expected prompt to contain 'main.go' in changed files")
	}
	if !strings.Contains(prompt, "Fix bug in user authentication") {
		t.Errorf("Expected prompt to contain recent commit message")
	}
}

func TestGetSystemPrompt(t *testing.T) {
	// Test conventional format
	conventionalPrompt := GetSystemPrompt(true)
	
	// Verify the system prompt is not empty
	if conventionalPrompt == "" {
		t.Errorf("Expected conventional system prompt to not be empty")
	}
	
	// Verify it contains key instructions
	if !strings.Contains(conventionalPrompt, "conventional commit format") {
		t.Errorf("Expected conventional system prompt to mention conventional commit format")
	}
	
	// Test standard format
	standardPrompt := GetSystemPrompt(false)
	
	// Verify the system prompt is not empty
	if standardPrompt == "" {
		t.Errorf("Expected standard system prompt to not be empty")
	}
	
	// Verify formats are different
	if conventionalPrompt == standardPrompt {
		t.Errorf("Expected conventional and standard prompts to be different")
	}
}