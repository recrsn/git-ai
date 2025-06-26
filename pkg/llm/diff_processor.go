package llm

import (
	"fmt"
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

// SummarizeDiff creates a concise summary of changes in a diff
func SummarizeDiff(cfg config.Config, fileDiff FileDiff) (string, error) {
	if cfg.Endpoint == "" || cfg.APIKey == "" {
		return "", ErrLLMNotConfigured
	}

	client, err := NewClient(cfg.Endpoint, cfg.APIKey)
	if err != nil {
		return "", fmt.Errorf("failed to create LLM client: %w", err)
	}

	systemPrompt := GetDiffSummarySystemPrompt()
	userPrompt := fmt.Sprintf("Summarize the changes in this diff:\n\n```diff\n%s\n```", fileDiff.Content)

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
		return "", fmt.Errorf("failed to get diff summary: %w", err)
	}

	return strings.TrimSpace(response), nil
}

// ProcessDiffWithSummarization handles large diffs by summarizing files in parallel
func ProcessDiffWithSummarization(cfg config.Config, diff string, tokenLimit int) (string, error) {
	// If diff is small enough, return as-is
	if EstimateTokens(diff) <= tokenLimit {
		return diff, nil
	}

	logger.Debug("Diff exceeds token limit (%d tokens estimated), summarizing by file", EstimateTokens(diff))

	// Parse diff by file
	fileDiffs := ParseDiffByFile(diff)
	if len(fileDiffs) == 0 {
		return diff, nil // Return original if parsing fails
	}

	// Process files in parallel
	var wg sync.WaitGroup
	summaries := make([]string, len(fileDiffs))
	errors := make([]error, len(fileDiffs))

	for i, fileDiff := range fileDiffs {
		wg.Add(1)
		go func(index int, fd FileDiff) {
			defer wg.Done()

			summary, err := SummarizeDiff(cfg, fd)
			if err != nil {
				logger.Warn("Failed to summarize diff for %s: %v", fd.Path, err)
				errors[index] = err
				// Use truncated version as fallback
				content := fd.Content
				if len(content) > 500 {
					content = content[:500] + "... (truncated)"
				}
				summaries[index] = fmt.Sprintf("File: %s\n%s", fd.Path, content)
			} else if summary == "MINOR CHANGES ONLY" {
				// Skip formatting-only changes
				summaries[index] = ""
			} else {
				summaries[index] = fmt.Sprintf("File: %s\nSummary: %s", fd.Path, summary)
			}
		}(i, fileDiff)
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
		return "Minor formatting and refactoring changes with no functional impact.", nil
	}

	// Combine summaries
	result := "# Summarized Changes\n\n" + strings.Join(filteredSummaries, "\n\n")

	logger.Debug("Summarized diff: %d tokens (from %d tokens)", EstimateTokens(result), EstimateTokens(diff))

	return result, nil
}
