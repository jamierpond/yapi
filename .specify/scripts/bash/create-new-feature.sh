#!/usr/bin/env bash

set -e

JSON_MODE=false
BRANCH_NUMBER=""
ARGS=()
i=1
while [ $i -le $# ]; do
    arg="${!i}"
    case "$arg" in
        --json)
            JSON_MODE=true
            ;;
        --number)
            if [ $((i + 1)) -gt $# ]; then
                echo 'Error: --number requires a value' >&2
                exit 1
            fi
            i=$((i + 1))
            next_arg="${!i}"
            if [[ "$next_arg" == --* ]]; then
                echo 'Error: --number requires a value' >&2
                exit 1
            fi
            BRANCH_NUMBER="$next_arg"
            ;;
        --help|-h)
            echo "Usage: $0 [--json] [--number N]"
            echo ""
            echo "Creates a spec directory based on your current git branch."
            echo ""
            echo "Options:"
            echo "  --json       Output in JSON format"
            echo "  --number N   Specify spec number manually (overrides auto-detection)"
            echo "  --help, -h   Show this help message"
            echo ""
            echo "Examples:"
            echo "  git checkout -b jp/oauth-fix"
            echo "  $0                          # Creates specs/001-jp-oauth-fix"
            echo "  $0 --number 5               # Creates specs/005-jp-oauth-fix"
            exit 0
            ;;
        *)
            ARGS+=("$arg")
            ;;
    esac
    i=$((i + 1))
done

# Function to find the repository root by searching for existing project markers
find_repo_root() {
    local dir="$1"
    while [ "$dir" != "/" ]; do
        if [ -d "$dir/.git" ] || [ -d "$dir/.specify" ]; then
            echo "$dir"
            return 0
        fi
        dir="$(dirname "$dir")"
    done
    return 1
}

# Function to get highest number from specs directory
get_highest_from_specs() {
    local specs_dir="$1"
    local highest=0

    if [ -d "$specs_dir" ]; then
        for dir in "$specs_dir"/*; do
            [ -d "$dir" ] || continue
            dirname=$(basename "$dir")
            number=$(echo "$dirname" | grep -o '^[0-9]\+' || echo "0")
            number=$((10#$number))
            if [ "$number" -gt "$highest" ]; then
                highest=$number
            fi
        done
    fi

    echo "$highest"
}

# Function to clean and format a branch name for folder use
clean_branch_name() {
    local name="$1"
    echo "$name" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g' | sed 's/-\+/-/g' | sed 's/^-//' | sed 's/-$//'
}

# Resolve repository root
SCRIPT_DIR="$(CDPATH="" cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if git rev-parse --show-toplevel >/dev/null 2>&1; then
    REPO_ROOT=$(git rev-parse --show-toplevel)
    HAS_GIT=true
else
    REPO_ROOT="$(find_repo_root "$SCRIPT_DIR")"
    if [ -z "$REPO_ROOT" ]; then
        echo "Error: Could not determine repository root. Please run this script from within the repository." >&2
        exit 1
    fi
    HAS_GIT=false
fi

cd "$REPO_ROOT"

SPECS_DIR="$REPO_ROOT/specs"
mkdir -p "$SPECS_DIR"

# --- Use current git branch ---

if [ "$HAS_GIT" = "false" ]; then
    echo "ERROR: This workflow requires a git repository." >&2
    exit 1
fi

CURRENT_GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)

# 1. Block Main/Develop
if [[ "$CURRENT_GIT_BRANCH" =~ ^(main|master|develop|dev|staging)$ ]]; then
    echo "ERROR: You are on '$CURRENT_GIT_BRANCH'." >&2
    echo "Please 'git checkout -b my-feature-name' first, then run this." >&2
    exit 1
fi

# 2. Sanitize branch name for folder (jp/auth-fix -> jp-auth-fix)
BRANCH_SUFFIX=$(clean_branch_name "$CURRENT_GIT_BRANCH")

# 3. Calculate Next Number
if [ -z "$BRANCH_NUMBER" ]; then
    HIGHEST=$(get_highest_from_specs "$SPECS_DIR")
    BRANCH_NUMBER=$((HIGHEST + 1))
fi

FEATURE_NUM=$(printf "%03d" "$((10#$BRANCH_NUMBER))")

# 4. Construct the Folder Name
NEW_FOLDER_NAME="${FEATURE_NUM}-${BRANCH_SUFFIX}"

echo "Selected Feature ID: $NEW_FOLDER_NAME (derived from branch $CURRENT_GIT_BRANCH)"

# 5. Create Directory (NO GIT CHECKOUT)
FEATURE_DIR="$SPECS_DIR/$NEW_FOLDER_NAME"

if [ -d "$FEATURE_DIR" ]; then
    echo "Resuming work in existing feature directory: $FEATURE_DIR"
else
    echo "Creating new spec directory: $FEATURE_DIR"
    mkdir -p "$FEATURE_DIR"
fi

TEMPLATE="$REPO_ROOT/.specify/templates/spec-template.md"
SPEC_FILE="$FEATURE_DIR/spec.md"
if [ -f "$TEMPLATE" ] && [ ! -f "$SPEC_FILE" ]; then
    cp "$TEMPLATE" "$SPEC_FILE"
elif [ ! -f "$SPEC_FILE" ]; then
    touch "$SPEC_FILE"
fi

# Set the SPECIFY_FEATURE environment variable for the current session
export SPECIFY_FEATURE="$NEW_FOLDER_NAME"

if $JSON_MODE; then
    printf '{"BRANCH_NAME":"%s","SPEC_FILE":"%s","FEATURE_NUM":"%s"}\n' "$NEW_FOLDER_NAME" "$SPEC_FILE" "$FEATURE_NUM"
else
    echo "BRANCH_NAME: $NEW_FOLDER_NAME"
    echo "SPEC_FILE: $SPEC_FILE"
    echo "FEATURE_NUM: $FEATURE_NUM"
    echo "SPECIFY_FEATURE environment variable set to: $NEW_FOLDER_NAME"
fi
