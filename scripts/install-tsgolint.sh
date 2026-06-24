#!/usr/bin/env bash
#
# Install a prebuilt tsgolint for this OS/arch from the fork's GitHub releases.
#
# Unlike a normal CLI, oxlint does NOT find tsgolint on PATH: it uses the
# OXLINT_TSGOLINT_PATH env var, and otherwise the binary bundled inside the
# oxlint-tsgolint npm package. So this installer drops the binary on disk and
# prints the OXLINT_TSGOLINT_PATH you should export (point oxlint --type-aware
# at this fork's binary).
#
# Usage:
#   scripts/install-tsgolint.sh [VERSION]      # VERSION defaults to latest
#   curl -fsSL <raw-url>/install-tsgolint.sh | bash
#   curl -fsSL <raw-url>/install-tsgolint.sh | bash -s -- v2026.06.24
#
# Works without auth on the public fork (curl fallback); uses gh when present.
#
# Env:
#   TSGOLINT_REPO     source repo (default: robinnagpal-newsela/oxlint-tsgolint)
#   TSGOLINT_BINDIR   install dir (default: $HOME/.local/bin)
set -euo pipefail

REPO="${TSGOLINT_REPO:-robinnagpal-newsela/oxlint-tsgolint}"
BIN="tsgolint"
BINDIR="${TSGOLINT_BINDIR:-$HOME/.local/bin}"
VERSION="${1:-}"

case "$(uname -s)" in
  Linux)  OS=linux ;;
  Darwin) OS=darwin ;;
  *) echo "unsupported OS: $(uname -s)" >&2; exit 1 ;;
esac
case "$(uname -m)" in
  x86_64|amd64)  ARCH=amd64 ;;
  arm64|aarch64) ARCH=arm64 ;;
  *) echo "unsupported arch: $(uname -m)" >&2; exit 1 ;;
esac
ASSET="$BIN-$OS-$ARCH"

mkdir -p "$BINDIR"
DEST="$BINDIR/$BIN"
TMP="$(mktemp)"
trap 'rm -f "$TMP"' EXIT

if command -v gh >/dev/null 2>&1; then
  echo ">> Downloading $ASSET (${VERSION:-latest}) via gh..."
  # No tag arg => latest release.
  gh release download ${VERSION:+"$VERSION"} --repo "$REPO" --pattern "$ASSET" --output "$TMP" --clobber
else
  if [ -n "$VERSION" ]; then
    URL="https://github.com/$REPO/releases/download/$VERSION/$ASSET"
  else
    URL="https://github.com/$REPO/releases/latest/download/$ASSET"
  fi
  echo ">> Downloading $URL via curl..."
  curl -fsSL "$URL" -o "$TMP"
fi

chmod +x "$TMP"
mv -f "$TMP" "$DEST"
trap - EXIT
echo ">> Installed $BIN -> $DEST"

cat >&2 <<EOF
>> Point oxlint at this binary by exporting:

     export OXLINT_TSGOLINT_PATH="$DEST"

   Add that to your shell profile (or your repo's lint script) so
   \`oxlint --type-aware\` uses this fork instead of the bundled tsgolint.
EOF
