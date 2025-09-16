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

# Generate changelog for new version
last_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
if [ -n "$last_tag" ]; then
    # Generate changelog from last tag to HEAD
    if command -v github-changelog-generator >/dev/null 2>&1 && [ -n "${GITHUB_TOKEN:-}" ]; then
        github-changelog-generator \
            --user digitalocean \
            --project godo \
            --token "$GITHUB_TOKEN" \
            --since-tag "$last_tag" \
            --no-issues \
            --no-verbose \
            --output /tmp/new_changelog.md 2>/dev/null || {
            echo "Warning: Could not generate changelog automatically. You may need to update CHANGELOG.md manually."
        }
        
        # If changelog was generated, update CHANGELOG.md
        if [ -f /tmp/new_changelog.md ] && [ -s /tmp/new_changelog.md ]; then
            # Extract just the new entries (skip the header)
            new_entries=$(sed '1,/^## \[/d' /tmp/new_changelog.md | sed '/^\\\\*/,$d' | head -n -1)
            
            if [ -n "$new_entries" ]; then
                # Create backup
                cp CHANGELOG.md CHANGELOG.md.bak
                
                # Insert new entries after the header
                {
                    # Keep header and add new version entry
                    head -n 2 CHANGELOG.md
                    echo ""
                    echo "## [$new_version] - $(date '+%Y-%m-%d')"
                    echo ""
                    echo "$new_entries"
                    echo ""
                    # Keep rest of the file
                    tail -n +3 CHANGELOG.md
                } > CHANGELOG.md.tmp && mv CHANGELOG.md.tmp CHANGELOG.md
                
                echo "Changelog updated with new entries"
            else
                echo "No significant changes found for changelog"
            fi
            
            rm -f /tmp/new_changelog.md
        fi
    else
        if [ -z "${GITHUB_TOKEN:-}" ]; then
            echo "Warning: GITHUB_TOKEN not set, skipping automatic changelog generation"
        else
            echo "Warning: github-changelog-generator not found, skipping automatic changelog generation"
        fi
    fi
else
    echo "No previous tags found, skipping changelog generation"
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
