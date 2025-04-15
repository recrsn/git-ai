package llm

import (
	"strings"
	"testing"
)

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

	// Test with empty diff
	emptyDiffPrompt, err := GetUserPrompt("", changedFiles, recentCommits)
	if err != nil {
		t.Errorf("Failed to generate user prompt with empty diff: %v", err)
	}
	if !strings.Contains(emptyDiffPrompt, "# Files changed:") {
		t.Errorf("Expected prompt with empty diff to still contain files changed section")
	}

	// Test with empty changed files
	emptyFilesPrompt, err := GetUserPrompt(diff, "", recentCommits)
	if err != nil {
		t.Errorf("Failed to generate user prompt with empty changed files: %v", err)
	}
	if strings.Contains(emptyFilesPrompt, "# Files changed:") && !strings.Contains(emptyFilesPrompt, "- ") {
		t.Errorf("Expected prompt with empty changed files to not have list items")
	}

	// Test with empty recent commits
	emptyCommitsPrompt, err := GetUserPrompt(diff, changedFiles, "")
	if err != nil {
		t.Errorf("Failed to generate user prompt with empty recent commits: %v", err)
	}
	if strings.Contains(emptyCommitsPrompt, "# Recent commit messages for context:") && !strings.Contains(emptyCommitsPrompt, "- ") {
		t.Errorf("Expected prompt with empty recent commits to not have list items")
	}
}

func TestGetSystemPrompt(t *testing.T) {
	// Test conventional format with descriptions
	conventionalPromptWithDesc, err := GetSystemPrompt(true, true)
	if err != nil {
		t.Errorf("Failed to generate system prompt: %v", err)
	}

	// Verify the system prompt is not empty
	if conventionalPromptWithDesc == "" {
		t.Errorf("Expected conventional system prompt with descriptions to not be empty")
	}

	// Verify it contains key instructions
	if !strings.Contains(conventionalPromptWithDesc, "conventional commit format") {
		t.Errorf("Expected conventional system prompt to mention conventional commit format")
	}

	// Test conventional format without descriptions (one-liner)
	conventionalPromptOneLiner, err := GetSystemPrompt(true, false)
	if err != nil {
		t.Errorf("Failed to generate system prompt: %v", err)
	}

	// Verify it contains one-liner instructions
	if !strings.Contains(conventionalPromptOneLiner, "one-line commit message") {
		t.Errorf("Expected conventional one-liner prompt to mention one-line commit message")
	}

	// Test standard format with descriptions
	standardPromptWithDesc, err := GetSystemPrompt(false, true)
	if err != nil {
		t.Errorf("Failed to generate system prompt: %v", err)
	}

	// Verify the system prompt is not empty
	if standardPromptWithDesc == "" {
		t.Errorf("Expected standard system prompt with descriptions to not be empty")
	}

	// Test standard format without descriptions (one-liner)
	standardPromptOneLiner, err := GetSystemPrompt(false, false)
	if err != nil {
		t.Errorf("Failed to generate system prompt: %v", err)
	}

	// Verify it contains one-liner instructions
	if !strings.Contains(standardPromptOneLiner, "one-line commit message") {
		t.Errorf("Expected standard one-liner prompt to mention one-line commit message")
	}

	// Verify formats are different
	if conventionalPromptWithDesc == standardPromptWithDesc {
		t.Errorf("Expected conventional and standard prompts with descriptions to be different")
	}

	// Verify one-liner vs. with-description formats are different
	if conventionalPromptWithDesc == conventionalPromptOneLiner {
		t.Errorf("Expected conventional prompts with and without descriptions to be different")
	}
}

func TestFormatAsList(t *testing.T) {
	// Test with multiple lines
	multiLine := "first\nsecond\nthird"
	formatted := formatAsList(multiLine)
	expected := "- first\n- second\n- third\n"
	if formatted != expected {
		t.Errorf("Expected %q but got %q", expected, formatted)
	}

	// Test with empty string
	emptyStr := ""
	emptyFormatted := formatAsList(emptyStr)
	if emptyFormatted != "" {
		t.Errorf("Expected empty string but got %q", emptyFormatted)
	}

	// Test with whitespace-only lines
	whitespaceLines := "first\n  \nthird"
	whitespaceFormatted := formatAsList(whitespaceLines)
	whitespaceExpected := "- first\n- third\n"
	if whitespaceFormatted != whitespaceExpected {
		t.Errorf("Expected %q but got %q", whitespaceExpected, whitespaceFormatted)
	}

	// Test with leading/trailing whitespace
	whitespaceInput := "  first  \n  second  "
	whitespaceOutput := formatAsList(whitespaceInput)
	whitespaceOutputExpected := "- first\n- second\n"
	if whitespaceOutput != whitespaceOutputExpected {
		t.Errorf("Expected %q but got %q", whitespaceOutputExpected, whitespaceOutput)
	}
}

func TestGetBranchSystemPrompt(t *testing.T) {
	// Test that the branch system prompt is not empty
	branchSystemPrompt := GetBranchSystemPrompt()
	if branchSystemPrompt == "" {
		t.Errorf("Expected branch system prompt to not be empty")
	}

	// Verify it contains key instructions for branch naming
	if !strings.Contains(branchSystemPrompt, "branch name") {
		t.Errorf("Expected branch system prompt to mention branch naming")
	}
}

func TestGetBranchUserPrompt(t *testing.T) {
	// Sample data for testing
	request := "Create a branch for fixing the login bug"
	localBranches := []string{"main", "dev", "feature/auth"}
	remoteBranches := []string{"origin/main", "origin/dev"}

	// Test the branch prompt generation function
	prompt, err := GetBranchUserPrompt(request, localBranches, remoteBranches)
	if err != nil {
		t.Errorf("Failed to generate branch user prompt: %v", err)
	}

	// Verify the prompt contains all necessary parts
	if !strings.Contains(prompt, request) {
		t.Errorf("Expected branch prompt to contain the user request")
	}
	if !strings.Contains(prompt, "main") {
		t.Errorf("Expected branch prompt to list local branches")
	}
	if !strings.Contains(prompt, "origin/main") {
		t.Errorf("Expected branch prompt to list remote branches")
	}

	// Test with empty branches
	emptyBranchPrompt, err := GetBranchUserPrompt(request, []string{}, []string{})
	if err != nil {
		t.Errorf("Failed to generate branch user prompt with empty branches: %v", err)
	}
	if !strings.Contains(emptyBranchPrompt, request) {
		t.Errorf("Expected branch prompt with empty branches to still contain the request")
	}
}
