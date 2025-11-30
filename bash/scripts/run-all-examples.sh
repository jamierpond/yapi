#!/bin/sh
# Run all example yapi files
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Testing all example files..."
for example in "$ROOT_DIR"/examples/*.yml "$ROOT_DIR"/examples/*.yaml; do
  if [ -f "$example" ]; then
    echo "Testing: $example"
    "$ROOT_DIR/yapi" -c "$example" || exit 1
  fi
done
echo "All examples tested successfully"
