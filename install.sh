#!/usr/bin/env bash
set -e

# Colors for pretty output
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_step() {
  echo -e "${BLUE}==>${NC} $1"
}

print_success() {
  echo -e "${GREEN}==>${NC} $1"
}

print_warning() {
  echo -e "${YELLOW}==>${NC} $1"
}

print_error() {
  echo -e "${RED}==>${NC} $1"
}

# Check if curl is installed
if ! command -v curl &> /dev/null; then
  print_error "curl is required but not installed. Please install curl and try again."
  exit 1
fi

# Create ~/.local/bin if it doesn't exist
INSTALL_DIR="$HOME/.local/bin"
if [ ! -d "$INSTALL_DIR" ]; then
  print_step "Creating directory $INSTALL_DIR"
  mkdir -p "$INSTALL_DIR"
fi

# Check if git-ai exists locally
BINARY_NAME="git-ai"
LOCAL_BINARY="$(pwd)/$BINARY_NAME"

if [ -f "$LOCAL_BINARY" ]; then
  print_step "Found local binary at $LOCAL_BINARY"
  cp "$LOCAL_BINARY" "$INSTALL_DIR/$BINARY_NAME"
  print_success "Local binary copied to $INSTALL_DIR/$BINARY_NAME"
else
  # If no local binary, build it if we're in the repo
  if [ -f "$(pwd)/go.mod" ] && grep -q "github.com/recrsn/git-ai" "$(pwd)/go.mod"; then
    print_step "Local binary not found, but we're in the git-ai repository. Building..."
    go build
    if [ -f "$LOCAL_BINARY" ]; then
      cp "$LOCAL_BINARY" "$INSTALL_DIR/$BINARY_NAME"
      print_success "Successfully built and copied to $INSTALL_DIR/$BINARY_NAME"
    else
      print_error "Build failed. Attempting to download from GitHub releases..."
    fi
  else
    print_step "No local binary found, downloading from GitHub releases..."

    # Get latest release from GitHub API
    LATEST_RELEASE_URL=$(curl -s https://api.github.com/repos/recrsn/git-ai/releases/latest | grep "browser_download_url.*$(uname -s | tr '[:upper:]' '[:lower:]')*$(uname -m)*" | cut -d : -f 2,3 | tr -d \")

    if [ -z "$LATEST_RELEASE_URL" ]; then
      print_error "Could not find a release for your platform ($(uname -s), $(uname -m))"
      exit 1
    fi

    print_step "Downloading from: $LATEST_RELEASE_URL"
    curl -L -o "$INSTALL_DIR/$BINARY_NAME" "$LATEST_RELEASE_URL"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    print_success "Downloaded git-ai to $INSTALL_DIR/$BINARY_NAME"
  fi
fi

# Make the binary executable
chmod +x "$INSTALL_DIR/$BINARY_NAME"

# Check if ~/.local/bin is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
  print_warning "$INSTALL_DIR is not in your PATH"

  # Determine shell and provide appropriate command
  SHELL_NAME="$(basename "$SHELL")"
  case "$SHELL_NAME" in
    bash)
      print_step "Run this command to add to your PATH:"
      echo "echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.bashrc && source ~/.bashrc"
      ;;
    zsh)
      print_step "Run this command to add to your PATH:"
      echo "echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.zshrc && source ~/.zshrc"
      ;;
    fish)
      print_step "Run this command to add to your PATH:"
      echo "fish_add_path $INSTALL_DIR && source ~/.config/fish/config.fish"
      ;;
    *)
      print_step "Run this command to add to your PATH:"
      echo "echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.profile && source ~/.profile"
      ;;
  esac
else
  print_success "$INSTALL_DIR is already in your PATH"
fi

print_success "Installation complete!"
print_step "Run 'git ai config' to set up your LLM provider"
