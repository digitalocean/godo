#!/usr/bin/env bash

set -euo pipefail

ORIGIN=${ORIGIN:-origin}

# Bump defaults to patch. We provide friendly aliases
# for patch, minor and major
BUMP=${BUMP:-patch}

# Get current version from godo.go
current_version=$(grep -E '^\s*libraryVersion\s*=' godo.go | sed 's/.*"\(.*\)".*/\1/')
version="$current_version"
IFS='.' read -r major minor patch <<< "$version"

case "$BUMP" in
  feature | minor)
    minor=$((minor + 1))
    patch=0
    ;;
  breaking | major)
    major=$((major + 1))
    minor=0
    patch=0
    ;;
  *)
    patch=$((patch + 1))
    ;;
esac

new_version="$major.$minor.$patch"

if [[ $(git status --porcelain) != "" ]]; then
  echo "Error: repo is dirty. Run git status, clean repo and try again."
  exit 1
elif [[ $(git status --porcelain -b | grep -e "ahead" -e "behind") != "" ]]; then
  echo "Error: repo has unpushed commits. Bumping the version should not include other changes."
  exit 1
fi

# Check if user has push access
if ! git ls-remote --exit-code "$ORIGIN" >/dev/null 2>&1; then
  echo "Error: Cannot access remote repository. Ensure you have push permissions."
  exit 1
fi  


# Update changelog using make changes output only
make_changes_output=$(make changes 2>/dev/null | grep -v '^==> Merged PRs since last release$' | sed '/^$/d' || true)
if [ -n "$make_changes_output" ]; then
    {
        head -n 1 CHANGELOG.md
        echo ""
        echo "## [$new_version] - $(date '+%Y-%m-%d')"
        echo ""
        echo "$make_changes_output"
        tail -n +2 CHANGELOG.md
    } > CHANGELOG.md.tmp && mv CHANGELOG.md.tmp CHANGELOG.md
    echo "Changelog updated with make changes output."
fi

# Update version in godo.go
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    sed -i '' "s/libraryVersion = \"$current_version\"/libraryVersion = \"$new_version\"/" godo.go
else
    # Linux
    sed -i "s/libraryVersion = \"$current_version\"/libraryVersion = \"$new_version\"/" godo.go
fi

echo "Version updated from $current_version to $new_version in godo.go"

