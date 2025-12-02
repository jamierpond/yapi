#!/bin/bash
# run all example yapi files
set -eou pipefail

script_dir="$(cd "$(dirname "$0")" && pwd)"
root_dir="$(cd "$script_dir/.." && pwd)"

echo "testing all example files..."
find "$root_dir/examples" -type f \( -name "*.yml" -o -name "*.yaml" \) | sort | while read -r example; do
  echo "testing: $example"
  yapi run "$example" || exit 1
done
echo "all examples tested successfully"
