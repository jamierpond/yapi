#!/bin/bash
set -e

echo "=== Vercel Build Script ==="

# Install Go if not available
if ! command -v go &> /dev/null; then
    echo "Installing Go..."
    curl -fsSL https://go.dev/dl/go1.23.4.linux-amd64.tar.gz -o /tmp/go.tar.gz
    tar -C /usr/local -xzf /tmp/go.tar.gz
    export PATH="/usr/local/go/bin:$PATH"
    rm /tmp/go.tar.gz
fi

echo "Go version: $(go version)"

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

# Install to PATH
echo "Installing yapi to /usr/local/bin..."
mkdir -p /usr/local/bin
cp ./bin/yapi /usr/local/bin/yapi
chmod +x /usr/local/bin/yapi

echo "yapi installed successfully: $(which yapi)"
yapi --version 2>/dev/null || echo "yapi built (version flag may not be implemented)"

# Run pnpm build for web
echo "Running pnpm build..."
cd web
pnpm install
pnpm run build

echo "=== Build Complete ==="
