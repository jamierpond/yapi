#!/bin/bash
set -e

cd "$(git rev-parse --show-toplevel)" || exit 1

# Only allow releases from main, develop, or next branches
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [[ "$CURRENT_BRANCH" != "main" && "$CURRENT_BRANCH" != "develop" && "$CURRENT_BRANCH" != "next" ]]; then
    echo "Error: Releases can only be made from 'main', 'develop', or 'next' branches"
    echo "Current branch: $CURRENT_BRANCH"
    exit 1
fi

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

# Build the action before tagging
echo "Building action..."
cd action
pnpm install
if [ $? -ne 0 ]; then
    echo "Error: Action install failed"
    exit 1
fi
pnpm run build
if [ $? -ne 0 ]; then
    echo "Error: Action build failed"
    exit 1
fi
cd ..
echo "Action built successfully"

# Sync version to all package.json files
VERSION_NUMBER="${NEW_VERSION#v}"
echo "Syncing package.json files to version $VERSION_NUMBER..."

for pkg in $(git ls-files '*/package.json' 'package.json'); do
    if [ -f "$pkg" ]; then
        sed -i '' "s/\"version\": \"[^\"]*\"/\"version\": \"$VERSION_NUMBER\"/" "$pkg"
        echo "  Updated $pkg"
    fi
done

# Commit version bump and tag
git add -A
git commit -m "[release] $NEW_VERSION"
git tag "$NEW_VERSION"
echo "Committed and tagged $NEW_VERSION"
