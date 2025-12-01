#!/bin/sh
# Run all example yapi files using the Go CLI
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CLI_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
ROOT_DIR="$(cd "$CLI_DIR/.." && pwd)"

echo "Testing all example files..."
find "$ROOT_DIR/examples" -name "*.yapi.yml" -o -name "*.yapi.yaml" | while read -r example; do
  echo "Testing: $example"
  "$CLI_DIR/yapi" run "$example" || exit 1
done
echo "All examples tested successfully"
