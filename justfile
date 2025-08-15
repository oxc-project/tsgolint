init:
  git submodule update --init

  cd typescript-go
  git am --3way --no-gpg-sign ../patches/*.patch
  cd ..

  just build

build:
  go build -o tsgolint ./cmd/tsgolint

test:
  ./test.sh
  go test ./internal/...

lint:
  golangci-lint run
