## How shims work?

For each internal typescript-go package that we want to use, shim module must be created.

### Shim Go module

Go module `module github.com/microsoft/typescript-go/shim/<package name>` must be initialized in the [shim/ directory](../../shim/).

Package name should be added to [`main.go`](./main.go) alongside the other packages.

To generate `shim.go` run:

```shell
go run ./tools/gen_shims
```

All public functions, types, constants and variables will be reexported in the generated shim via [`go:linkname` directive](https://pkg.go.dev/cmd/compile#hdr-Linkname_Directive).

Optionally `extra-shim.json` can be created in the shim module directory.
It allows to reexport additional private functions, type methods and struct fields.

### `replace` for shim package

[`replace`](https://go.dev/ref/mod#go-mod-file-replace) directive must be added to the root [`go.mod`](../../go.mod).
