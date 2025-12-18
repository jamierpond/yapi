#!/bin/bash
git_root=$(git rev-parse --show-toplevel 2>/dev/null)
cd "$git_root" || exit 1

"$HOME/.config/bin/scripts/repo-print" cmd internal -e ".*test.go$"

