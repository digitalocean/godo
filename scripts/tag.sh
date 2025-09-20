#!/usr/bin/env bash

set -eo pipefail

ORIGIN=${ORIGIN:-origin}
COMMIT=${COMMIT:-HEAD}

if [[ $(git status --porcelain) != "" ]]; then
  echo "Error: repo is dirty. Run git status, clean repo and try again."
  exit 1
elif [[ $(git status --porcelain -b | grep -e "ahead" -e "behind") != "" ]]; then
  echo "Error: repo has unpushed commits. Push commits to remote and try again."
  exit 1
fi  

# Check if user has push access
if ! git ls-remote --exit-code "$ORIGIN" >/dev/null 2>&1; then
  echo "Error: Cannot access remote repository. Ensure you have push permissions."
  exit 1
fi  

# Get version from godo.go
version=$(grep -E '^\s*libraryVersion\s*=' godo.go | sed 's/.*"\(.*\)".*/\1/')
tag="v${version}"
changelog=$(awk -v ver="$version" '/^## \['ver'\]/{flag=1;next}/^## \[/{flag=0}flag' CHANGELOG.md)
git tag -a "$tag" -m "release $tag\n$changelog" $COMMIT && git push "$ORIGIN" tag "$tag"

if [ -z "$version" ]; then
    echo "Error: Could not find version in godo.go"
    exit 1
fi

git tag -m "release $tag" -a "$tag" $COMMIT && git push "$ORIGIN" tag "$tag"

echo ""