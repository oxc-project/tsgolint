ready:
  just fmt
  just lint
  just build
  just test

init:
  git submodule update --init
  pushd typescript-go && git am --3way --no-gpg-sign ../patches/*.patch && popd

build:
  go build -o tsgolint ./cmd/tsgolint

test:
  ./test.sh
  go test ./internal/...

lint:
  golangci-lint run

fmt:
  gofmt -w internal cmd tools

shim:
  ./tools/gen-npm-packages.mjs
