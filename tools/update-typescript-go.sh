#!/usr/bin/env -S bash -euxo pipefail

# Updates the typescript-go submodule, its dependent modules, and regenerates code.
# All changes are staged for a new commit.
# Use the "--no-build" flag to skip the final build verification.

BUILD=true
if [[ "${1:-}" == "--no-build" ]]; then
    BUILD=false
fi

pushd typescript-go
git switch main
git reset --hard origin/main
git pull --prune
git am --3way --no-gpg-sign ../patches/*.patch
popd

go work sync

find ./shim -type f -name 'go.mod' -execdir go get -u -x github.com/microsoft/typescript-go@latest \; -execdir go mod tidy -v \;
go mod tidy

go run ./tools/gen_shims

git add ./typescript-go ./shim ./go.mod ./go.sum

if [[ "$BUILD" == "true" ]]; then
    go build ./cmd/tsgolint
fi
