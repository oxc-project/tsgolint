#!/usr/bin/env -S bash -euxo pipefail

pushd typescript
TSGO_COMMIT="$(git rev-parse HEAD)"
git am --3way --no-gpg-sign ../patches/*.patch
popd

find ./shim -type f -name 'go.mod' -execdir go get -x "github.com/microsoft/TypeScript/tsc@$TSGO_COMMIT" \; -execdir go mod tidy -v \;
go mod tidy

go run ./tools/gen_shims

git add ./shim ./go.mod ./go.sum
