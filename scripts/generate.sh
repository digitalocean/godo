#!/usr/bin/env bash
# Bundle the full DigitalOcean OpenAPI spec (or use a pre-bundled SPEC_FILE),
# then regenerate package v2.
#
# Usage:
#   make generate
#   OPENAPI=/path/to/digitalocean/openapi make generate
#   SPEC_FILE=/path/to/DigitalOcean-public.v2.yaml make generate
#
# Defaults OPENAPI to ../openapi (sibling checkout of github.com/digitalocean/openapi).
# When SPEC_FILE is set, bundling is skipped and that YAML is used as-is (CI path).
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

OPENAPI="${OPENAPI:-../openapi}"
SPEC_ENTRY="${SPEC_ENTRY:-specification/DigitalOcean-public.v2.yaml}"
BUNDLED="v2/gen/openapi/openapi.bundled.yaml"

mkdir -p v2/gen/openapi

if [[ -n "${SPEC_FILE:-}" ]]; then
  if [[ ! -f "$SPEC_FILE" ]]; then
    echo "SPEC_FILE not found: $SPEC_FILE" >&2
    exit 1
  fi
  echo "==> using pre-bundled OpenAPI at $SPEC_FILE"
  # Resolve paths so SPEC_FILE can already be the bundle destination.
  src="$(cd "$(dirname "$SPEC_FILE")" && pwd)/$(basename "$SPEC_FILE")"
  dst="$ROOT_DIR/$BUNDLED"
  if [[ "$src" != "$dst" ]]; then
    cp "$src" "$dst"
  fi
else
  if [[ ! -f "$OPENAPI/$SPEC_ENTRY" ]]; then
    echo "Full OpenAPI spec not found at $OPENAPI/$SPEC_ENTRY" >&2
    echo "Clone https://github.com/digitalocean/openapi and set OPENAPI=/path/to/openapi" >&2
    echo "Or pass SPEC_FILE=/path/to/bundled.yaml" >&2
    exit 1
  fi
  echo "==> bundling full OpenAPI from $OPENAPI"
  npx -y @redocly/cli@1.25.13 bundle \
    "$OPENAPI/$SPEC_ENTRY" \
    --output "$ROOT_DIR/$BUNDLED" \
    --ext yaml
fi

echo "    wrote $BUNDLED ($(wc -c < "$BUNDLED") bytes)"

echo "==> generating package v2 from full bundled spec"
# Remove hand-written stub so generated client.gen.go owns Client/New.
rm -f v2/client.go
(
  cd tools/openapi-gen
  go run . -root "$ROOT_DIR" -config v2/gen/godo-gen.yaml
)

gofmt -w v2/*.gen.go
echo "==> done"
echo "    client.V2.<Service>.List/Get(...)"
