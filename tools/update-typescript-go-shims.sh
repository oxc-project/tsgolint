#!/usr/bin/env -S bash -euxo pipefail

pushd typescript-go
TSGO_COMMIT="$(git rev-parse HEAD)"
git am --3way --no-gpg-sign ../patches/*.patch
popd

go work sync

find ./shim -type f -name 'go.mod' -execdir go get -x "github.com/microsoft/typescript-go@$TSGO_COMMIT" \; -execdir go mod tidy -v \;
go mod tidy

go run ./tools/gen_shims

git add ./shim ./go.mod ./go.sum
