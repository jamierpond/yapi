#!/bin/bash
set -e

echo "Installing yapi for macOS..."

# Detect architecture
ARCH=$(uname -m)
if [ "$ARCH" = "arm64" ]; then
  ASSET="yapi_darwin_arm64.tar.gz"
elif [ "$ARCH" = "x86_64" ]; then
  ASSET="yapi_darwin_amd64.tar.gz"
else
  echo "Unsupported architecture: $ARCH"
  exit 1
fi

# Download and extract
TMPDIR=$(mktemp -d)
cd "$TMPDIR"
curl -sL "https://github.com/jamierpond/yapi/releases/latest/download/$ASSET" | tar xz

# Install
sudo mv yapi /usr/local/bin/
rm -rf "$TMPDIR"

echo "yapi installed successfully!"
yapi version
