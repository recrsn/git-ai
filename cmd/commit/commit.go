package commit

import (
	"os"

	"github.com/recrsn/git-ai/pkg/config"
	"github.com/recrsn/git-ai/pkg/git"
	"github.com/recrsn/git-ai/pkg/llm"
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
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config: %v", err)
	}

	// Check if there are staged changes
	if !git.HasStagedChanges() {
		if amendCommit {
			logger.PrintMessage("No staged changes found. Will amend the previous commit message only.")
		} else {
			logger.Error("No staged changes found. Please stage your changes with 'git add' first.")
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

	// Get the staged changes diff
	diff := git.GetStagedDiff()
	if diff == "" {
		logger.Error("Could not retrieve diff of staged changes.")
		os.Exit(1)
	}

	// Get recent commit history
	recentCommits := git.GetRecentCommits()

	// Determine whether to use conventional commits format
	useConventionalCommits := shouldUseConventionalCommits()

	// Generate commit message based on staged changes and history - with spinner
	spinner, err := ui.ShowSpinner("Generating commit message with LLM...")
	if err != nil {
		logger.Error("Failed to start spinner: %v", err)
	}

	message := llm.GenerateCommitMessage(cfg, diff, recentCommits, useConventionalCommits, commitsWithDescriptions)

	if spinner != nil {
		spinner.Success("Commit message generated!")
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

	if amendCommit {
		logger.PrintMessage("Commit amended successfully!")
	} else {
		logger.PrintMessage("Commit created successfully!")
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
	configKey := "git-ai.conventionalCommits"
	value := "false"
	if useConventional {
		value = "true"
	}

	err := git.SetConfig(configKey, value)
	if err != nil {
		logger.Warn("Could not save commit format preference: %v", err)
	}
}

// saveCommitDescriptionPreference saves the user's preference for commit description format to git config
func saveCommitDescriptionPreference(useDescriptions bool) {
	configKey := "git-ai.commitsWithDescriptions"
	value := "false"
	if useDescriptions {
		value = "true"
	}

	err := git.SetConfig(configKey, value)
	if err != nil {
		logger.Warn("Could not save commit description preference: %v", err)
	}
}
