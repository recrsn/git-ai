package main

import (
	"github.com/recrsn/git-ai/cmd/branch"
	"github.com/recrsn/git-ai/cmd/commit"
	cmdConfig "github.com/recrsn/git-ai/cmd/config"
	"github.com/recrsn/git-ai/pkg/config"
	"github.com/recrsn/git-ai/pkg/logger"
	"github.com/spf13/cobra"
	"os"
)

var (
	configPath string
	verbose    int

	rootCmd = &cobra.Command{
		Use:   "git-ai",
		Short: "Git AI - LLM-powered Git extension",
		Long:  `Git AI enhances your Git workflow with AI-powered features like automatic commit message generation.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logger.SetLevel(verbose)

			logger.Debug("Git AI session started")

			// Set up logging level based on verbose flags
			if verbose > 0 {
				logger.Debug("Verbose logging enabled (level %d)", verbose)
			}

			// If config path is explicitly provided, set it
			if configPath != "" {
				logger.Debug("Using explicit config file: %s", configPath)
				config.ExplicitConfigPath = configPath
			}
		},
	}
)

func init() {
	// Add global flags
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to config file (default is $HOME/.git-ai.yaml and ./.git-ai.yaml)")
	rootCmd.PersistentFlags().CountVarP(&verbose, "verbose", "v", "Enable verbose output (-v for info, -vv for debug)")

	// Add subcommands
	rootCmd.AddCommand(branch.Cmd)
	rootCmd.AddCommand(commit.Cmd)
	rootCmd.AddCommand(cmdConfig.Cmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logger.Fatal("%v", err)
		os.Exit(1)
	}
}
