package main

import (
	"fmt"
	"os"

	"github.com/recrsn/git-ai/cmd/commit"
	"github.com/recrsn/git-ai/cmd/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "git-ai",
	Short: "Git AI - LLM-powered Git extension",
	Long:  `Git AI enhances your Git workflow with AI-powered features like automatic commit message generation.`,
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(commit.Cmd)
	rootCmd.AddCommand(config.Cmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
