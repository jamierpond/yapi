#!/bin/bash
# run all example yapi files in parallel using GNU parallel
set -eou pipefail

script_dir="$(cd "$(dirname "$0")" && pwd)"
root_dir="$(cd "$script_dir/.." && pwd)"

# Check if parallel is installed
if ! command -v parallel &> /dev/null; then
  echo "error: GNU parallel is not installed"
  echo "install with: brew install parallel"
  exit 1
fi

echo "testing all example files in parallel..."

find "$root_dir/examples" -type f \( -name "*.yml" -o -name "*.yaml" \) | \
  grep -v -e '/invalid/' -e '\.fail\.' -e '\.local\.' | sort | \
  parallel --halt now,fail=1 --jobs 25 --tag 'yapi run {}'

echo "all examples tested successfully"
