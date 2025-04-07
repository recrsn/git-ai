package config

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/recrsn/git-ai/pkg/config"
	"github.com/recrsn/git-ai/pkg/logger"
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

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			MarginBottom(1)

	inputLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("170"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
)

// ConfigModel is the main bubble tea model for configuration
type ConfigModel struct {
	config           config.Config
	existingConfig   config.Config
	step             int
	selectedProvider int
	selectedModel    int
	providerOptions  []string
	modelOptions     []string
	customProvider   textinput.Model
	customModel      textinput.Model
	endpoint         textinput.Model
	apiKey           textinput.Model
	error            string
	quitting         bool
}

func initialModel(existingConfig config.Config) ConfigModel {
	// Initialize textinputs
	customProvider := textinput.New()
	customProvider.Placeholder = "Enter custom provider name"
	customProvider.Focus()
	customProvider.CharLimit = 100
	customProvider.Width = 40

	customModel := textinput.New()
	customModel.Placeholder = "Enter custom model name"
	customModel.CharLimit = 100
	customModel.Width = 40

	endpoint := textinput.New()
	endpoint.Placeholder = "Enter API endpoint URL"
	endpoint.CharLimit = 200
	endpoint.Width = 60

	apiKey := textinput.New()
	apiKey.Placeholder = "Enter API key"
	apiKey.CharLimit = 200
	apiKey.Width = 60
	apiKey.EchoMode = textinput.EchoPassword
	apiKey.EchoCharacter = '•'

	// Provider options
	providerOptions := append(availableProviders, "custom")

	// Selected provider
	selectedProvider := 0
	for i, p := range providerOptions {
		if p == existingConfig.Provider {
			selectedProvider = i
			break
		}
	}
	if !contains(availableProviders, existingConfig.Provider) && existingConfig.Provider != "" {
		for i, p := range providerOptions {
			if p == "custom" {
				selectedProvider = i
				customProvider.SetValue(existingConfig.Provider)
				break
			}
		}
	}

	// Model options
	modelOptions := []string{}
	if info, ok := providers[existingConfig.Provider]; ok {
		modelOptions = info.availableModels
	} else {
		modelOptions = []string{"custom"}
	}

	// Selected model
	selectedModel := 0
	for i, m := range modelOptions {
		if m == existingConfig.Model {
			selectedModel = i
			break
		}
	}
	if !contains(modelOptions, existingConfig.Model) && existingConfig.Model != "" && existingConfig.Model != "custom" {
		for i, m := range modelOptions {
			if m == "custom" {
				selectedModel = i
				customModel.SetValue(existingConfig.Model)
				break
			}
		}
	}

	// Set endpoint default
	endpoint.SetValue(existingConfig.Endpoint)

	return ConfigModel{
		config:           existingConfig,
		existingConfig:   existingConfig,
		step:             0,
		selectedProvider: selectedProvider,
		selectedModel:    selectedModel,
		providerOptions:  providerOptions,
		modelOptions:     modelOptions,
		customProvider:   customProvider,
		customModel:      customModel,
		endpoint:         endpoint,
		apiKey:           apiKey,
		error:            "",
	}
}

func (m ConfigModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyEsc:
			// Go back to previous step or quit
			if m.step > 0 {
				m.step--
				m.error = ""
			} else {
				m.quitting = true
				return m, tea.Quit
			}
		case tea.KeyEnter:
			// Go to next step
			switch m.step {
			case 0: // Provider selection
				selectedProvider := m.providerOptions[m.selectedProvider]
				if selectedProvider == "custom" {
					// Go to custom provider input
					m.step = 1
					m.customProvider.Focus()
				} else {
					// Set provider and move to model selection
					m.config.Provider = selectedProvider
					// If provider changed, update endpoint and available models
					if m.config.Provider != m.existingConfig.Provider {
						if providerConfig, ok := providers[m.config.Provider]; ok {
							m.config.Endpoint = providerConfig.endpoint
							m.endpoint.SetValue(m.config.Endpoint)
							m.modelOptions = providerConfig.availableModels
							m.selectedModel = 0
							for i, model := range m.modelOptions {
								if model == providerConfig.defaultModel {
									m.selectedModel = i
									break
								}
							}
						}
					}
					m.step = 3 // Skip custom provider input
				}
			case 1: // Custom provider input
				if m.customProvider.Value() == "" {
					m.error = "Provider name cannot be empty"
					return m, nil
				}
				m.config.Provider = m.customProvider.Value()
				m.error = ""
				m.step = 2 // Go to endpoint
				m.endpoint.Focus()
			case 2: // Endpoint input
				if m.endpoint.Value() == "" {
					m.error = "Endpoint URL cannot be empty"
					return m, nil
				}
				m.config.Endpoint = m.endpoint.Value()
				m.error = ""
				m.modelOptions = []string{"custom"}
				m.selectedModel = 0
				m.step = 3
			case 3: // Model selection
				selectedModel := m.modelOptions[m.selectedModel]
				if selectedModel == "custom" {
					// Go to custom model input
					m.step = 4
					m.customModel.Focus()
				} else {
					m.config.Model = selectedModel
					m.step = 5 // Skip custom model input, go to API key
					m.apiKey.Focus()
				}
			case 4: // Custom model input
				if m.customModel.Value() == "" {
					m.error = "Model name cannot be empty"
					return m, nil
				}
				m.config.Model = m.customModel.Value()
				m.error = ""
				m.step = 5 // Go to API key
				m.apiKey.Focus()
			case 5: // API key input
				newAPIKey := m.apiKey.Value()
				if newAPIKey == "" && m.existingConfig.APIKey != "" {
					// Keep existing key
					m.config.APIKey = m.existingConfig.APIKey
				} else if newAPIKey == "" {
					m.error = "API key cannot be empty"
					return m, nil
				} else {
					m.config.APIKey = newAPIKey
				}
				// Save config and quit
				if err := config.SaveConfig(m.config); err != nil {
					m.error = fmt.Sprintf("Error saving config: %v", err)
					return m, nil
				}
				// Success - quit the application
				return m, tea.Quit
			}
		case tea.KeyUp, tea.KeyShiftTab:
			if m.step == 0 && m.selectedProvider > 0 {
				m.selectedProvider--
			} else if m.step == 3 && m.selectedModel > 0 {
				m.selectedModel--
			}
		case tea.KeyDown, tea.KeyTab:
			if m.step == 0 && m.selectedProvider < len(m.providerOptions)-1 {
				m.selectedProvider++
			} else if m.step == 3 && m.selectedModel < len(m.modelOptions)-1 {
				m.selectedModel++
			}
		}
	}

	// Handle input fields
	switch m.step {
	case 1:
		m.customProvider, cmd = m.customProvider.Update(msg)
		return m, cmd
	case 2:
		m.endpoint, cmd = m.endpoint.Update(msg)
		return m, cmd
	case 4:
		m.customModel, cmd = m.customModel.Update(msg)
		return m, cmd
	case 5:
		m.apiKey, cmd = m.apiKey.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m ConfigModel) View() string {
	var s string

	if m.quitting {
		return ""
	}

	s = titleStyle.Render("Configure git-ai") + "\n\n"

	// Show error if any
	if m.error != "" {
		s += errorStyle.Render(m.error) + "\n\n"
	}

	// Show current step
	switch m.step {
	case 0: // Provider selection
		s += "Select LLM provider:\n\n"
		for i, provider := range m.providerOptions {
			cursor := " "
			style := lipgloss.NewStyle()
			if i == m.selectedProvider {
				cursor = ">"
				style = selectedItemStyle
			}

			displayName := provider
			if provider == "custom" {
				if !contains(availableProviders, m.existingConfig.Provider) && m.existingConfig.Provider != "" {
					displayName = fmt.Sprintf("custom (%s)", m.existingConfig.Provider)
				} else {
					displayName = "custom"
				}
			} else if info, ok := providers[provider]; ok {
				displayName = info.name
			}

			s += fmt.Sprintf("%s %s\n", cursor, style.Render(displayName))
		}
	case 1: // Custom provider input
		s += "Enter custom provider name:\n\n"
		s += m.customProvider.View() + "\n"
	case 2: // Endpoint input
		s += fmt.Sprintf("Enter API endpoint URL for %s:\n\n", m.config.Provider)
		s += m.endpoint.View() + "\n"
	case 3: // Model selection
		s += fmt.Sprintf("Select model for %s:\n\n", m.config.Provider)
		for i, model := range m.modelOptions {
			cursor := " "
			style := lipgloss.NewStyle()
			if i == m.selectedModel {
				cursor = ">"
				style = selectedItemStyle
			}

			displayName := model
			if model == "custom" {
				if !contains(m.modelOptions, m.existingConfig.Model) && m.existingConfig.Model != "" && m.existingConfig.Model != "custom" {
					displayName = fmt.Sprintf("custom (%s)", m.existingConfig.Model)
				} else {
					displayName = "custom"
				}
			}

			s += fmt.Sprintf("%s %s\n", cursor, style.Render(displayName))
		}
	case 4: // Custom model input
		s += "Enter custom model name:\n\n"
		s += m.customModel.View() + "\n"
	case 5: // API key input
		message := fmt.Sprintf("Enter API key for %s:", m.config.Provider)
		if m.existingConfig.APIKey != "" {
			message += " (leave blank to keep existing key)"
		}
		s += message + "\n\n"
		s += m.apiKey.View() + "\n"
	}

	s += "\n"
	if m.step > 0 {
		s += "Press ESC to go back, Enter to continue\n"
	} else {
		s += "Press ESC to cancel, Enter to continue\n"
	}

	return s
}

func executeConfig() {
	// Load existing config if available
	existingConfig, err := config.LoadConfig()
	if err != nil {
		logger.Warn("Could not load existing configuration: %v. Using defaults.", err)
		existingConfig = config.DefaultConfig()
	}

	// Initialize model with existing config
	m := initialModel(existingConfig)

	// Run the Bubble Tea program
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		logger.Fatal("Error running Bubble Tea program: %v", err)
		os.Exit(1)
	}

	// Check if configuration was successful
	finalM, ok := finalModel.(ConfigModel)
	if !ok {
		logger.Fatal("Could not convert final model")
		os.Exit(1)
	}

	// If we got here and didn't quit early, configuration was successful
	if !finalM.quitting {
		logger.PrintMessage("Configuration saved successfully!")
		logger.PrintMessage("You can now use 'git ai commit' to generate commit messages with your LLM.")
	} else {
		logger.PrintMessage("Configuration cancelled.")
	}
}
