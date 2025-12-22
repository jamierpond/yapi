#!/bin/bash
set -e

cd "$(git rev-parse --show-toplevel)" || exit 1

# Get highest semantic version tag (must match vX.Y.Z pattern)
CURRENT_TAG=$(git tag -l 'v*.*.*' | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | tail -1)
CURRENT_TAG=${CURRENT_TAG:-v0.0.0}

MAJOR=$(echo "$CURRENT_TAG" | sed 's/v//' | cut -d. -f1)
MINOR=$(echo "$CURRENT_TAG" | sed 's/v//' | cut -d. -f2)
PATCH=$(echo "$CURRENT_TAG" | sed 's/v//' | cut -d. -f3)

bump_type="${1:-patch}"

case "$bump_type" in
    patch)
        NEW_VERSION="v${MAJOR}.${MINOR}.$((PATCH + 1))"
        ;;
    minor)
        NEW_VERSION="v${MAJOR}.$((MINOR + 1)).0"
        ;;
    major)
        NEW_VERSION="v$((MAJOR + 1)).0.0"
        ;;
    *)
        echo "Usage: $0 [patch|minor|major]"
        exit 1
        ;;
esac

echo "Current highest version: $CURRENT_TAG"

if git tag -l | grep -q "^${NEW_VERSION}$"; then
    echo "Error: $NEW_VERSION already exists"
    exit 1
fi

echo "New version: $NEW_VERSION"

# Update gh-action package.json
SEMVER="${MAJOR}.${MINOR}.${PATCH}"
if [ -d "gh-action" ]; then
    echo "Updating gh-action/package.json to $SEMVER..."

    if command -v jq &> /dev/null; then
        jq ".version = \"$SEMVER\"" gh-action/package.json > gh-action/package.json.tmp && mv gh-action/package.json.tmp gh-action/package.json
    else
        # Fallback to node if jq is not available
        cd gh-action
        node -e "const pkg = require('./package.json'); pkg.version = '$SEMVER'; require('fs').writeFileSync('package.json', JSON.stringify(pkg, null, 2) + '\n');"
        cd ..
    fi

    # Build the action
    echo "Building gh-action..."
    cd gh-action
    pnpm run build 2>/dev/null || npm run build
    cd ..

    # Commit gh-action changes
    git add gh-action/package.json gh-action/dist/
    git commit -m "Bump gh-action version to $SEMVER"

    echo "gh-action updated and committed"
fi

# Create main version tag
git tag "$NEW_VERSION"
echo "Tagged $NEW_VERSION - run 'make release' to push"
