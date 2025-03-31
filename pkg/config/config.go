package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the Git AI configuration
type Config struct {
	Provider string `yaml:"provider"`
	APIKey   string `yaml:"api_key"`
	Model    string `yaml:"model"`
	Endpoint string `yaml:"endpoint"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Provider: "openai",
		APIKey:   "",
		Model:    "gpt-4-turbo",
		Endpoint: "https://api.openai.com/v1",
	}
}

// LoadConfig loads the configuration from the user's home directory
func LoadConfig() (Config, error) {
	config := DefaultConfig()

	// Override from environment variables
	if apiKey := os.Getenv("GIT_AI_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
	}

	if model := os.Getenv("GIT_AI_MODEL"); model != "" {
		config.Model = model
	}

	if endpoint := os.Getenv("GIT_AI_API_URL"); endpoint != "" {
		config.Endpoint = endpoint
	}

	// Try to load config file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return config, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configFile := filepath.Join(homeDir, ".git-ai.yaml")
	file, err := os.Open(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, use defaults
			return config, nil
		}
		return config, fmt.Errorf("failed to open config file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("Warning: failed to close config file: %v\n", err)
		}
	}(file)

	content, err := io.ReadAll(file)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(content, &config); err != nil {
		return config, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}
