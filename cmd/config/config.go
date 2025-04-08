package config

import (
	"fmt"
	"os"

	"github.com/recrsn/git-ai/pkg/config"
	"github.com/recrsn/git-ai/pkg/git"
	"github.com/recrsn/git-ai/pkg/logger"
	"github.com/recrsn/git-ai/pkg/ui"
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

	configResult := existingConfig

	// Set up header
	ui.DisplayHeader("Configure git-ai")

	// Step 1: Select provider
	ui.DisplaySection("Provider Selection")

	providerOptions := append(availableProviders, "custom")
	providerDisplayOptions := make([]string, len(providerOptions))

	selectedProviderIndex := 0
	for i, p := range providerOptions {
		displayName := p
		if p == "custom" {
			if !contains(availableProviders, existingConfig.Provider) && existingConfig.Provider != "" {
				displayName = fmt.Sprintf("custom (%s)", existingConfig.Provider)
			}
		} else if info, ok := providers[p]; ok {
			displayName = info.name
		}
		providerDisplayOptions[i] = displayName

		if p == existingConfig.Provider {
			selectedProviderIndex = i
		}
	}

	if !contains(availableProviders, existingConfig.Provider) && existingConfig.Provider != "" {
		// For custom provider, select the "custom" option
		for i, p := range providerOptions {
			if p == "custom" {
				selectedProviderIndex = i
				break
			}
		}
	}

	defaultOption := ""
	if len(providerDisplayOptions) > selectedProviderIndex {
		defaultOption = providerDisplayOptions[selectedProviderIndex]
	} else if len(providerDisplayOptions) > 0 {
		defaultOption = providerDisplayOptions[0]
	}

	selectedProvider, err := ui.PromptForSelection(
		providerDisplayOptions,
		defaultOption,
		"Select LLM provider:",
	)

	if err != nil {
		ui.ExitWithError(fmt.Sprintf("Error selecting provider: %v", err))
	}

	// Determine the actual provider value
	for i, displayName := range providerDisplayOptions {
		if displayName == selectedProvider {
			configResult.Provider = providerOptions[i]
			break
		}
	}

	// Step 2: Handle custom provider if selected
	if configResult.Provider == "custom" {
		ui.DisplaySection("Custom Provider")

		customProviderName := ""
		if !contains(availableProviders, existingConfig.Provider) && existingConfig.Provider != "" {
			customProviderName = existingConfig.Provider
		}

		customProvider, err := ui.PromptForInput("Enter custom provider name:", customProviderName)
		if err != nil {
			ui.ExitWithError(fmt.Sprintf("Error getting custom provider: %v", err))
		}

		if customProvider == "" {
			ui.ExitWithError("Provider name cannot be empty")
		}

		configResult.Provider = customProvider

		// Step 2a: Get endpoint for custom provider
		defaultEndpoint := existingConfig.Endpoint
		endpoint, err := ui.PromptForInput(
			fmt.Sprintf("Enter API endpoint URL for %s:", configResult.Provider),
			defaultEndpoint,
		)

		if err != nil {
			ui.ExitWithError(fmt.Sprintf("Error getting endpoint: %v", err))
		}

		if endpoint == "" {
			ui.ExitWithError("Endpoint URL cannot be empty")
		}

		configResult.Endpoint = endpoint
	} else {
		// Set endpoint for known provider
		if providerConfig, ok := providers[configResult.Provider]; ok {
			configResult.Endpoint = providerConfig.endpoint
		}
	}

	// Step 3: Model selection
	ui.DisplaySection("Model Selection")

	modelOptions := []string{}
	if info, ok := providers[configResult.Provider]; ok {
		modelOptions = info.availableModels
	} else {
		modelOptions = []string{"custom"}
	}

	modelDisplayOptions := make([]string, len(modelOptions))
	selectedModelIndex := 0

	for i, m := range modelOptions {
		displayName := m
		if m == "custom" {
			if !contains(modelOptions, existingConfig.Model) && existingConfig.Model != "" && existingConfig.Model != "custom" {
				displayName = fmt.Sprintf("custom (%s)", existingConfig.Model)
			}
		}
		modelDisplayOptions[i] = displayName

		if m == existingConfig.Model {
			selectedModelIndex = i
		}
	}

	if !contains(modelOptions, existingConfig.Model) && existingConfig.Model != "" && existingConfig.Model != "custom" {
		// For custom model, select the "custom" option
		for i, m := range modelOptions {
			if m == "custom" {
				selectedModelIndex = i
				break
			}
		}
	}

	defaultModelOption := ""
	if len(modelDisplayOptions) > selectedModelIndex {
		defaultModelOption = modelDisplayOptions[selectedModelIndex]
	} else if len(modelDisplayOptions) > 0 {
		defaultModelOption = modelDisplayOptions[0]
	}

	selectedModel, err := ui.PromptForSelection(
		modelDisplayOptions,
		defaultModelOption,
		fmt.Sprintf("Select model for %s:", configResult.Provider),
	)

	if err != nil {
		ui.ExitWithError(fmt.Sprintf("Error selecting model: %v", err))
	}

	// Determine the actual model value
	for i, displayName := range modelDisplayOptions {
		if displayName == selectedModel {
			configResult.Model = modelOptions[i]
			break
		}
	}

	// Step 4: Handle custom model if selected
	if configResult.Model == "custom" {
		ui.DisplaySection("Custom Model")

		customModelName := ""
		if !contains(modelOptions, existingConfig.Model) && existingConfig.Model != "" && existingConfig.Model != "custom" {
			customModelName = existingConfig.Model
		}

		customModel, err := ui.PromptForInput("Enter custom model name:", customModelName)
		if err != nil {
			ui.ExitWithError(fmt.Sprintf("Error getting custom model: %v", err))
		}

		if customModel == "" {
			ui.ExitWithError("Model name cannot be empty")
		}

		configResult.Model = customModel
	}

	// Step 5: API Key
	ui.DisplaySection("API Key")

	apiKeyPrompt := fmt.Sprintf("Enter API key for %s:", configResult.Provider)
	if existingConfig.APIKey != "" {
		apiKeyPrompt += " (leave blank to keep existing key)"
	}

	apiKey, err := ui.PromptForPassword(apiKeyPrompt)
	if err != nil {
		ui.ExitWithError(fmt.Sprintf("Error getting API key: %v", err))
	}

	if apiKey == "" && existingConfig.APIKey != "" {
		// Keep existing key
		configResult.APIKey = existingConfig.APIKey
	} else if apiKey == "" {
		ui.ExitWithError("API key cannot be empty")
	} else {
		configResult.APIKey = apiKey
	}

	// Step 6: Editor preference
	ui.DisplaySection("Editor Preference")

	currentEditor := git.GetPreferredEditor()
	editorPrompt := fmt.Sprintf("Configure custom editor for git-ai (current: %s):", currentEditor)
	editorValue, err := ui.PromptForInput(editorPrompt, existingConfig.Editor)
	if err != nil {
		ui.ExitWithError(fmt.Sprintf("Error getting editor preference: %v", err))
	}

	// Only set if the user entered something
	if editorValue != "" {
		configResult.Editor = editorValue
	}

	// Save configuration
	spinner, _ := ui.ShowSpinner("Saving configuration...")
	if err := config.SaveConfig(configResult); err != nil {
		spinner.Fail(fmt.Sprintf("Error saving config: %v", err))
		os.Exit(1)
	}
	spinner.Success("Configuration saved successfully!")

	ui.DisplayInfo("You can now use 'git ai commit' to generate commit messages with your LLM.")
}
