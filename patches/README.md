# Patches

These patches do not change the behavior of typescript-go.
The main purpose of the patches is to tune tsgo performance a bit.

Module resolution caching is tracked [here](https://github.com/microsoft/typescript-go/issues/673).

TODO: propose upstreaming other patches

TODO: right now patches are created via `git format-patch` and applied via `git am`.
We should probably use something like [this](https://github.com/pulumi/ci-mgmt/blob/d98489a822ebd290978a238d54c1d32e4aaca208/provider-ci/internal/pkg/templates/base/scripts/upstream.sh) or [this](https://github.com/microsoft/go-infra/tree/9ac588cb17d2f3713c37efe33babe37f2f4d625f/cmd/git-go-patch).

## Patch notes

- `0006-fix-collections-ordered-map-public-json.patch`: updates `internal/collections/ordered_map.go` to use `github.com/go-json-experiment/json` APIs instead of `github.com/microsoft/typescript-go/internal/json`, so copied collections files compile in the parent module.
