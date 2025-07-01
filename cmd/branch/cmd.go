package branch

import (
	"github.com/recrsn/git-ai/pkg/git"
	"github.com/recrsn/git-ai/pkg/logger"
	"github.com/recrsn/git-ai/pkg/ui"
	"github.com/spf13/cobra"
	"os"
)

var (
	autoApprove bool
	description string
)

// Cmd represents the branch command
var Cmd = &cobra.Command{
	Use:   "branch [description]",
	Short: "Generate a meaningful Git branch name",
	Long:  `Analyzes your input and existing branches to generate a meaningful branch name.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// If description is provided as an argument, use it
		if len(args) > 0 {
			description = args[0]
		}

		// Check if there are any changes in the working directory
		diff := git.GetStagedDiff()
		if diff == "" {
			diff = git.GetUnstagedDiff()
		}

		// If no description is provided and no diff is available, prompt for one
		if description == "" && diff == "" {
			var err error
			description, err = ui.PromptForInput("Enter a brief description of your branch:", "")
			if err != nil {
				logger.Fatal("Error prompting for description: %v", err)
			}
			if description == "" {
				ui.PrintError("Description cannot be empty.")
				os.Exit(1)
			}
		}

		executeBranch(description, diff)
	},
}

func init() {
	Cmd.Flags().BoolVar(&autoApprove, "auto", false, "Automatically approve the generated branch name without prompting")
	Cmd.Flags().StringVarP(&description, "description", "d", "", "Brief description of the branch purpose")
}
