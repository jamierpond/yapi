#!/bin/bash
set -e

echo "=== Vercel Build Script ==="

# Working directory is the project root when called from vercel.json
PROJECT_ROOT="$(pwd)"

echo "Project root: $PROJECT_ROOT"

# Install Go if not available
if ! command -v go &> /dev/null; then
    echo "Installing Go..."

    # Detect OS and architecture
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    # Map architecture names
    case "$ARCH" in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
    esac

    # Map OS names for Go download
    case "$OS" in
        darwin) GO_OS="darwin" ;;
        linux) GO_OS="linux" ;;
        *) echo "Unsupported OS: $OS"; exit 1 ;;
    esac

    GO_VERSION="1.23.4"
    GO_URL="https://go.dev/dl/go${GO_VERSION}.${GO_OS}-${ARCH}.tar.gz"

    echo "Downloading Go from: $GO_URL"
    curl -fsSL "$GO_URL" -o /tmp/go.tar.gz

    if [ "$OS" = "darwin" ]; then
        # macOS - install to user directory if no sudo
        GO_INSTALL_DIR="${HOME}/.local/go"
        mkdir -p "$GO_INSTALL_DIR"
        tar -C "$(dirname "$GO_INSTALL_DIR")" -xzf /tmp/go.tar.gz
        export PATH="${GO_INSTALL_DIR}/bin:$PATH"
    else
        # Linux (Vercel) - install to /usr/local
        tar -C /usr/local -xzf /tmp/go.tar.gz
        export PATH="/usr/local/go/bin:$PATH"
    fi

    rm /tmp/go.tar.gz
fi

echo "Go version: $(go version)"

# Already in project root

# Build variables (matching Makefile)
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS="-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

# Add PostHog keys if available
if [ -n "$POSTHOG_API_KEY" ]; then
    LDFLAGS="${LDFLAGS} -X yapi.run/cli/internal/telemetry.PosthogAPIKey=${POSTHOG_API_KEY}"
fi
if [ -n "$POSTHOG_API_HOST" ]; then
    LDFLAGS="${LDFLAGS} -X yapi.run/cli/internal/telemetry.PosthogAPIHost=${POSTHOG_API_HOST}"
fi

echo "Building yapi CLI..."
go build -ldflags "${LDFLAGS}" -o ./bin/yapi ./cmd/yapi

# Install to GOPATH/bin (same as Makefile)
INSTALL_DIR="$(go env GOPATH)/bin"
echo "Installing yapi to $INSTALL_DIR..."
mkdir -p "$INSTALL_DIR"
cp ./bin/yapi "$INSTALL_DIR/yapi"
codesign --sign - --force "$INSTALL_DIR/yapi" 2>/dev/null || true

echo "yapi installed successfully: $(which yapi)"
yapi --version 2>/dev/null || echo "yapi built (version flag may not be implemented)"

# Run pnpm build for web via workspace filter
echo "Running pnpm build for web..."
pnpm --filter web build

echo "=== Build Complete ==="
