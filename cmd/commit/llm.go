package commit

import (
	"fmt"
	"github.com/recrsn/git-ai/pkg/llm"
	"strings"

	"github.com/recrsn/git-ai/pkg/config"
	"github.com/recrsn/git-ai/pkg/git"
	"github.com/recrsn/git-ai/pkg/logger"
)

// GenerateCommitMessage generates a commit message based on staged changes and commit history
func GenerateCommitMessage(cfg config.Config, diff, recentCommits string, useConventionalCommits bool, commitsWithDescriptions bool) (string, error) {
	// Use the LLM for commit message generation
	if cfg.APIKey == "" {
		return "", config.ErrLLMNotConfigured
	}

	logger.Debug("Using provider: %s, endpoint: %s, model: %s", cfg.Provider, cfg.Endpoint, cfg.Model)

	client, err := llm.NewClientWithProvider(cfg.Provider, cfg.Endpoint, cfg.APIKey)
	if err != nil {
		return "", fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Process diff with summarization if needed (32k token limit)
	processedDiff, isSummarized, err := git.ProcessDiffWithSummarization(cfg, diff, 32000)
	if err != nil {
		logger.Warn("Failed to process diff with summarization, using original: %v", err)
		processedDiff = diff
		isSummarized = false
	}

	// Get the changed files
	changedFiles := git.GetChangedFiles()

	// Get system and user prompts
	systemPrompt, err := llm.GetSystemPrompt(useConventionalCommits, commitsWithDescriptions, isSummarized)
	if err != nil {
		return "", fmt.Errorf("failed to build system prompt: %w", err)
	}

	userPrompt, err := llm.GetUserPrompt(processedDiff, changedFiles, recentCommits)
	if err != nil {
		return "", fmt.Errorf("failed to build user prompt: %w", err)
	}

	// Call the LLM API
	messages := []llm.Message{
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
