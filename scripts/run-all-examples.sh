#!/bin/sh
# Run all example yapi files
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Testing all example files..."
find "$ROOT_DIR/examples" -type f \( -name "*.yml" -o -name "*.yaml" \) | sort | while read -r example; do
  echo "Testing: $example"
  yapi run "$example" || exit 1
done
echo "All examples tested successfully"
