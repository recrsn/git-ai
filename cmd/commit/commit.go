package commit

import (
	"fmt"
	"os"

	"github.com/recrsn/git-ai/pkg/config"
	"github.com/recrsn/git-ai/pkg/git"
	"github.com/recrsn/git-ai/pkg/llm"
	"github.com/recrsn/git-ai/pkg/ui"
	"github.com/spf13/cobra"
)

var (
	autoApprove           bool
	conventionalCommits   bool
	noConventionalCommits bool
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
}

func executeCommit() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Check if there are staged changes
	if !git.HasStagedChanges() {
		fmt.Println("No staged changes found. Please stage your changes with 'git add' first.")
		os.Exit(1)
	}

	// Get the staged changes diff
	diff := git.GetStagedDiff()
	if diff == "" {
		fmt.Println("Could not retrieve diff of staged changes.")
		os.Exit(1)
	}

	// Get recent commit history
	recentCommits := git.GetRecentCommits()

	// Determine whether to use conventional commits format
	useConventionalCommits := shouldUseConventionalCommits()

	// Generate commit message based on staged changes and history
	fmt.Println("Generating commit message with LLM...")
	message := llm.GenerateCommitMessage(cfg, diff, recentCommits, useConventionalCommits)

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

	// Create the commit with the message
	err = git.CreateCommit(message)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Commit created successfully!")
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
		fmt.Printf("Warning: Could not save commit format preference: %v\n", err)
	}
}
