#!/usr/bin/env -S just --justfile

set windows-shell := ["powershell.exe", "-NoLogo", "-Command"]
set shell := ["bash", "-cu"]

ready:
  just fmt
  just lint
  just test

[unix]
init:
  git submodule update --init
  pushd typescript-go && git am --3way --no-gpg-sign ../patches/*.patch && popd
  mkdir -p internal/collections && find ./typescript-go/internal/collections -type f ! -name '*_test.go' -exec cp {} internal/collections/ \;

[windows]
init:
  git submodule update --init
  pushd typescript-go; Get-ChildItem ../patches/*.patch | ForEach-Object { git am --3way --no-gpg-sign $_.FullName }; popd
  New-Item -ItemType Directory -Force -Path internal\collections
  Get-ChildItem -Path .\typescript-go\internal\collections\* -File | Where-Object { $_.Name -notlike '*_test.go' } | ForEach-Object { Copy-Item $_.FullName -Destination .\internal\collections\ }

[unix]
build:
  GOEXPERIMENT=greenteagc go build -o tsgolint ./cmd/tsgolint

[windows]
build:
  $env:GOEXPERIMENT="greenteagc"; $env:GOOS="windows"; $env:GOARCH="amd64"; go build -o tsgolint.exe ./cmd/tsgolint

test: build
  cd e2e && pnpm run test --run && cd ..
  go test ./internal/...

lint:
  golangci-lint run

fmt:
  gofmt -w internal cmd tools

shim:
  ./tools/gen-npm-packages.mjs

pull:
  pushd typescript-go && git reset --hard origin/main
  git pull
  just init
