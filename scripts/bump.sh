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
git tag "$NEW_VERSION"
echo "Tagged $NEW_VERSION"
