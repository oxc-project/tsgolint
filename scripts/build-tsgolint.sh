#!/usr/bin/env bash
#
# Cross-compile tsgolint for every supported platform into dist/.
#
# tsgolint is pure Go (CGO disabled), so a single machine can produce every
# target. The build links the typescript-go submodule via the go.work file, so
# the build tree must be initialized first:
#
#   just init        # fetch the submodule, apply patches/, copy collections
#   scripts/build-tsgolint.sh
#
# Output layout (consumed by release-tsgolint.sh):
#   dist/<os>-<arch>/tsgolint[.exe]
#
# Env:
#   TSGOLINT_DIST       output dir (default: dist)
#   TSGOLINT_PLATFORMS  space-separated <os>-<arch> list (default: all below)
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

BIN="tsgolint"
OUT="${TSGOLINT_DIST:-dist}"
DEFAULT_PLATFORMS="darwin-arm64 darwin-amd64 linux-amd64 linux-arm64 windows-amd64"
read -r -a PLATFORMS <<<"${TSGOLINT_PLATFORMS:-$DEFAULT_PLATFORMS}"

# The build needs the submodule populated and the patched collections copied.
# Mirror the guard against `just init` not having been run.
if [ ! -f typescript-go/go.mod ]; then
  echo "error: typescript-go submodule is not initialized. Run 'just init' first." >&2
  exit 1
fi
if [ ! -d internal/collections ] || [ -z "$(ls -A internal/collections 2>/dev/null)" ]; then
  echo "error: internal/collections is empty. Run 'just init' first." >&2
  exit 1
fi

rm -rf "$OUT"
for plat in "${PLATFORMS[@]}"; do
  os="${plat%-*}"
  arch="${plat#*-}"
  ext=""
  [ "$os" = "windows" ] && ext=".exe"
  dest="$OUT/$plat/$BIN$ext"
  mkdir -p "$(dirname "$dest")"
  echo ">> building $plat -> $dest"
  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
    go build -trimpath -ldflags="-s -w" -o "$dest" ./cmd/tsgolint
done

echo ">> built:"
find "$OUT" -type f | sort
