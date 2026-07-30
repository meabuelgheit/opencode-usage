#!/bin/bash
set -e

REPO="abuelgheit/opencode-usage"
BIN="opencode-usage"
INSTALL_DIR="${HOME}/.local/bin"

# Detect OS and arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64)  ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Get latest release tag
LATEST=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$LATEST" ]; then
    echo "Failed to fetch latest release. Check https://github.com/${REPO}/releases"
    exit 1
fi

echo "Installing opencode-usage ${LATEST} for ${OS}/${ARCH}..."

# Download
URL="https://github.com/${REPO}/releases/download/${LATEST}/${BIN}_${OS}_${ARCH}.tar.gz"
TMPDIR=$(mktemp -d)
curl -sSL "$URL" | tar xz -C "$TMPDIR"

# Install
mkdir -p "$INSTALL_DIR"
mv "$TMPDIR/$BIN" "$INSTALL_DIR/$BIN"
chmod +x "$INSTALL_DIR/$BIN"
rm -rf "$TMPDIR"

echo "Installed to ${INSTALL_DIR}/${BIN}"
if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
    echo "NOTE: Add ${INSTALL_DIR} to your PATH if not already:"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
fi
