#!/usr/bin/env -S just --justfile

set windows-shell := ["powershell.exe", "-NoLogo", "-Command"]
set shell := ["bash", "-cu"]

ready:
  just fmt
  just lint
  just test

init:
  git submodule update --init
  pushd typescript-go && git am --3way --no-gpg-sign ../patches/*.patch && popd
  mkdir -p internal/collections && find ./typescript-go/internal/collections -type f ! -name '*_test.go' -exec cp {} internal/collections/ \;

build:
  GOEXPERIMENT=greenteagc go build -o tsgolint ./cmd/tsgolint

test: build
  ./test.sh
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
