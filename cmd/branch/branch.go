package branch

import (
	"errors"
	"fmt"
	"github.com/recrsn/git-ai/pkg/config"
	"github.com/recrsn/git-ai/pkg/git"
	"github.com/recrsn/git-ai/pkg/llm"
	"github.com/recrsn/git-ai/pkg/logger"
	"github.com/recrsn/git-ai/pkg/ui"
	"os"
	"strings"
)

func executeBranch(description, diff string) {
	cfg := config.LoadConfigOrFatal()

	// Generate branch name - with spinner
	branchName, err := ui.WithSpinnerResult("Generating branch name with LLM...", func() (string, error) {
		return generateBranchNameWithDiff(cfg, description, diff)
	})
	if err != nil {
		if errors.Is(err, config.ErrLLMNotConfigured) {
			ui.PrintError("LLM endpoint or API key not configured. Please run 'git ai config' to set up.")
			os.Exit(1)
		}
		ui.PrintErrorf("Failed to generate branch name: %v", err)
		os.Exit(1)
	}

	// If auto-approve flag is not set, ask user to confirm or edit
	var proceed bool
	if !autoApprove {
		ui.DisplayBox("Generated Branch Name", branchName)

		options := []string{"Create branch", "Edit name", "Print name only", "Cancel"}
		selectedOption, err := ui.PromptForSelection(options, "Create branch", "What would you like to do?")
		if err != nil {
			logger.Fatal("Error prompting for selection: %v", err)
		}

		switch selectedOption {
		case "Create branch":
			proceed = true
		case "Edit name":
			// Use text input for editing with pre-filled value
			branchName, err = ui.PromptForInput("Edit branch name:", branchName)
			if err != nil {
				logger.Fatal("Error prompting for input: %v", err)
			}
			if branchName == "" {
				logger.Error("Branch name cannot be empty.")
				os.Exit(1)
			}
			proceed = true
		case "Print name only":
			ui.PrintMessage(branchName)
			os.Exit(0)
		case "Cancel":
			ui.PrintMessage("Branch creation cancelled.")
			os.Exit(0)
		}
	} else {
		proceed = true
	}

	if proceed {
		// Create the branch
		err = git.CreateBranch(branchName)
		if err != nil {
			logger.Fatal("Failed to create branch: %v", err)
		}

		ui.PrintMessagef("Branch '%s' created successfully!", branchName)
	}
}

// generateBranchNameWithDiff generates a branch name based on user input, diff, and existing branches
func generateBranchNameWithDiff(cfg config.Config, request, diff string) (string, error) {
	if cfg.Endpoint == "" || cfg.APIKey == "" {
		return "", config.ErrLLMNotConfigured
	}

	client, err := llm.NewClientWithProvider(cfg.Provider, cfg.Endpoint, cfg.APIKey)
	if err != nil {
		return "", fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Process diff with summarization if needed (32k token limit)
	processedDiff := ""
	isSummarized := false
	if diff != "" {
		var err error
		processedDiff, isSummarized, err = git.ProcessDiffWithSummarization(cfg, diff, 32000)
		if err != nil {
			logger.Warn("Failed to process diff with summarization, using original: %v", err)
			processedDiff = diff
			isSummarized = false
		}
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
	systemPrompt, err := llm.GetBranchSystemPrompt(isSummarized)
	if err != nil {
		return "", fmt.Errorf("failed to build system prompt: %w", err)
	}

	userPrompt, err := llm.GetBranchUserPrompt(request, localBranches, remoteBranches, processedDiff)
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
