# Git AI

Git AI is a Git extension that enhances your Git workflow with AI-powered features.

## Features

- `git ai commit`: Analyze staged changes and generate commit messages automatically
  - Interactive approval process with option to edit the suggested message
  - `--auto` flag to automatically approve and commit without prompt

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
```

## Usage

```bash
# Stage your changes
git add .

# Generate an AI commit message
git ai commit

# Automatically approve and commit without prompting
git ai commit --auto
```

## How it works

Git AI analyzes your staged changes and commit history to generate contextually relevant commit messages. The tool uses a terminal-based UI to provide an interactive experience, allowing you to approve, edit, or cancel the proposed commit message.

## Development

This is an early-stage project. Future improvements will include:
- Integration with popular LLM providers
- PR summarization
- Code review assistance
- Custom configuration options