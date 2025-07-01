package commit

import (
	"errors"
	"os"

	"github.com/recrsn/git-ai/pkg/config"
	"github.com/recrsn/git-ai/pkg/git"
	"github.com/recrsn/git-ai/pkg/logger"
	"github.com/recrsn/git-ai/pkg/ui"
	"github.com/spf13/cobra"
)

var (
	autoApprove             bool
	conventionalCommits     bool
	noConventionalCommits   bool
	commitsWithDescriptions bool
	amendCommit             bool
)

// Cmd represents the commit command
var Cmd = &cobra.Command{
	Use:   "commit",
	Short: "Generate an AI commit message based on staged changes",
	Long:  `Analyzes your staged changes and git history to generate a descriptive commit message.`,
	Run: func(cmd *cobra.Command, args []string) {
		executeCommit()
	},
}

func init() {
	Cmd.Flags().BoolVar(&autoApprove, "auto", false, "Automatically approve the generated commit message without prompting")
	Cmd.Flags().BoolVar(&conventionalCommits, "conventional", false, "Use conventional commit format (type(scope): description)")
	Cmd.Flags().BoolVar(&noConventionalCommits, "no-conventional", false, "Don't use conventional commit format")
	Cmd.Flags().BoolVar(&commitsWithDescriptions, "with-descriptions", false, "Generate commit messages with detailed descriptions")
	Cmd.Flags().BoolVarP(&amendCommit, "amend", "a", false, "Amend the previous commit instead of creating a new one")
}

func executeCommit() {
	cfg := config.LoadConfigOrFatal()

	// Check if there are staged changes
	if !git.HasStagedChanges() {
		if amendCommit {
			ui.PrintMessage("No staged changes found. Will amend the previous commit message only.")
		} else {
			ui.PrintErrorf("No staged changes found. Please stage your changes with 'git add' first.")
			os.Exit(1)
		}
	}

	// If the --with-descriptions flag wasn't explicitly set, check git config
	if !commitsWithDescriptions {
		value, err := git.GetConfig("git-ai.commitsWithDescriptions")
		if err == nil && value == "true" {
			commitsWithDescriptions = true
		}
	}

	// Get the staged changes diff, filtering out generated files
	diff := git.GetStagedDiffFiltered()
	if diff == "" {
		logger.Fatal("Could not retrieve diff of staged changes.")
	}

	// Get recent commit history
	recentCommits := git.GetRecentCommits()

	// Determine whether to use conventional commits format
	useConventionalCommits := shouldUseConventionalCommits()

	// Generate commit message based on staged changes and history - with spinner
	message, err := ui.WithSpinnerResult("Generating commit message with LLM...", func() (string, error) {
		return GenerateCommitMessage(cfg, diff, recentCommits, useConventionalCommits, commitsWithDescriptions)
	})
	if err != nil {
		if errors.Is(err, config.ErrLLMNotConfigured) {
			ui.PrintError("LLM endpoint or API key not configured. Please run 'git ai config' to set up.")
			os.Exit(1)
		}
		logger.Fatal("Failed to generate commit message: %v", err)
	}

	// If auto-approve flag is not set, ask user to confirm or edit
	var proceed bool
	if !autoApprove {
		message, proceed = ui.PromptForConfirmation(message)
		if !proceed {
			os.Exit(0)
		}
	}

	// If the user explicitly chose a commit format, save the preference
	if conventionalCommits || noConventionalCommits {
		saveCommitFormatPreference(conventionalCommits)
	}

	// Save commit description preference if the flag was explicitly provided
	if commitsWithDescriptions {
		saveCommitDescriptionPreference(true)
	}

	// Create the commit with the message
	err = git.CreateCommit(message, amendCommit)
	if err != nil {
		logger.Fatal("Failed to create commit: %v", err)
	}

	// Get the commit hash for logging
	commitHash, err := git.GetLatestCommitHash()
	if err != nil {
		logger.Warn("Failed to get commit hash for logging: %v", err)
	} else {
		if commitHash != "" {
			logger.Debug("Commit created: %s", commitHash)
		}
	}

	if amendCommit {
		ui.PrintSuccess("Commit amended successfully!")
	} else {
		ui.PrintSuccess("Commit created successfully!")
	}
}

// shouldUseConventionalCommits determines whether to use conventional commit format
// based on command-line flags, git config, and repository history
func shouldUseConventionalCommits() bool {
	// Command line flags take precedence
	if conventionalCommits {
		return true
	}
	if noConventionalCommits {
		return false
	}

	// Check git config for saved preference
	configKey := "git-ai.conventionalCommits"
	value, err := git.GetConfig(configKey)
	if err == nil {
		// Config exists, use it
		return value == "true"
	}

	// Check repository history
	return git.UsesConventionalCommits()
}

// saveCommitFormatPreference saves the user's preference for commit format to git config
func saveCommitFormatPreference(useConventional bool) {
	value := "false"
	if useConventional {
		value = "true"
	}
	git.SaveConfigPreference("git-ai.conventionalCommits", value)
}

// saveCommitDescriptionPreference saves the user's preference for commit description format to git config
func saveCommitDescriptionPreference(useDescriptions bool) {
	value := "false"
	if useDescriptions {
		value = "true"
	}
	git.SaveConfigPreference("git-ai.commitsWithDescriptions", value)
}
