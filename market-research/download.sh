#!/bin/bash

set -e

OUTPUT_DIR="/Users/jamiepond/projects/yapi/market-research"
cd "$OUTPUT_DIR"

echo "=== Downloading .http files ==="
gh api search/code -X GET -f q="extension:http language:HTTP" -f per_page=100 --jq '.items[:40] | .[] | "\(.repository.full_name)|\(.path)"' > /tmp/http_files.txt

while IFS='|' read -r repo path; do
    # Skip unwanted paths
    case "$path" in
        *.history*|*node_modules*|*Dockerfile*|*.HTTP) continue ;;
    esac
    [[ "$path" != *.http ]] && continue

    safe_name=$(echo "${repo}__${path}" | tr '/' '_' | tr ' ' '_')

    echo "Fetching: $repo - $path"
    if gh api "repos/$repo/contents/$path" --jq '.content' 2>/dev/null | base64 -d > "http/$safe_name" 2>/dev/null; then
        if [ -s "http/$safe_name" ]; then
            echo "  Saved: $safe_name"
        else
            rm -f "http/$safe_name"
        fi
    fi
done < /tmp/http_files.txt

echo ""
echo "=== Downloading .rest files ==="
gh api search/code -X GET -f q="extension:rest" -f per_page=100 --jq '.items[:40] | .[] | "\(.repository.full_name)|\(.path)"' > /tmp/rest_files.txt

while IFS='|' read -r repo path; do
    case "$path" in
        *.history*|*node_modules*|*vendor*) continue ;;
    esac
    [[ "$path" != *.rest ]] && continue

    safe_name=$(echo "${repo}__${path}" | tr '/' '_' | tr ' ' '_')

    echo "Fetching: $repo - $path"
    if gh api "repos/$repo/contents/$path" --jq '.content' 2>/dev/null | base64 -d > "rest/$safe_name" 2>/dev/null; then
        if [ -s "rest/$safe_name" ]; then
            echo "  Saved: $safe_name"
        else
            rm -f "rest/$safe_name"
        fi
    fi
done < /tmp/rest_files.txt

echo ""
echo "=== Downloading .bru files (Bruno) ==="
gh api search/code -X GET -f q="extension:bru" -f per_page=100 --jq '.items[:40] | .[] | "\(.repository.full_name)|\(.path)"' > /tmp/bru_files.txt

while IFS='|' read -r repo path; do
    case "$path" in
        *.history*|*node_modules*|*vendor*) continue ;;
    esac
    [[ "$path" != *.bru ]] && continue

    safe_name=$(echo "${repo}__${path}" | tr '/' '_' | tr ' ' '_')

    echo "Fetching: $repo - $path"
    if gh api "repos/$repo/contents/$path" --jq '.content' 2>/dev/null | base64 -d > "bru/$safe_name" 2>/dev/null; then
        if [ -s "bru/$safe_name" ]; then
            echo "  Saved: $safe_name"
        else
            rm -f "bru/$safe_name"
        fi
    fi
done < /tmp/bru_files.txt

echo ""
echo "=== Summary ==="
echo "HTTP files: $(ls -1 http/ 2>/dev/null | wc -l)"
echo "REST files: $(ls -1 rest/ 2>/dev/null | wc -l)"
echo "BRU files: $(ls -1 bru/ 2>/dev/null | wc -l)"
