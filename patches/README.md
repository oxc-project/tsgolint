# Patches

Most patches adapt typescript-go for TSGolint or tune its performance.

`0006-fix-context-free-expression-signature-caching.patch` makes
`getContextFreeTypeOfExpression` recompute call signatures and contextual
callback types without reading or writing their normal caches. This keeps
assertion checks independent of whether semantic diagnostics already checked
the call and prevents a context-free query from changing later normal type
queries. See
[TSGolint issue #1141](https://github.com/oxc-project/tsgolint/issues/1141).

Module resolution caching is tracked [here](https://github.com/microsoft/typescript-go/issues/673).

TODO: propose upstreaming other patches

TODO: right now patches are created via `git format-patch` and applied via `git am`.
We should probably use something like [this](https://github.com/pulumi/ci-mgmt/blob/d98489a822ebd290978a238d54c1d32e4aaca208/provider-ci/internal/pkg/templates/base/scripts/upstream.sh) or [this](https://github.com/microsoft/go-infra/tree/9ac588cb17d2f3713c37efe33babe37f2f4d625f/cmd/git-go-patch).
