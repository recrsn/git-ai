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
  print_step "Installing from local binary found at $LOCAL_BINARY"
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
    # Get OS and architecture in the format used by GoReleaser
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    if [ "$ARCH" = "x86_64" ]; then
      ARCH="x86_64"
    elif [ "$ARCH" = "amd64" ]; then
      ARCH="x86_64"
    elif [ "$ARCH" = "arm64" ]; then
      ARCH="arm64"
    else
      print_error "Unsupported architecture: $ARCH"
      exit 1
    fi

    # Convert macOS to darwin for GoReleaser naming
    if [ "$OS" = "darwin" ]; then
      OS_TITLE="Darwin"
    elif [ "$OS" = "linux" ]; then
      OS_TITLE="Linux"
    elif [ "$OS" = "windows" ]; then
      OS_TITLE="Windows"
    else
      print_error "Unsupported OS: $OS"
      exit 1
    fi

    # Form the expected archive name
    ARCHIVE_NAME="git-ai_${OS_TITLE}_${ARCH}"
    if [ "$OS" = "windows" ]; then
      ARCHIVE_EXT=".zip"
    else
      ARCHIVE_EXT=".tar.gz"
    fi

    # Get the download URL for the appropriate archive
    LATEST_RELEASE_JSON=$(curl -s https://api.github.com/repos/recrsn/git-ai/releases/latest)
    LATEST_RELEASE_URL=$(echo "$LATEST_RELEASE_JSON" | grep -o "\"browser_download_url\":[[:space:]]*\"[^\"]*${ARCHIVE_NAME}${ARCHIVE_EXT}\"" | cut -d '"' -f 4)

    if [ -z "$LATEST_RELEASE_URL" ]; then
      print_error "Could not find a release for your platform ($OS, $ARCH)"
      exit 1
    fi

    print_step "Downloading from: $LATEST_RELEASE_URL"

    # Create a temporary directory for the download
    TMP_DIR=$(mktemp -d)
    TMP_ARCHIVE="$TMP_DIR/archive$ARCHIVE_EXT"

    # Download the archive
    curl -L -o "$TMP_ARCHIVE" "$LATEST_RELEASE_URL"

    # Extract the binary
    if [ "$OS" = "windows" ]; then
      # For Windows (zip file)
      unzip -o "$TMP_ARCHIVE" -d "$TMP_DIR"
      mv "$TMP_DIR/$BINARY_NAME.exe" "$INSTALL_DIR/$BINARY_NAME"
    else
      # For Unix-like systems (tar.gz)
      tar -xzf "$TMP_ARCHIVE" -C "$TMP_DIR"
      mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    fi

    # Clean up the temporary directory
    rm -rf "$TMP_DIR"

    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    print_success "Downloaded and installed git-ai to $INSTALL_DIR/$BINARY_NAME"
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
fi

print_success "Installation complete!"
print_step "Run 'git ai config' to set up your LLM provider"
