#!/bin/bash
# yapi quickstart - Clone and configure yapi
set -e

YAPI_HOME="${YAPI_HOME:-$HOME/.config/yapi}"
REPO_URL="${YAPI_REPO_URL:-https://github.com/jamiepond/yapi.git}"
REPO_BRANCH="${YAPI_BRANCH:-main}"

echo "Installing yapi to $YAPI_HOME..."

# Clone or update
if [ -d "$YAPI_HOME/.git" ]; then
  echo "Updating existing installation..."
  git -C "$YAPI_HOME" pull --ff-only
else
  mkdir -p "$(dirname "$YAPI_HOME")"
  git clone --branch "$REPO_BRANCH" "$REPO_URL" "$YAPI_HOME"
fi

# Make executable
chmod +x "$YAPI_HOME/yapi" "$YAPI_HOME/lib/"*.sh

# Detect shell
SHELL_NAME=$(basename "$SHELL")
case "$SHELL_NAME" in
  zsh)  SHELLRC="$HOME/.zshrc" ;;
  bash) SHELLRC="$HOME/.bashrc" ;;
  *)    SHELLRC="$HOME/.${SHELL_NAME}rc" ;;
esac

# Config to add (uses single quotes so $HOME expands at runtime)
CONFIG='
# yapi - YAML API Testing Tool
YAPI_HOME="${YAPI_HOME:-$HOME/.config/yapi}"
[ -f "$YAPI_HOME/bin/yapi.zsh" ] && source "$YAPI_HOME/bin/yapi.zsh"
alias a="yapi"'

# Check if already configured
if grep -q "YAPI_HOME" "$SHELLRC" 2>/dev/null; then
  echo "yapi already configured in $SHELLRC"
else
  echo "$CONFIG" >> "$SHELLRC"
  echo "Added yapi config to $SHELLRC"
fi

echo ""
echo "Installation complete!"
echo "Run 'source $SHELLRC' or restart your shell to use yapi."
echo "Usage: yapi -c config.yaml  or  a -c config.yaml"
