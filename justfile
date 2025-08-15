init:
  git submodule update --init
  pushd typescript-go && git am --3way --no-gpg-sign ../patches/*.patch && popd

build:
  go build -o tsgolint ./cmd/tsgolint

test:
  ./test.sh
  go test ./internal/...
  ./test-snapshot.sh

lint:
  golangci-lint run
