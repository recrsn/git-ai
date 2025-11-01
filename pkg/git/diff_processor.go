package git

import (
	"fmt"
	"github.com/recrsn/git-ai/pkg/llm"
	"regexp"
	"strings"
	"sync"

	"github.com/recrsn/git-ai/pkg/config"
	"github.com/recrsn/git-ai/pkg/logger"
)

// EstimateTokens provides a rough estimate of tokens in text
// Uses approximation: 1 token ~= 4 characters for most text
func EstimateTokens(text string) int {
	return len(text) / 4
}

// FileDiff represents a single file's diff
type FileDiff struct {
	Path    string
	Content string
}

// ParseDiffByFile splits a unified diff into individual file diffs
func ParseDiffByFile(diff string) []FileDiff {
	var files []FileDiff

	// Split by file headers (lines starting with "diff --git")
	parts := regexp.MustCompile(`(?m)^diff --git`).Split(diff, -1)

	// Skip the first empty part
	for i := 1; i < len(parts); i++ {
		part := "diff --git" + parts[i]

		// Extract file path from the header
		lines := strings.Split(part, "\n")
		var path string
		for _, line := range lines {
			if strings.HasPrefix(line, "+++ b/") {
				path = strings.TrimPrefix(line, "+++ b/")
				break
			} else if strings.HasPrefix(line, "+++ ") && !strings.Contains(line, "/dev/null") {
				path = strings.TrimPrefix(line, "+++ ")
				break
			}
		}

		if path == "" {
			// Try to extract from diff header
			for _, line := range lines {
				if strings.HasPrefix(line, "diff --git a/") {
					parts := strings.Fields(line)
					if len(parts) >= 4 {
						path = strings.TrimPrefix(parts[3], "b/")
					}
					break
				}
			}
		}

		files = append(files, FileDiff{
			Path:    path,
			Content: part,
		})
	}

	return files
}

// createFileBatches groups files into batches where each batch doesn't exceed token limit
func createFileBatches(fileDiffs []FileDiff, tokenLimit int) [][]FileDiff {
	var batches [][]FileDiff
	var currentBatch []FileDiff
	currentBatchTokens := 0

	for _, fileDiff := range fileDiffs {
		fileTokens := EstimateTokens(fileDiff.Content)

		// If single file exceeds limit, put it in its own batch
		if fileTokens > tokenLimit {
			if len(currentBatch) > 0 {
				batches = append(batches, currentBatch)
				currentBatch = nil
				currentBatchTokens = 0
			}
			batches = append(batches, []FileDiff{fileDiff})
			continue
		}

		// If adding this file would exceed limit, start new batch
		if currentBatchTokens+fileTokens > tokenLimit && len(currentBatch) > 0 {
			batches = append(batches, currentBatch)
			currentBatch = nil
			currentBatchTokens = 0
		}

		currentBatch = append(currentBatch, fileDiff)
		currentBatchTokens += fileTokens
	}

	// Add final batch if not empty
	if len(currentBatch) > 0 {
		batches = append(batches, currentBatch)
	}

	return batches
}

// summarizeBatch summarizes a batch of file diffs together
func summarizeBatch(cfg config.Config, fileBatch []FileDiff) (string, error) {
	if cfg.Endpoint == "" || cfg.APIKey == "" {
		return "", config.ErrLLMNotConfigured
	}

	client, err := llm.NewClientWithProvider(cfg.Provider, cfg.Endpoint, cfg.APIKey)
	if err != nil {
		return "", fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Combine all files in batch into single diff
	var combinedContent strings.Builder
	for i, fileDiff := range fileBatch {
		if i > 0 {
			combinedContent.WriteString("\n\n")
		}
		combinedContent.WriteString(fileDiff.Content)
	}

	systemPrompt := llm.GetDiffSummarySystemPrompt()
	userPrompt := fmt.Sprintf("Summarize the changes in this diff:\n\n```diff\n%s\n```", combinedContent.String())

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
		return "", fmt.Errorf("failed to get batch summary: %w", err)
	}

	summary := strings.TrimSpace(response)
	return summary, nil
}

// ProcessDiffWithSummarization handles large diffs by summarizing files in parallel
// Returns the processed diff, a boolean indicating if summarization occurred, and any error
func ProcessDiffWithSummarization(cfg config.Config, diff string, tokenLimit int) (string, bool, error) {
	// If diff is small enough, return as-is
	if EstimateTokens(diff) <= tokenLimit {
		return diff, false, nil
	}

	logger.Debug("Diff exceeds token limit (%d tokens estimated), summarizing by file", EstimateTokens(diff))

	// Parse diff by file
	fileDiffs := ParseDiffByFile(diff)
	if len(fileDiffs) == 0 {
		return diff, false, nil // Return original if parsing fails
	}

	// Create batches of files to process together
	batches := createFileBatches(fileDiffs, tokenLimit)

	// Process batches in parallel with limited concurrency
	semaphore := make(chan struct{}, 4) // Limit to 4 concurrent requests
	var wg sync.WaitGroup
	summaries := make([]string, len(batches))
	errors := make([]error, len(batches))

	for i, batch := range batches {
		wg.Add(1)
		go func(index int, fileBatch []FileDiff) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }() // Release semaphore

			summary, err := summarizeBatch(cfg, fileBatch)
			if err != nil {
				logger.Warn("Failed to summarize batch %d: %v", index, err)
				errors[index] = err
				// Use truncated version as fallback
				var fallbackParts []string
				for _, fd := range fileBatch {
					content := fd.Content
					if len(content) > 500 {
						content = content[:500] + "... (truncated)"
					}
					fallbackParts = append(fallbackParts, fmt.Sprintf("File: %s\n%s", fd.Path, content))
				}
				summaries[index] = strings.Join(fallbackParts, "\n\n")
			} else if summary == "MINOR CHANGES ONLY" {
				// Skip formatting-only changes
				summaries[index] = ""
			} else {
				summaries[index] = summary
			}
		}(i, batch)
	}

	wg.Wait()

	// Filter out empty summaries (formatting-only changes)
	var filteredSummaries []string
	for _, summary := range summaries {
		if summary != "" {
			filteredSummaries = append(filteredSummaries, summary)
		}
	}

	// If all changes were formatting-only, return a simple message
	if len(filteredSummaries) == 0 {
		return "Minor formatting and refactoring changes with no functional impact.", true, nil
	}

	// Combine summaries
	result := "# Summarized Changes\n\n" + strings.Join(filteredSummaries, "\n\n")

	logger.Debug("Summarized diff: %d tokens (from %d tokens)", EstimateTokens(result), EstimateTokens(diff))

	return result, true, nil
}
