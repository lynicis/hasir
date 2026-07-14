#!/usr/bin/env bash
# path: scripts/release.sh
set -euo pipefail

if [ $# -lt 2 ]; then
  echo "Usage: scripts/release.sh <app> <patch|minor|major>"
  echo "Example: scripts/release.sh api minor"
  exit 1
fi

APP=$1
BUMP=$2

# Determine current version from git tags
LATEST=$(git tag -l "$APP/v*" | sort -V | tail -n1 || true)
if [ -z "$LATEST" ]; then
  echo "No existing tag found for $APP. Starting at v0.1.0."
  NEW_VERSION="v0.1.0"
else
  # simple semantic version bump logic
  VERSION=${LATEST#"$APP/v"}
  IFS='.' read -r -a PARTS <<< "$VERSION"
  
  # Ensure we have 3 parts
  if [ ${#PARTS[@]} -ne 3 ]; then
    echo "Error: Latest tag '$LATEST' is not in semver format (vX.Y.Z)"
    exit 1
  fi
  
  case $BUMP in
    major) NEW_VERSION="v$((PARTS[0]+1)).0.0" ;;
    minor) NEW_VERSION="v${PARTS[0]}.$((PARTS[1]+1)).0" ;;
    patch) NEW_VERSION="v${PARTS[0]}.${PARTS[1]}.$((PARTS[2]+1))" ;;
    *) echo "Unknown bump type $BUMP. Must be major, minor, or patch."; exit 1 ;;
  esac
fi

TAG="$APP/$NEW_VERSION"
echo "Creating tag $TAG"
git tag "$TAG"
echo "Pushing tag $TAG to origin..."
git push origin "$TAG"
