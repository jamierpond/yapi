#!/bin/bash
set -e

IS_DARWIN=false
if [ "$(uname -s)" = "Darwin" ]; then
    IS_DARWIN=true
    echo "Running on macOS - skipping Go build, only building web"
fi

echo "=== Vercel Build Script ==="

# Install Go if not available
if ! command -v go &> /dev/null; then
    echo "Installing Go..."
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
    esac

    GO_VERSION="1.23.4"
    curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz" -o /tmp/go.tar.gz
    tar -C /usr/local -xzf /tmp/go.tar.gz
    export PATH="/usr/local/go/bin:$PATH"
    rm /tmp/go.tar.gz
fi

echo "Go version: $(go version)"

# Build variables
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "")
if [ -z "$COMMIT" ] && [ -n "$VERCEL_GIT_COMMIT_SHA" ]; then
    COMMIT="${VERCEL_GIT_COMMIT_SHA:0:7}"
fi
COMMIT="${COMMIT:-unknown}"
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS="-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"


echo "Building yapi CLI..."
go build -ldflags "${LDFLAGS}" -o ./bin/yapi ./cmd/yapi

echo "Installing yapi to /usr/local/bin..."
mkdir -p /usr/local/bin
cp ./bin/yapi /usr/local/bin/yapi
chmod +x /usr/local/bin/yapi

echo "yapi installed: $(which yapi)"
yapi version 2>/dev/null || true

echo "Generating CLI documentation..."
go run scripts/gendocs.go

echo "Building madea-blog-core..."
pnpm --filter madea-blog-core build

echo "Running pnpm build for web..."
pnpm --filter web build

echo "=== Build Complete ==="
