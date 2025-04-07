package config

import (
	"errors"
	"github.com/recrsn/git-ai/pkg/config"
	"github.com/recrsn/git-ai/pkg/logger"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

// contains checks if a string is present in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

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
	providerOptions := append(availableProviders, "custom")

	// If the existing provider is "custom", add it as a default
	providerDefault := cfg.Provider
	customProviderValue := ""
	if !contains(availableProviders, cfg.Provider) && cfg.Provider != "" {
		providerDefault = "custom"
		customProviderValue = cfg.Provider
	}

	providerPrompt := &survey.Select{
		Message: "Select LLM provider:",
		Options: providerOptions,
		Default: providerDefault,
		Description: func(value string, index int) string {
			if value == "custom" {
				if customProviderValue != "" {
					return "Custom provider (" + customProviderValue + ")"
				}
				return "Custom provider"
			}
			if info, ok := providers[value]; ok {
				return info.name
			}
			return ""
		},
	}
	err = survey.AskOne(providerPrompt, &cfg.Provider)
	if err != nil {
		logger.Fatal("Error reading from input: %v", err)
	}

	// Ask for custom provider name if "custom" is selected
	if cfg.Provider == "custom" {
		customDefault := ""
		if !contains(availableProviders, existingConfig.Provider) && existingConfig.Provider != "" {
			customDefault = existingConfig.Provider
		}

		customProviderPrompt := &survey.Input{
			Message: "Enter custom provider name:",
			Default: customDefault,
		}
		err := survey.AskOne(customProviderPrompt, &cfg.Provider, survey.WithValidator(func(ans interface{}) error {
			answer := ans.(string)
			if answer == "" {
				return errors.New("Provider name cannot be empty")
			}
			return nil
		}))
		if err != nil {
			logger.Fatal("Error reading from input: %v", err)
		}
	}

	// Set default values based on provider, if it was changed
	if cfg.Provider != existingConfig.Provider {
		if providerConfig, ok := providers[cfg.Provider]; ok {
			cfg.Model = providerConfig.defaultModel
			cfg.Endpoint = providerConfig.endpoint
		}
	}

	// Ask for endpoint URL for custom providers or if the provider isn't recognized
	if _, ok := providers[cfg.Provider]; !ok {
		endpointPrompt := &survey.Input{
			Message: "Enter API endpoint URL for " + cfg.Provider + ":",
			Default: cfg.Endpoint,
		}
		err = survey.AskOne(endpointPrompt, &cfg.Endpoint, survey.WithValidator(func(ans interface{}) error {
			answer := ans.(string)
			if answer == "" {
				return errors.New("Endpoint URL cannot be empty")
			}
			return nil
		}))
		if err != nil {
			logger.Fatal("Error reading from input: %v", err)
		}
	}

	// Ask for API key
	apiKeyMessage := "Enter API key for " + cfg.Provider + ":"
	if existingConfig.APIKey != "" {
		apiKeyMessage += " (leave blank to keep existing key)"
	}

	apiKeyPrompt := &survey.Password{
		Message: apiKeyMessage,
	}

	var newAPIKey string
	err = survey.AskOne(apiKeyPrompt, &newAPIKey, survey.WithValidator(func(ans interface{}) error {
		answer := ans.(string)
		// If there's an existing API key, allow empty input to keep it
		if answer == "" && existingConfig.APIKey != "" {
			return nil
		}
		// Otherwise require a value
		if answer == "" {
			return errors.New("API key cannot be empty")
		}
		return nil
	}))

	if err != nil {
		logger.Fatal("Error reading from input: %v", err)
	}

	// Only update if a new value was provided
	if newAPIKey != "" {
		cfg.APIKey = newAPIKey
	} else if existingConfig.APIKey != "" {
		// Keep existing key if input was blank and we have an existing key
		cfg.APIKey = existingConfig.APIKey
	}

	// Get available models or use custom options
	var modelOptions []string
	var modelDefault string
	customModelValue := ""

	if providerInfo, ok := providers[cfg.Provider]; ok {
		modelOptions = providerInfo.availableModels
	} else {
		// For custom providers, offer only "custom" option
		modelOptions = []string{"custom"}
	}

	// Check if current model exists in options or should be custom
	modelDefault = cfg.Model
	if !contains(modelOptions, cfg.Model) && cfg.Model != "" && cfg.Model != "custom" {
		// If the current model isn't in the list and isn't empty, set default to custom
		modelDefault = "custom"
		customModelValue = cfg.Model
	}

	modelPrompt := &survey.Select{
		Message: "Select model:",
		Options: modelOptions,
		Default: modelDefault,
		Description: func(value string, index int) string {
			if value == "custom" && customModelValue != "" {
				return "Custom model (" + customModelValue + ")"
			}
			return ""
		},
	}

	var selectedModel string
	err = survey.AskOne(modelPrompt, &selectedModel)
	if err != nil {
		logger.Fatal("Error reading from input: %v", err)
	}

	// Ask for custom model if "custom" is selected
	if selectedModel == "custom" {
		customDefault := ""
		// If current model is custom, use it as default
		if !contains(modelOptions, cfg.Model) && cfg.Model != "" && cfg.Model != "custom" {
			customDefault = cfg.Model
		}

		customModelPrompt := &survey.Input{
			Message: "Enter custom model name:",
			Default: customDefault,
		}
		err := survey.AskOne(customModelPrompt, &cfg.Model, survey.WithValidator(func(ans interface{}) error {
			answer := ans.(string)
			if answer == "" {
				return errors.New("Model name cannot be empty")
			}
			return nil
		}))
		if err != nil {
			logger.Fatal("Error reading from input: %v", err)
		}
	} else {
		cfg.Model = selectedModel
	}

	// We don't ask for endpoint for known providers as they have predefined endpoints

	// Save the config
	if err := config.SaveConfig(cfg); err != nil {
		logger.Fatal("Error saving configuration: %v", err)
		os.Exit(1)
	}

	logger.PrintMessage("Configuration saved successfully!")
	logger.PrintMessage("You can now use 'git ai commit' to generate commit messages with your LLM.")
}
