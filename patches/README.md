# Patches

These patches adapt typescript-go for TSGolint and tune its performance.

`0006-perf-distribute-checker-files-by-descending-node-count.patch` is an
experimental checker-pool scheduling change. It stably sorts a copy of the source
files by descending parser-recorded node count before round-robin assignment.
Program order and traversal order remain unchanged, and single-checker pools skip
the sort. The patch includes assignment, tie-order, and program-order tests. See
the [node-count benchmark](../benchmarks/checker-affinity/NODE-SORT.md) for the
four-way comparison and correctness limitations.

Module resolution caching is tracked [here](https://github.com/microsoft/typescript-go/issues/673).

TODO: propose upstreaming other patches

TODO: right now patches are created via `git format-patch` and applied via `git am`.
We should probably use something like [this](https://github.com/pulumi/ci-mgmt/blob/d98489a822ebd290978a238d54c1d32e4aaca208/provider-ci/internal/pkg/templates/base/scripts/upstream.sh) or [this](https://github.com/microsoft/go-infra/tree/9ac588cb17d2f3713c37efe33babe37f2f4d625f/cmd/git-go-patch).
