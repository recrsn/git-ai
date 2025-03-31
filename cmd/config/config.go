package config

import (
	"fmt"
	"github.com/recrsn/git-ai/pkg/config"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
	existingConfig, _ := config.LoadConfig()

	// Prepare a new config with existing values as defaults
	config := existingConfig

	// Ask for provider
	providers := []string{"openai", "anthropic", "ollama", "other"}
	providerPrompt := &survey.Select{
		Message: "Select LLM provider:",
		Options: providers,
		Default: config.Provider,
	}
	survey.AskOne(providerPrompt, &config.Provider)

	// Ask for API key
	apiKeyPrompt := &survey.Password{
		Message: "Enter API key for " + config.Provider + ":",
	}
	survey.AskOne(apiKeyPrompt, &config.APIKey)

	// Ask for model
	var modelOptions []string
	switch config.Provider {
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
		Default: config.Model,
	}
	survey.AskOne(modelPrompt, &config.Model)

	// Ask for custom model if "other" is selected
	if config.Model == "custom" {
		customModelPrompt := &survey.Input{
			Message: "Enter custom model name:",
		}
		survey.AskOne(customModelPrompt, &config.Model)
	}

	// Ask for endpoint URL
	endpointPrompt := &survey.Input{
		Message: "Enter API endpoint URL:",
		Default: config.Endpoint,
	}
	survey.AskOne(endpointPrompt, &config.Endpoint)

	// Save the config
	if err := saveConfig(config); err != nil {
		fmt.Printf("Error saving configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Configuration saved successfully!")
	fmt.Println("You can now use 'git ai commit' to generate commit messages with your LLM.")
}

func saveConfig(config config.Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configFile := filepath.Join(homeDir, ".git-ai.yaml")
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
