#!/bin/sh
# Run all bats tests
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Running all tests..."
bats "$ROOT_DIR"/test/*.bats
echo "All tests passed"
