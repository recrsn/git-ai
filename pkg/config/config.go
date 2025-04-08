package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/recrsn/git-ai/pkg/logger"
	"github.com/spf13/viper"
)

var (
	// ExplicitConfigPath is set when a config file is explicitly provided by commandline
	ExplicitConfigPath string
)

// Config represents the Git AI configuration
type Config struct {
	Provider string `mapstructure:"provider"`
	APIKey   string `mapstructure:"api_key"`
	Model    string `mapstructure:"model"`
	Endpoint string `mapstructure:"endpoint"`
	Editor   string `mapstructure:"editor"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Provider: "openai",
		APIKey:   "",
		Model:    "gpt-4-turbo",
		Endpoint: "https://api.openai.com/v1",
		Editor:   "",
	}
}

// initViper initializes viper with the configuration sources
func initViper() (*viper.Viper, error) {
	v := viper.New()

	// If explicit config file is provided, only use that
	if ExplicitConfigPath != "" {
		v.SetConfigFile(ExplicitConfigPath)

		// Set defaults
		defaults := DefaultConfig()
		v.SetDefault("provider", defaults.Provider)
		v.SetDefault("api_key", defaults.APIKey)
		v.SetDefault("model", defaults.Model)
		v.SetDefault("endpoint", defaults.Endpoint)

		// Read configuration
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", ExplicitConfigPath, err)
		}

		logger.Debug("Using configuration from %s", v.ConfigFileUsed())

		return v, nil
	}

	// Otherwise, merge config from multiple sources
	v.SetConfigName(".git-ai")
	v.SetConfigType("yaml")

	// Add search paths in order of priority (lowest to highest)
	// 1. User home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	v.AddConfigPath(homeDir)

	// 2. Current working directory
	cwd, err := os.Getwd()
	if err != nil {
		logger.Warn("Failed to get current working directory: %v", err)
	} else {
		v.AddConfigPath(cwd)
	}

	// Set defaults
	defaults := DefaultConfig()
	v.SetDefault("provider", defaults.Provider)
	v.SetDefault("api_key", defaults.APIKey)
	v.SetDefault("model", defaults.Model)
	v.SetDefault("endpoint", defaults.Endpoint)
	v.SetDefault("editor", defaults.Editor)

	// Environment variables
	v.SetEnvPrefix("GIT_AI")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Map for environment variables
	v.BindEnv("api_key", "GIT_AI_API_KEY")
	v.BindEnv("model", "GIT_AI_MODEL")
	v.BindEnv("endpoint", "GIT_AI_API_URL")
	v.BindEnv("editor", "GIT_AI_EDITOR")

	// Read configuration
	err = v.ReadInConfig()
	if err != nil {
		// Config file not found is not an error
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		logger.Debug("Config file not found, using defaults and environment variables")
	} else {
		logger.Debug("Using configuration from %s", v.ConfigFileUsed())

		// If we read a config from home, check if there's also one in cwd
		if cwd != "" && v.ConfigFileUsed() == filepath.Join(homeDir, ".git-ai.yaml") {
			// Create a new viper instance for the cwd config
			cwdViper := viper.New()
			cwdViper.SetConfigName(".git-ai")
			cwdViper.SetConfigType("yaml")
			cwdViper.AddConfigPath(cwd)

			// Try to read cwd config
			if err := cwdViper.ReadInConfig(); err == nil {
				logger.Debug("Merging configuration from %s", cwdViper.ConfigFileUsed())

				// Merge configs, with cwd taking precedence
				for _, key := range cwdViper.AllKeys() {
					v.Set(key, cwdViper.Get(key))
				}
			}
		}
	}

	return v, nil
}

// LoadConfig loads the configuration from all sources
func LoadConfig() (Config, error) {
	v, err := initViper()
	if err != nil {
		return DefaultConfig(), fmt.Errorf("failed to initialize configuration: %w", err)
	}

	var config Config
	err = v.Unmarshal(&config)
	if err != nil {
		return DefaultConfig(), fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

// SaveConfig saves the configuration to the user's home directory
func SaveConfig(config Config) error {
	// If explicit config path was provided, save to that location
	configPath := ExplicitConfigPath

	// Otherwise default to home directory
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, ".git-ai.yaml")
	}

	v := viper.New()
	v.SetConfigFile(configPath)

	v.Set("provider", config.Provider)
	v.Set("api_key", config.APIKey)
	v.Set("model", config.Model)
	v.Set("endpoint", config.Endpoint)
	v.Set("editor", config.Editor)

	if err := v.WriteConfig(); err != nil {
		// Check if the file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Create directory if it doesn't exist
			dir := filepath.Dir(configPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}
			return v.SafeWriteConfig()
		}
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
