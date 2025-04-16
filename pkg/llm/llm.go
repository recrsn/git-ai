package llm

import (
	"errors"
	"fmt"
	"github.com/recrsn/git-ai/pkg/config"
	"github.com/recrsn/git-ai/pkg/logger"
	"strings"

	"github.com/recrsn/git-ai/pkg/git"
)

// GenerateCommitMessage generates a commit message based on staged changes and commit history
func GenerateCommitMessage(cfg config.Config, diff, recentCommits string, useConventionalCommits bool, commitsWithDescriptions bool) (string, error) {
	// Use the LLM for commit message generation
	message, err := generateWithLLM(cfg, diff, recentCommits, useConventionalCommits, commitsWithDescriptions)
	if err != nil {
		return "", fmt.Errorf("failed to generate commit message with LLM: %w", err)
	}

	return message, nil
}

// ErrLLMNotConfigured is returned when the LLM endpoint or API key is not configured
var ErrLLMNotConfigured = errors.New("LLM endpoint or API key not configured")

// generateWithLLM uses an LLM to generate a commit message
func generateWithLLM(cfg config.Config, diff, recentCommits string, useConventionalCommits bool, commitsWithDescriptions bool) (string, error) {
	if cfg.Endpoint == "" || cfg.APIKey == "" {
		return "", ErrLLMNotConfigured
	}

	client, err := NewClient(cfg.Endpoint, cfg.APIKey)
	if err != nil {
		return "", fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Get the changed files
	changedFiles := git.GetChangedFiles()

	// Get system and user prompts
	systemPrompt, err := GetSystemPrompt(useConventionalCommits, commitsWithDescriptions)
	if err != nil {
		return "", fmt.Errorf("failed to build system prompt: %w", err)
	}

	userPrompt, err := GetUserPrompt(diff, changedFiles, recentCommits)
	if err != nil {
		return "", fmt.Errorf("failed to build user prompt: %w", err)
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

// GenerateBranchName generates a branch name based on user input and existing branches
func GenerateBranchName(cfg config.Config, request string) (string, error) {
	if cfg.Endpoint == "" || cfg.APIKey == "" {
		return "", ErrLLMNotConfigured
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
