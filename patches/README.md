# Patches

`0007-Fix-tests-for-TSGolint-patches.patch` updates snapshot filesystem tests to
use the project types exported by patch `0002`. Checker-pool constructor defaults
are tested separately from the session sizing override in patch `0001`. It also
updates `ReadFile` tests for patch `0005`, covering present, empty, and missing
files, content changes, cache clearing, and disabled caching. This keeps the
affected upstream test packages runnable after applying the patch stack. It
removes a redundant comment from patch `0003` without changing the config parser's
behavior.

`0006-perf-add-opt-in-node-count-checker-assignment.patch` enables descending
node-count round-robin assignment only when `OXLINT_TSGOLINT_CHECKER_SCHEDULING=sorted`.
It sorts a stable copy of the program files, including declarations, before
associating them with checkers. Program order is unchanged, and single-checker
pools skip sorting. All other values retain the original assignment. See the
[checker scheduling options](https://github.com/oxc-project/tsgolint/pull/1240).

These patches do not change the behavior of typescript-go.
The main purpose of the patches is to tune tsgo performance a bit.

Module resolution caching is tracked [here](https://github.com/microsoft/typescript-go/issues/673).

TODO: propose upstreaming other patches

TODO: right now patches are created via `git format-patch` and applied via `git am`.
We should probably use something like [this](https://github.com/pulumi/ci-mgmt/blob/d98489a822ebd290978a238d54c1d32e4aaca208/provider-ci/internal/pkg/templates/base/scripts/upstream.sh) or [this](https://github.com/microsoft/go-infra/tree/9ac588cb17d2f3713c37efe33babe37f2f4d625f/cmd/git-go-patch).
