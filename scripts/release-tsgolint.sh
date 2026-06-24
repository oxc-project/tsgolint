#!/usr/bin/env bash
#
# Build the cross-platform tsgolint binaries and publish them as a GitHub
# release on this fork. Designed for ad-hoc releases from a dev machine.
#
# Usage:
#   scripts/release-tsgolint.sh [VERSION]
#
# VERSION defaults to "v$(date +%Y.%m.%d)". If a release/tag with that name
# already exists, the short commit SHA is appended (e.g. v2026.06.24-1643b50)
# so same-day re-releases don't collide. Pass an explicit VERSION to override.
#
# Requirements: gh (authenticated), go, and one of sha256sum/shasum. The build
# tree must be initialized (`just init`) so the typescript-go submodule and
# patches/ are in place. The working tree must be clean so the uploaded
# binaries match the tagged commit.
set -euo pipefail

REPO="${TSGOLINT_REPO:-robinnagpal-newsela/oxlint-tsgolint}"
BIN="tsgolint"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

# 1. Refuse on a dirty tree: binaries are built from the working tree but the
#    release tag points at a commit, so they must agree. Submodule changes are
#    ignored (`just init` moves the submodule HEAD by applying patches/).
if [ -n "$(git status --porcelain --ignore-submodules=all)" ]; then
  echo "error: working tree is dirty; commit (and push) before releasing." >&2
  exit 1
fi

SHA="$(git rev-parse HEAD)"
SHORT="$(git rev-parse --short HEAD)"
BRANCH="$(git rev-parse --abbrev-ref HEAD)"

# 2. Choose the version.
VERSION="${1:-}"
if [ -z "$VERSION" ]; then
  VERSION="v$(date +%Y.%m.%d)"
  if gh release view "$VERSION" --repo "$REPO" >/dev/null 2>&1 \
     || git rev-parse -q --verify "refs/tags/$VERSION" >/dev/null; then
    VERSION="$VERSION-$SHORT"
  fi
fi
echo ">> Releasing $BIN $VERSION  (commit $SHORT on $BRANCH -> $REPO)"

# 3. The target commit must exist on the remote for gh to tag it.
git push origin "HEAD:$BRANCH"

# 4. Cross-compile every platform.
echo ">> Building (scripts/build-tsgolint.sh)..."
scripts/build-tsgolint.sh

# 5. Stage assets with platform-suffixed names + a checksum manifest.
STAGE="$(mktemp -d)"
trap 'rm -rf "$STAGE"' EXIT
for d in dist/*/; do
  plat="$(basename "$d")" # e.g. linux-amd64
  ext=""
  [ "${plat%-*}" = "windows" ] && ext=".exe"
  src="$d/$BIN$ext"
  [ -f "$src" ] || { echo "error: missing $src" >&2; exit 1; }
  cp "$src" "$STAGE/$BIN-$plat$ext"
done
sha256() { if command -v sha256sum >/dev/null 2>&1; then sha256sum "$@"; else shasum -a 256 "$@"; fi; }
( cd "$STAGE" && sha256 "$BIN"-* > SHA256SUMS )
echo ">> Assets:"; ls -1 "$STAGE"

# 6. Publish the release, tagging the exact commit that was built.
gh release create "$VERSION" \
  --repo "$REPO" \
  --target "$SHA" \
  --title "tsgolint $VERSION" \
  --notes "Prebuilt \`$BIN\` binaries from \`$BRANCH\` @ \`$SHORT\`.

Install the matching binary for this machine and point oxlint at it:
\`\`\`sh
curl -fsSL https://raw.githubusercontent.com/$REPO/$BRANCH/scripts/install-tsgolint.sh | bash
\`\`\`
Pin this version:
\`\`\`sh
curl -fsSL https://raw.githubusercontent.com/$REPO/$BRANCH/scripts/install-tsgolint.sh | bash -s -- $VERSION
\`\`\`
Then \`export OXLINT_TSGOLINT_PATH\` to the installed path (the installer prints it)." \
  "$STAGE/$BIN"-* "$STAGE/SHA256SUMS"

echo ">> Released: https://github.com/$REPO/releases/tag/$VERSION"
