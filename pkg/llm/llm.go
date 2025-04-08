package llm

import (
	"fmt"
	"github.com/recrsn/git-ai/pkg/config"
	"github.com/recrsn/git-ai/pkg/logger"
	"strings"

	"github.com/recrsn/git-ai/pkg/git"
)

// GenerateCommitMessage generates a commit message based on staged changes and commit history
func GenerateCommitMessage(cfg config.Config, diff, recentCommits string, useConventionalCommits bool, commitsWithDescriptions bool) string {
	// Try to use the LLM for commit message generation
	message, err := generateWithLLM(cfg, diff, recentCommits, useConventionalCommits, commitsWithDescriptions)
	if err != nil {
		logger.Warn("Failed to generate commit message with LLM: %v\n", err)
		logger.Warn("Falling back to simple commit message generation")
		return generateSimpleMessage()
	}

	return message
}

// generateWithLLM uses an LLM to generate a commit message
func generateWithLLM(cfg config.Config, diff, recentCommits string, useConventionalCommits bool, commitsWithDescriptions bool) (string, error) {
	if cfg.Endpoint == "" || cfg.APIKey == "" {
		return "", fmt.Errorf("LLM endpoint or API key not configured")
	}

	client, err := NewClient(cfg.Endpoint, cfg.APIKey)
	if err != nil {
		return "", fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Get the changed files
	changedFiles := git.GetChangedFiles()

	// Get system and user prompts
	systemPrompt := GetSystemPrompt(useConventionalCommits, commitsWithDescriptions)
	userPrompt, err := GetUserPrompt(diff, changedFiles, recentCommits)
	if err != nil {
		return "", fmt.Errorf("failed to build prompt: %w", err)
	}

	// Call the LLM API
	messages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: userPrompt,
		},
	}

	response, err := client.ChatCompletion(cfg.Model, messages)
	if err != nil {
		return "", fmt.Errorf("failed to get completion: %w", err)
	}

	// Clean up the response
	return strings.TrimSpace(response), nil
}

// generateSimpleMessage generates a simple commit message without using an LLM
// This is used as a fallback when the LLM call fails
func generateSimpleMessage() string {
	changedFiles := git.GetChangedFiles()
	files := strings.Split(changedFiles, "\n")

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

	return message
}

// GenerateBranchName generates a branch name based on user input and existing branches
func GenerateBranchName(cfg config.Config, request string) (string, error) {
	if cfg.Endpoint == "" || cfg.APIKey == "" {
		return "", fmt.Errorf("LLM endpoint or API key not configured")
	}

	client, err := NewClient(cfg.Endpoint, cfg.APIKey)
	if err != nil {
		return "", fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Get lists of existing branches
	localBranches, err := git.GetLocalBranches()
	if err != nil {
		logger.Warn("Failed to get local branches: %v", err)
		localBranches = []string{}
	}

	remoteBranches, err := git.GetRemoteBranches()
	if err != nil {
		logger.Warn("Failed to get remote branches: %v", err)
		remoteBranches = []string{}
	}

	// Get system and user prompts
	systemPrompt := GetBranchSystemPrompt()
	userPrompt, err := GetBranchUserPrompt(request, localBranches, remoteBranches)
	if err != nil {
		return "", fmt.Errorf("failed to build prompt: %w", err)
	}

	// Call the LLM API
	messages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: userPrompt,
		},
	}

	response, err := client.ChatCompletion(cfg.Model, messages)
	if err != nil {
		return "", fmt.Errorf("failed to get completion: %w", err)
	}

	// Clean up the response
	branchName := strings.TrimSpace(response)

	// Ensure it doesn't contain any invalid characters
	branchName = sanitizeBranchName(branchName)

	return branchName, nil
}

// sanitizeBranchName ensures the branch name follows Git conventions
func sanitizeBranchName(name string) string {
	// Replace spaces with hyphens
	name = strings.ReplaceAll(name, " ", "-")

	// Remove any Git-unfriendly characters
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '/' || r == '_' || r == '.' {
			return r
		}
		return '-'
	}, name)

	// Convert multiple hyphens to a single one
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}

	// Trim hyphens from the start and end
	name = strings.Trim(name, "-")

	return name
}
