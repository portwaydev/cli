#!/bin/bash

# Exit on error
set -e

# Get the latest release version
LATEST_VERSION=$(curl -s https://api.github.com/repos/portwaydev/cli/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

# Determine OS and architecture
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
    OS="windows"
    # Check if running in WSL
    if grep -q Microsoft /proc/version 2>/dev/null; then
        OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    fi
else
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
fi

ARCH=$(uname -m)

# Map architecture to GitHub's naming convention
case $ARCH in
  x86_64)
    ARCH="amd64"
    ;;
  aarch64)
    ARCH="arm64"
    ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# Download URL
DOWNLOAD_URL="https://github.com/portwaydev/cli/releases/download/${LATEST_VERSION}/portway-${OS}-${ARCH}"

# Download the binary
echo "Downloading version ${LATEST_VERSION}..."
curl -L -o yourbinary "${DOWNLOAD_URL}"

# Make it executable
chmod +x yourbinary

# Move to a directory in PATH
if [[ "$OS" == "windows" ]]; then
    # For Windows, move to a directory in PATH
    INSTALL_DIR="$HOME/AppData/Local/Programs/portway"
    mkdir -p "$INSTALL_DIR"
    mv yourbinary "$INSTALL_DIR/portway.exe"
    echo "Please add $INSTALL_DIR to your PATH environment variable"
else
    # For Unix-like systems
    sudo mv yourbinary /usr/local/bin/portway
fi

echo "Installation complete! Version ${LATEST_VERSION} has been installed."
