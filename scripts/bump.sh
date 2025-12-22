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

    cd gh-action

    # Update package.json
    if command -v jq &> /dev/null; then
        jq ".version = \"$SEMVER\"" package.json > package.json.tmp && mv package.json.tmp package.json
    else
        node -e "const pkg = require('./package.json'); pkg.version = '$SEMVER'; require('fs').writeFileSync('package.json', JSON.stringify(pkg, null, 2) + '\n');"
    fi

    # Build the action
    echo "Building gh-action..."
    pnpm run build 2>/dev/null || npm run build

    # Commit changes within the submodule
    git add package.json dist/
    git commit -m "Bump version to $SEMVER"

    # Tag the submodule
    git tag "v$SEMVER"

    cd ..

    # Update submodule reference in parent repo
    git add gh-action
    git commit -m "Update gh-action to v$SEMVER"

    echo "gh-action updated and committed"
fi

# Create main version tag
git tag "$NEW_VERSION"
echo "Tagged $NEW_VERSION - run 'make release' to push"
