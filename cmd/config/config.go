package config

import (
	"fmt"
	"os"

	"github.com/recrsn/git-ai/pkg/config"
	"github.com/recrsn/git-ai/pkg/logger"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

// Cmd represents the config command
var Cmd = &cobra.Command{
	Use:   "config",
	Short: "Configure git-ai settings",
	Long:  `Set up or update your git-ai configuration, including LLM API keys and settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		executeConfig()
	},
}

func executeConfig() {
	// Load existing config if available
	existingConfig, err := config.LoadConfig()
	if err != nil {
		logger.Warn("Could not load existing configuration: %v. Using defaults.", err)
		existingConfig = config.DefaultConfig()
	}

	// Prepare a new config with existing values as defaults
	cfg := existingConfig

	// Ask for provider
	providers := []string{"openai", "anthropic", "ollama", "other"}
	providerPrompt := &survey.Select{
		Message: "Select LLM provider:",
		Options: providers,
		Default: cfg.Provider,
	}
	survey.AskOne(providerPrompt, &cfg.Provider)

	// Ask for API key
	apiKeyPrompt := &survey.Password{
		Message: "Enter API key for " + cfg.Provider + ":",
	}
	survey.AskOne(apiKeyPrompt, &cfg.APIKey)

	// Ask for model
	var modelOptions []string
	switch cfg.Provider {
	case "openai":
		modelOptions = []string{"gpt-4-turbo", "gpt-4", "gpt-3.5-turbo"}
	case "anthropic":
		modelOptions = []string{"claude-3-5-sonnet", "claude-3-opus", "claude-3-haiku"}
	case "ollama":
		modelOptions = []string{"llama3", "mistral", "codellama"}
	default:
		modelOptions = []string{"custom"}
	}

	modelPrompt := &survey.Select{
		Message: "Select model:",
		Options: modelOptions,
		Default: cfg.Model,
	}
	survey.AskOne(modelPrompt, &cfg.Model)

	// Ask for custom model if "other" is selected
	if cfg.Model == "custom" {
		customModelPrompt := &survey.Input{
			Message: "Enter custom model name:",
		}
		survey.AskOne(customModelPrompt, &cfg.Model)
	}

	// Ask for endpoint URL
	endpointPrompt := &survey.Input{
		Message: "Enter API endpoint URL:",
		Default: cfg.Endpoint,
	}
	survey.AskOne(endpointPrompt, &cfg.Endpoint)

	// Save the config
	if err := config.SaveConfig(cfg); err != nil {
		logger.Fatal("Error saving configuration: %v", err)
		os.Exit(1)
	}

	fmt.Println("Configuration saved successfully!")
	fmt.Println("You can now use 'git ai commit' to generate commit messages with your LLM.")
}