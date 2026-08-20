# Patches

These patches adapt the native TypeScript compiler for TSGolint and tune tsgo performance.

Module resolution caching is tracked in the former staging repository [here](https://github.com/microsoft/typescript-go/issues/673).

TODO: propose upstreaming other patches

TODO: right now patches are created via `git format-patch` and applied via `git am`.
We should probably use something like [this](https://github.com/pulumi/ci-mgmt/blob/d98489a822ebd290978a238d54c1d32e4aaca208/provider-ci/internal/pkg/templates/base/scripts/upstream.sh) or [this](https://github.com/microsoft/go-infra/tree/9ac588cb17d2f3713c37efe33babe37f2f4d625f/cmd/git-go-patch).
