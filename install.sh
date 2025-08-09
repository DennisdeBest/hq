#!/usr/bin/env bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# GitHub repository information
REPO_OWNER="dennisdebest"
REPO_NAME="hq"
GITHUB_API="https://api.github.com"
GITHUB_REPO="$GITHUB_API/repos/$REPO_OWNER/$REPO_NAME"

# Installation directory
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
fi

echo -e "${GREEN}Installing hq...${NC}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture to Go's GOARCH
case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo -e "${RED}Unsupported architecture: $ARCH${NC}"
        echo "hq is available for amd64 and arm64 architectures."
        exit 1
        ;;
esac

# Check if OS is supported
if [ "$OS" != "linux" ] && [ "$OS" != "darwin" ] && [ "$OS" != "windows" ]; then
    echo -e "${RED}Unsupported operating system: $OS${NC}"
    echo "hq is available for Linux, macOS, and Windows."
    exit 1
fi

# Special handling for Windows
if [ "$OS" = "windows" ]; then
    echo -e "${YELLOW}Windows detected. Please note that this installer script works best on Linux/macOS.${NC}"
    echo -e "${YELLOW}For Windows, you might need to manually place the executable in your PATH.${NC}"
    # Set a default install directory for Windows
    INSTALL_DIR="$HOME/bin"
    mkdir -p "$INSTALL_DIR"
fi

echo -e "${YELLOW}Detected OS: $OS, Architecture: $ARCH${NC}"

# Get the latest release
echo -e "${YELLOW}Fetching the latest release...${NC}"
LATEST_RELEASE=$(curl -s "$GITHUB_REPO/releases/latest")
TAG_NAME=$(echo "$LATEST_RELEASE" | grep -o '"tag_name": "[^"]*' | cut -d'"' -f4)

if [ -z "$TAG_NAME" ]; then
    echo -e "${RED}Failed to fetch the latest release.${NC}"
    exit 1
fi

echo -e "${GREEN}Latest release: $TAG_NAME${NC}"

# Construct the download URL based on OS and architecture
# The URL format matches the artifact paths in the GitHub Actions workflow
if [ "$OS" = "windows" ]; then
    DOWNLOAD_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$TAG_NAME/hq-$OS-$ARCH-$TAG_NAME.exe"
else
    DOWNLOAD_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$TAG_NAME/hq-$OS-$ARCH-$TAG_NAME"
fi

# Create a temporary directory
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

# Download the binary
echo -e "${YELLOW}Downloading hq...${NC}"
# Extract the filename from the download URL
BINARY_FILENAME=$(basename "$DOWNLOAD_URL")
curl -sL "$DOWNLOAD_URL" -o "$TMP_DIR/$BINARY_FILENAME"

# Make the binary executable
chmod +x "$TMP_DIR/$BINARY_FILENAME"

# Move the binary to the installation directory with a simplified name
echo -e "${YELLOW}Installing to $INSTALL_DIR...${NC}"
# Use 'hq' as the final name (or 'hq.exe' for Windows)
FINAL_NAME="hq"
if [ "$OS" = "windows" ]; then
    FINAL_NAME="hq.exe"
fi
mv "$TMP_DIR/$BINARY_FILENAME" "$INSTALL_DIR/$FINAL_NAME"

# Check if the installation directory is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo -e "${YELLOW}Adding $INSTALL_DIR to your PATH...${NC}"
    echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$HOME/.bashrc"
    echo "You may need to restart your shell or run 'source ~/.bashrc' to use hq."
fi

echo -e "${GREEN}hq $TAG_NAME has been installed successfully!${NC}"
echo -e "Run ${YELLOW}hq --help${NC} to get started."