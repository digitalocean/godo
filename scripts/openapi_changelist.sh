#!/usr/bin/env bash
# Print merged PRs / commits on digitalocean/openapi between two short SHAs.
# Usage: scripts/openapi_changelist.sh <current_sha> <target_sha>
set -euo pipefail

current="${1:-}"
target="${2:-}"

if [[ -z "$current" || -z "$target" ]]; then
  echo "usage: $0 <current_sha> <target_sha>" >&2
  exit 1
fi

if [[ "$current" == "0000000" || "$current" == "$target" ]]; then
  echo "_No prior OpenAPI pin (or unchanged). Target: \`$target\`._"
  exit 0
fi

echo "Commits on digitalocean/openapi between \`$current\` and \`$target\`:"
echo ""
gh api \
  -H "Accept: application/vnd.github+json" \
  "/repos/digitalocean/openapi/compare/${current}...${target}" \
  --jq '.commits[] | "- \(.sha[0:7]) \(.commit.message | split("\n")[0])"' \
  2>/dev/null || echo "_Could not fetch compare (check token permissions)._"
