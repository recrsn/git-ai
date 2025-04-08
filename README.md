# Git AI

Git AI enhances your Git workflow with AI-powered features.

## Features

- `git ai commit`: Generates commit messages from staged changes
  - Create relevant, well-formatted commit messages
  - Provide interactive approval with edit option
  - Commit automatically with `--auto` flag
  - Add detailed descriptions with `--with-descriptions`
  - Control format with `--conventional` and `--no-conventional` flags
- `git ai config`: Manages LLM settings
  - Set up API keys for your preferred provider
  - Offer various models (OpenAI, Anthropic, Ollama, etc.)
  - Support custom endpoints for self-hosted options
  - Enable custom providers through API configuration

## Installation

### Quick Install (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/recrsn/git-ai/main/install.sh | bash
```

This will download and install the latest release to `~/.local/bin/git-ai`.

### Manual Installation

#### Download Binary

```bash
# Create ~/.local/bin if it doesn't exist
mkdir -p ~/.local/bin

# Download the latest release for your platform
# Replace OS with darwin/linux and ARCH with amd64/arm64 as needed
curl -L -o ~/.local/bin/git-ai https://github.com/recrsn/git-ai/releases/latest/download/git-ai-OS-ARCH

# Make executable
chmod +x ~/.local/bin/git-ai

# Ensure ~/.local/bin is in your PATH
# Add to your shell profile file (.bashrc, .zshrc, etc.) if needed:
# export PATH="$PATH:$HOME/.local/bin"

# Configure your LLM settings
git ai config
```

#### Install from Source

```bash
# Using go install
go install github.com/recrsn/git-ai@latest

# OR clone and build
git clone https://github.com/recrsn/git-ai.git
cd git-ai
go build
mkdir -p ~/.local/bin
cp git-ai ~/.local/bin/

# Ensure ~/.local/bin is in your PATH
# Add to your shell profile file (.bashrc, .zshrc, etc.) if needed:
# export PATH="$PATH:$HOME/.local/bin"

# Configure your LLM settings
git ai config
```

## Setup

Before using Git AI, configure your LLM provider:

1. Run `git ai config`
2. Select your LLM provider (OpenAI, Anthropic, Ollama, or Other)
3. Enter your API key
4. Select your preferred model
5. Customize the API endpoint if needed

### Configuration Files

Git AI checks configuration in these locations (highest to lowest precedence):

1. Command-line config: `git ai --config /path/to/config.yaml`
2. Project config: `./.git-ai.yaml` in current directory
3. User config: `~/.git-ai.yaml` in home directory
4. Environment variables:
   - `GIT_AI_API_KEY`: LLM provider API key
   - `GIT_AI_MODEL`: Model name (e.g., "gpt-4-turbo")
   - `GIT_AI_API_URL`: API endpoint URL
5. Git config variables:
   - `git-ai.conventionalCommits`: Use conventional format (true/false)
   - `git-ai.commitsWithDescriptions`: Include detailed descriptions (true/false)
6. Default values

This provides flexible configuration at global and project-specific levels.

## Usage

```bash
# Stage your changes
git add .

# Generate commit message
git ai commit

# Auto-approve and commit
git ai commit --auto

# Include detailed description
git ai commit --with-descriptions

# Use conventional format (type(scope): description)
git ai commit --conventional

# Avoid conventional format
git ai commit --no-conventional

# Amend previous commit
git ai commit --amend

# Use specific config file
git ai --config /path/to/config.yaml commit
```

Git AI analyzes your commit history to detect conventional commit format usage. When over 50% of recent commits follow the `type(scope): description` pattern, Git AI defaults to this style.

Override detection with `--conventional` or `--no-conventional` flags. Git AI saves your format and description preferences for future commits.

### Setting Git Config Options

Set preferences directly with Git's config system:

```bash
# Set conventional commit format preference
git config git-ai.conventionalCommits true

# Set detailed descriptions preference
git config git-ai.commitsWithDescriptions true
```

## How it works

Git AI analyzes your staged changes and commit history, then sends this data to your configured LLM to generate relevant commit messages. The prompt includes:

1. Staged changes diff
2. Changed files list
3. Recent commit messages
4. Instructions for message formatting

Git AI presents an interactive terminal UI to approve, edit, or cancel the proposed message.

## Supported LLM Providers

- OpenAI (GPT-4, GPT-3.5)
- Anthropic (Claude)
- Ollama (local deployment)
- Custom providers via API endpoints

## Customizing Prompts

Git AI embeds prompt templates from text files into the binary at compile time. Edit files in `pkg/llm/prompts/` before building:

- `commit_system.txt`: LLM instructions with sections for conventional vs. standard format
- `commit_user.txt`: User prompt template with placeholders for content

Both files use Go's template syntax:
- `{{if .UseConventional}}...{{else}}...{{end}}` controls format instructions
- `{{.Diff}}`, `{{.ChangedFiles}}`, `{{.RecentCommits}}` insert content

Rebuild with `go build` after modifying prompts.
