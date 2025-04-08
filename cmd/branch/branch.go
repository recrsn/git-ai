package branch

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

		// If no description is provided, prompt for one
		if description == "" {
			var err error
			description, err = ui.PromptForInput("Enter a brief description of your branch:", "")
			if err != nil {
				logger.Fatal("Error prompting for description: %v", err)
			}
			if description == "" {
				logger.Error("Description cannot be empty.")
				os.Exit(1)
			}
		}

		executeBranch(description)
	},
}

func init() {
	Cmd.Flags().BoolVar(&autoApprove, "auto", false, "Automatically approve the generated branch name without prompting")
	Cmd.Flags().StringVarP(&description, "description", "d", "", "Brief description of the branch purpose")
}

func executeBranch(description string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config: %v", err)
	}

	// Generate branch name - with spinner
	spinner, err := ui.ShowSpinner("Generating branch name with LLM...")
	if err != nil {
		logger.Error("Failed to start spinner: %v", err)
	}

	branchName, err := llm.GenerateBranchName(cfg, description)
	if err != nil {
		if spinner != nil {
			spinner.Fail("Failed to generate branch name!")
		}
		logger.Fatal("Failed to generate branch name: %v", err)
	}

	if spinner != nil {
		spinner.Success("Branch name generated!")
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
			logger.PrintMessage(branchName)
			os.Exit(0)
		case "Cancel":
			logger.PrintMessage("Branch creation cancelled.")
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

		logger.PrintMessagef("Branch '%s' created successfully!", branchName)
	}
}
