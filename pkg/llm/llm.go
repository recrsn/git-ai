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

	// Process diff with summarization if needed (32k token limit)
	processedDiff, isSummarized, err := ProcessDiffWithSummarization(cfg, diff, 32000)
	if err != nil {
		logger.Warn("Failed to process diff with summarization, using original: %v", err)
		processedDiff = diff
		isSummarized = false
	}

	// Get the changed files
	changedFiles := git.GetChangedFiles()

	// Get system and user prompts
	systemPrompt, err := GetSystemPrompt(useConventionalCommits, commitsWithDescriptions, isSummarized)
	if err != nil {
		return "", fmt.Errorf("failed to build system prompt: %w", err)
	}

	userPrompt, err := GetUserPrompt(processedDiff, changedFiles, recentCommits)
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
