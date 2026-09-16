# Shim Generator

Shims expose internal typescript-go APIs to tsgolint. Each internal package used by tsgolint has a corresponding Go module under [`shim/`](../../shim/).

## Shim Go Module

To expose another package, create a module named `github.com/microsoft/typescript-go/shim/<package>` in `shim/<package>/`, following an existing module's `go.mod`.

Add the package name to `packagesToShim` in [`main.go`](./main.go).

After initializing the project with `just init`, generate the `shim.go` files from the repository root:

```shell
just shim
```

The generator exposes supported exported declarations using type aliases, generated declarations, and `go:linkname` directives. Do not edit generated `shim.go` files directly.

An optional `extra-shim.json` in each shim module configures additional private functions, methods, and struct fields to expose, as well as functions to ignore. After changing this configuration, run `just shim` and commit both the configuration and generated files.

## Module Replacement

Add a `replace` directive to the root [`go.mod`](../../go.mod) pointing the shim module to its local directory, alongside the existing shim replacements. Add a `require` entry when the module is used.
