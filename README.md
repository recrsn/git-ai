# Git AI

Git AI is a Git extension that enhances your Git workflow with AI-powered features.

## Features

- `git ai commit`: Analyze staged changes and generate commit messages automatically
  - Uses LLMs to generate contextually relevant and well-formatted commit messages
  - Interactive approval process with option to edit the suggested message
  - `--auto` flag to automatically approve and commit without prompt
- `git ai config`: Configure your LLM settings
  - Set up API keys for your preferred LLM provider
  - Choose from different models (OpenAI, Anthropic, Ollama, etc.)
  - Customize API endpoints for self-hosted or alternative providers

## Installation

```bash
# Clone the repository
git clone https://github.com/recrsn/git-ai.git

# Build the binary
cd git-ai
go build

# Add to your PATH
# For example, you can create a symlink in a directory that's already in your PATH
ln -s "$(pwd)/git-ai" /usr/local/bin/git-ai

# Configure your LLM settings
git ai config
```

## Setup

Before using Git AI, you need to configure your LLM provider:

1. Run `git ai config` to set up your configuration
2. Select your LLM provider (OpenAI, Anthropic, Ollama, or Other)
3. Enter your API key
4. Select the model you want to use
5. Customize the API endpoint if needed

The configuration is stored in `~/.git-ai.yaml`. You can also use environment variables:
- `GIT_AI_API_KEY`: Your LLM provider API key
- `GIT_AI_MODEL`: The model to use (e.g., "gpt-4-turbo")
- `GIT_AI_API_URL`: The API endpoint URL

## Usage

```bash
# Stage your changes
git add .

# Generate an AI commit message
git ai commit

# Automatically approve and commit without prompting
git ai commit --auto

# Explicitly use conventional commit format (type(scope): description)
git ai commit --conventional

# Explicitly avoid using conventional commit format
git ai commit --no-conventional
```

Git AI automatically detects whether your repository uses conventional commit format by analyzing your commit history. If more than 50% of your recent commits follow the conventional format (`type(scope): description`), Git AI will default to generating messages in that style.

You can override this detection with the `--conventional` or `--no-conventional` flags. Your preference will be saved in the git config for future commits.

## How it works

Git AI analyzes your staged changes and commit history, then sends this data to your configured LLM to generate contextually relevant commit messages. The prompt includes:

1. The diff of staged changes
2. A list of changed files
3. Recent commit messages for context
4. Instructions to generate a commit message following conventions

The tool uses a terminal-based UI to provide an interactive experience, allowing you to approve, edit, or cancel the proposed commit message.

## Supported LLM Providers

- OpenAI (GPT-4, GPT-3.5)
- Anthropic (Claude)
- Ollama (for local LLM deployment)
- Custom providers via API endpoint configuration

## Customizing Prompts

The prompts used to generate commit messages are stored in text files and embedded into the binary at compile time. You can customize these prompts by editing the files in `pkg/llm/prompts/` before building:

- `commit_system.txt`: Contains the system instructions for the LLM, with conditional sections for conventional vs. standard format
- `commit_user.txt`: Contains the template for the user prompt with placeholders for diff, changed files, and recent commits

Both files use Go's template syntax for dynamic content:
- `{{if .UseConventional}}...{{else}}...{{end}}` blocks in the system prompt control format-specific instructions
- `{{.Diff}}`, `{{.ChangedFiles}}`, and `{{.RecentCommits}}` in the user prompt insert the relevant content

After modifying the prompts, rebuild the binary with `go build` to incorporate your changes.

## Development

This is an early-stage project. Future improvements will include:
- PR summarization
- Code review assistance 
- More intelligent analysis of code changes
- Support for additional Git operations like branch naming and PR descriptions