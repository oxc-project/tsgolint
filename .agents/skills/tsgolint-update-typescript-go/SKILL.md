---
name: tsgolint-update-typescript-go
description: "Update the `typescript-go` submodule in the `oxc-project/tsgolint` repository, refresh the local `patches/*.patch` stack against the new upstream base, regenerate shims, open a PR, and watch CI to completion. Use when working on routine `typescript-go` bump PRs for TSGolint or when repeating the same workflow as prior submodule update PRs."
---

# Update typescript-go

Use this workflow for the standard TSGolint `typescript-go` refresh.

## Workflow

1. Inspect the last successful submodule-update PR in the repo history.

Use it to mirror the expected shape of the change:
- advance the root `typescript-go` gitlink to a real upstream `microsoft/typescript-go` commit
- rebase the local `patches/*.patch` series onto that upstream base
- regenerate shims if the refreshed patch stack changes the exported surface
- make any small TSGolint compatibility fix required by upstream API drift
- open a PR titled `chore: update typescript-go submodule`
- wait for CI to finish and report the result

2. Fetch the latest `typescript-go` upstream state.

Check the current root gitlink and the current checkout inside `typescript-go/` separately.
The submodule checkout may temporarily sit on a patched local branch while you refresh the patch series, but the final root gitlink committed in TSGolint should point at the upstream base commit, not at a local patched commit.

3. Rebuild the patch stack on top of the new upstream base.

Inside `typescript-go/`:
- create a temporary branch from the target upstream commit
- replay `../patches/*.patch` with `git am --3way --no-gpg-sign`
- resolve drift in the submodule if upstream moved APIs or signatures
- continue until the full patch stack applies cleanly

If a patch fails, treat that as the main job. Update the patch content so it still expresses the same local intent on the new upstream files.

4. Regenerate repo artifacts from the refreshed patched submodule state.

From the TSGolint root:
- run `just shim` or `go run ./tools/gen_shims` when shim output changes
- update root call sites that depend on upstream signature changes
- prefer the smallest compatibility edit needed

5. Re-export the patch series from the refreshed submodule branch.

Use `git format-patch --no-signature <upstream-base>..HEAD` from inside `typescript-go/`, then replace the root `patches/*.patch` files with the regenerated series.
Keep the existing numbering and filenames when possible.

6. Verify using the same model as CI.

Important: direct local `go test ./...` against the bare upstream gitlink can fail because the root repo expects the patch series to be applied first.
CI behavior matters more:
- PR workflows check out submodules
- `.github/actions/setup/action.yml` applies `patches/*.patch`
- then build, test, and lint run against the patched submodule state

Locally, verify against the patched state before resetting the gitlink back to the upstream commit.
At minimum run the same Go surface used by CI when practical.

7. Prepare the final root repo state.

Before committing in TSGolint:
- set the root `typescript-go` gitlink to the upstream commit being adopted
- keep the refreshed `patches/*.patch`
- keep regenerated shim files and any compatibility fix in TSGolint
- do not commit unrelated local workspace changes

8. Create and monitor the PR.

Push a branch on `origin`, open a PR against `main`, and watch the checks until they finish.
For this repo, the relevant PR checks usually include:
- `test-go`
- `test-e2e`
- `lint`
- `Typos`
- `autofix`
- `check-schemas`
- `test-windows` may be skipped on PRs

If GitHub exposes Actions runs instead of classic statuses, query the check runs for the PR head SHA and wait until all required checks are completed.

## Guardrails

- Never commit a root gitlink pointing at a local-only patched `typescript-go` commit.
- Never discard unrelated root worktree changes unless the user explicitly asks.
- Treat the previous successful `chore: update typescript-go submodule` PR as the template for branch name, commit message, PR title, and expected diff shape.
- If CI is red, inspect the failing job before changing anything. Prefer a minimal follow-up fix over speculative refactors.
