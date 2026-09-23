# Run all four variants on your own repository

Use a separate checkout of this branch. Requirements: Git, Python 3.10+, and Go
1.26 or newer (the recorded benchmarks used Go 1.27.1). The commands below work
from the TSGolint checkout on macOS/Linux. No Node installation is needed to build
the benchmark tool; your target repository's dependencies must already be installed.

## 1. Clone and build

```sh
git clone --branch codex/checker-scheduling-benchmarks \
  https://github.com/oxc-project/tsgolint.git tsgolint-checker-bench
cd tsgolint-checker-bench
python3 benchmarks/checker-affinity/build.py --init
```

`--init` initializes the pinned compiler submodule, applies all six committed
patches, and copies the compiler collections. It does not run npm/pnpm or modify
your target project. It is a first-time operation: for subsequent builds, omit
`--init`. If it reports an already modified/patched submodule, use a fresh clone
or, if initialization already completed, run without `--init`.

All four binaries are built sequentially with the same harness, dependencies,
and Go settings. Go overlays select the committed sources in `variants/` without
editing checkout files. Do not move the binaries between operating systems or
CPU architectures; rebuild on the destination machine.

| Binary | Lint scheduling | Initial compiler assignment |
| --- | --- | --- |
| `baseline` | Original shared queue | Original round robin |
| `affinity` | Each checker processes its own queue | Original round robin |
| `stealing` | Owner queue first, then steal pending files | Original round robin |
| `sorted` | Owner queue first, then steal pending files | Stable descending node-count sort, then round robin |

The compiler sort uses a copy of all program files, including declarations;
it preserves the original program order. A running file cannot be split or stolen.
Sources are frozen snapshots of this experiment, so editing the production linter
does not silently change these comparisons. `local/bin/build-manifest.json`
records source/harness/binary hashes and the Go version; the runner verifies the
binary hashes. Go build information is saved beside each binary.

## 2. Run against a local project

Install the target repository's dependencies using its normal workflow first.
Choose a concrete tsconfig that includes source files, rather than a solution-only
config that contains just project references. Paths containing spaces can be quoted.

```sh
python3 benchmarks/checker-affinity/run.py \
  --binaries-dir benchmarks/checker-affinity/local/bin \
  --project-root /absolute/path/to/your/repo \
  --config tsconfig.json \
  --phase measure --workers 4 --samples 10 \
  --scenarios full-lint full-check \
  --output benchmarks/checker-affinity/local/work-repo

python3 benchmarks/checker-affinity/summarize.py \
  benchmarks/checker-affinity/local/work-repo
```

This runs 80 measured fresh processes: four variants × two scenarios × ten
samples. `matrix.md` gives a table with the **fastest median bolded** in each row;
`summary.json` includes allocation statistics, dispersion, and diagnostic counts.
The lowest median alone does not establish a statistically significant improvement.

For a quick setup check, use `--samples 2` and a separate output directory. For
controls, use `--workers 4 1 0` (0 uses host GOMAXPROCS). This changes the lint
worker limit and GOMAXPROCS, not the compiler's usual four-checker pool. Actual
counts are recorded. For the expanded six-scenario matrix, omit `--scenarios`.
For multiple configs/repos, run once per config with separate output directories
and optionally `--project-name my-label`.

Run on an otherwise idle machine after builds have finished, with consistent
power/thermal settings. The runner balances order using A/B/C/D/D/C/B/A batches
(five samples per batch at the default ten samples per variant). It does not
automatically detect background load. If another build starts, interrupt the run
and repeat into a new output directory once the machine is quiet. Historical
`monitor-script.py` files are archived evidence with machine-specific paths, not
portable launchers; the previous 400% background-CPU threshold is not imposed here.

Rerunning an identical command resumes complete cases and restarts any partial
case. Changed binary/settings or selected-source/config hashes are rejected.
Always use a new output directory after changing dependencies, extended configs,
declaration files, or other project inputs; those are not all fingerprinted.
`--timeout 1800` increases the per-process timeout from ten to thirty minutes.

## What the measurements mean

- **full-lint:** Create and bind a fresh program, then lint all selected files.
- **full-check:** The same, including semantic diagnostics before linting.
- **one-cold / ten-cold:** Build the full program, then lint one file / 10% of files.
- **one-warm / ten-warm:** Build and semantically check the full selection outside
  the timer, then measure the first lint of one file / 10% of files.

The harness selects program files inside `--project-root`, excluding declarations,
node_modules, and imported JSON. It supports TS/TSX and JavaScript enabled by the
config. All variants must produce identical manifests before measurements start.
`corpus.json` records selected file counts, physical line counts by extension,
and content/config hashes; physical lines include blank lines and comments.
`manifests.json` contains the exact selected relative paths and deterministic subsets.

These are Go API timings with **all registered lint rules at default options**;
your Oxlint/ESLint rule configuration is not loaded. Config loading and process
startup are outside the timer, cold program creation/binding is inside it, and
OS filesystem caches are not cleared. Existing TypeScript diagnostics are counted;
invalid configuration or an empty source selection stops the run. The harness
does not emit compiled JavaScript or modify the target repository.

Performance is the focus of this POC. It does not establish diagnostic parity:
semantic checker reuse can expose the existing [#1141 correctness issue](https://github.com/oxc-project/tsgolint/issues/1141).
Timing runs count diagnostics but do not serialize diagnostic contents.

## Optional workload profiles

After timing finishes, run the same command with `--phase profile` and a separate
output directory. This collects per-checker/file rule time, selected file counts,
and stolen-file counts for full and 10% selections, with/without semantic checking.
Profiles use four workers and include instrumentation overhead, so do not compare
their durations with uninstrumented timing samples. This phase does not run concurrently
with timing and does not require repeating the build.

## Results and prior evidence

`local/` is ignored by Git. Everything stays on your machine; nothing is uploaded.
Results can contain internal file paths, repository paths, and profile details.
To discuss timings without sharing source, start with `matrix.md`, `summary.json`,
and hardware/Go information from `environment-measure.json`; review their labels
and paths before sharing. Keep `samples.csv` for analyzing noise.

The committed [four-way report](NODE-SORT.md) and `node-sort-results/` preserve
the prior public-repository measurements. Older two-way results are explained in
[REPORT.md](REPORT.md) and [WORK-STEALING.md](WORK-STEALING.md). Historical paths and
binary hashes refer to that machine, not your new build. The current harness also
excludes imported JSON, added after those measurements. Sentry setup was prepared
but its timing run was paused; no Sentry performance result is claimed.
