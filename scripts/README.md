# Release & install scripts

These scripts distribute prebuilt `tsgolint` binaries from this fork via GitHub
Releases, so consumers can use this fork's `tsgolint` without building Go.

`oxlint --type-aware` locates the tsgolint binary through the
`OXLINT_TSGOLINT_PATH` environment variable (falling back to the binary bundled
in the `oxlint-tsgolint` npm package). The install script drops this fork's
binary on disk and prints the `OXLINT_TSGOLINT_PATH` to export — no PATH entry
and no npm publish required.

## `build-tsgolint.sh` — cross-compile

Cross-compiles `tsgolint` for every supported platform into `dist/<os>-<arch>/`.
Pure Go (`CGO_ENABLED=0`), so one machine builds every target.

```sh
just init                  # one-time: fetch submodule, apply patches/, copy collections
scripts/build-tsgolint.sh
```

Override targets with `TSGOLINT_PLATFORMS` (e.g. `TSGOLINT_PLATFORMS="darwin-arm64 linux-amd64"`).

## `release-tsgolint.sh` — publish a GitHub release

Builds all platforms, stages them as `tsgolint-<os>-<arch>` assets with a
`SHA256SUMS` manifest, and runs `gh release create` against `TSGOLINT_REPO`
(default `robinnagpal-newsela/oxlint-tsgolint`).

```sh
scripts/release-tsgolint.sh            # version defaults to vYYYY.MM.DD
scripts/release-tsgolint.sh v2026.06.24
```

Requires `gh` (authenticated), `go`, a clean working tree, and an initialized
build tree (`just init`).

## `install-tsgolint.sh` — install on a dev machine / CI

Downloads the matching `tsgolint-<os>-<arch>` from the latest (or a pinned)
release into `$HOME/.local/bin` (override with `TSGOLINT_BINDIR`) and prints the
`OXLINT_TSGOLINT_PATH` to export.

```sh
# latest
curl -fsSL https://raw.githubusercontent.com/robinnagpal-newsela/oxlint-tsgolint/add-github-release-install-scripts/scripts/install-tsgolint.sh | bash
# pin a version
curl -fsSL .../install-tsgolint.sh | bash -s -- v2026.06.24

export OXLINT_TSGOLINT_PATH="$HOME/.local/bin/tsgolint"
oxlint --type-aware
```

## Yarn PnP support

This fork includes Yarn Plug'n'Play resolution support, applied to the
typescript-go submodule by `just init` via `patches/0006-Add-Yarn-PnP-support.patch`
(ported from upstream PR microsoft/typescript-go#1966) plus the `shim/pnp` and
`shim/vfs/pnpvfs` shims and the `PnpApi()` wiring in `internal/utils/host.go`.

A `.pnp.cjs` manifest is auto-detected from the working directory, so a binary
built and installed from this fork resolves a Yarn PnP workspace's modules and
`tsconfig extends` with no extra flags. Point oxlint at it with
`OXLINT_TSGOLINT_PATH` (see below).
