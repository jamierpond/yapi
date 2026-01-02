#!/bin/bash
set -e

echo "Installing yapi for Linux..."

# Detect architecture
ARCH=$(uname -m)
if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
  ASSET="yapi_linux_arm64.tar.gz"
elif [ "$ARCH" = "x86_64" ]; then
  ASSET="yapi_linux_amd64.tar.gz"
else
  echo "Unsupported architecture: $ARCH"
  exit 1
fi

BASE_URL="https://github.com/jamierpond/yapi/releases/latest/download"

# Download
TMPDIR=$(mktemp -d)
cd "$TMPDIR"
curl -sL "$BASE_URL/$ASSET" -o "$ASSET"
curl -sL "$BASE_URL/checksums.txt" -o checksums.txt

# Verify checksum
echo "Verifying checksum..."
EXPECTED=$(grep "$ASSET" checksums.txt | awk '{print $1}')
ACTUAL=$(sha256sum "$ASSET" | awk '{print $1}')
if [ "$EXPECTED" != "$ACTUAL" ]; then
  echo "Checksum verification failed!"
  echo "Expected: $EXPECTED"
  echo "Actual:   $ACTUAL"
  rm -rf "$TMPDIR"
  exit 1
fi
echo "Checksum verified."

# Extract and install
tar xzf "$ASSET"
sudo mv yapi /usr/local/bin/
rm -rf "$TMPDIR"

echo "yapi installed successfully!"
yapi version
