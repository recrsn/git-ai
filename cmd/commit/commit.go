package commit

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/recrsn/git-ai/pkg/git"
	"github.com/recrsn/git-ai/pkg/llm"
	"github.com/recrsn/git-ai/pkg/ui"
)

var autoApprove bool

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
}

func executeCommit() {
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

	// Generate commit message based on staged changes and history
	message := llm.GenerateCommitMessage(diff, recentCommits)

	// If auto-approve flag is not set, ask user to confirm or edit
	var proceed bool
	if !autoApprove {
		message, proceed = ui.PromptForConfirmation(message)
		if !proceed {
			os.Exit(0)
		}
	}

	// Create the commit with the message
	err := git.CreateCommit(message)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Commit created successfully!")
}
