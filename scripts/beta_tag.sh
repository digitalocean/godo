#!/usr/bin/env bash

set -eo pipefail

# beta_tag.sh
#
# Creates and pushes a pre-release beta tag for the current libraryVersion
# defined in godo.go. The base version is taken as-is from godo.go and is
# NOT modified by this script. Beta numbers auto-increment based on the
# existing git tags for that base version.
#
# Examples (assuming libraryVersion = "1.23.0" in godo.go):
#   - No existing v1.23.0-beta.* tags  -> creates v1.23.0-beta.1
#   - v1.23.0-beta.1 already exists    -> creates v1.23.0-beta.2
#   - v1.23.0-beta.3 already exists    -> creates v1.23.0-beta.4
#
# When the base version in godo.go is bumped to a new GA (e.g. 1.24.0)
# the beta counter naturally resets because there are no v1.24.0-beta.*
# tags yet.

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

# Get base version from godo.go (the current GA version)
base_version=$(grep -E '^\s*libraryVersion\s*=' godo.go | sed 's/.*"\(.*\)".*/\1/')
if [ -z "$base_version" ]; then
  echo "Error: Could not find libraryVersion in godo.go"
  exit 1
fi

# Make sure we know about every existing beta tag before deciding the next number
git fetch "$ORIGIN" --tags --quiet >/dev/null 2>&1 || true

prefix="v${base_version}-beta."

# Find the highest existing beta number for the current base version.
latest=$(git tag --list "${prefix}*" \
  | sed "s|^${prefix}||" \
  | grep -E '^[0-9]+$' \
  | sort -n \
  | tail -n 1 || true)

if [ -z "$latest" ]; then
  next=1
else
  next=$((latest + 1))
fi

new_tag="${prefix}${next}"

# Double-check the tag does not already exist locally or remotely
if git rev-parse -q --verify "refs/tags/${new_tag}" >/dev/null; then
  echo "Error: tag ${new_tag} already exists locally."
  exit 1
fi
if git ls-remote --exit-code --tags "$ORIGIN" "refs/tags/${new_tag}" >/dev/null 2>&1; then
  echo "Error: tag ${new_tag} already exists on ${ORIGIN}."
  exit 1
fi

message="Beta pre-release ${new_tag} (base version ${base_version})"

git tag -a "$new_tag" -m "$message" "$COMMIT"
git push "$ORIGIN" tag "$new_tag"

echo "Created and pushed beta tag: $new_tag"
echo "Base version (godo.go): $base_version"
