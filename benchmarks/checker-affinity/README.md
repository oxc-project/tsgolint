# Checker-affinity experiment

**To benchmark all four variants on another machine or your own repository,
follow [RUN-LOCAL.md](RUN-LOCAL.md).** It includes one-command initialization/build,
local tsconfig selection, controls, and a table highlighting the fastest medians.

This POC evaluates [#1175](https://github.com/oxc-project/tsgolint/issues/1175):
give each compiler-owned checker its assigned files, then let idle checkers steal
pending files. It preserves the existing lint worker pool. The original
[pure-affinity experiment](REPORT.md) predates work stealing; its results do not
describe the current hybrid implementation.

The [node-count follow-up](NODE-SORT.md) compares four variants in one fresh
matrix: the shared baseline, pure affinity, affinity with stealing, and affinity
with stealing plus node-count-based round-robin assignment.

See the [work-stealing follow-up](WORK-STEALING.md) and the
[original affinity report](REPORT.md) before considering adoption. In particular,
reusing semantic-checker state can expose the existing correctness bug in
[#1141](https://github.com/oxc-project/tsgolint/issues/1141) on more files.

## Prepare the comparison

Initialize the repository normally, including its pinned typescript-go submodule,
patches, and copied collections. Use Go 1.27.1 for the recorded experiment.

Build two test binaries with the same `checker_affinity_bench_test.go`: one with
the original `internal/linter/linter.go` from baseline commit
`acd88e1b1ef8f219e8956f804f2548dafd949705`, and one with the POC. Use separate
checkout directories, the same patched compiler, and identical Go module inputs.
No production scheduling flag is needed.

```sh
# Run in the baseline checkout after copying the new benchmark harness into it.
go test -c -o /tmp/affinity-baseline.test ./cmd/tsgolint

# Run in the candidate checkout.
go test -c -o /tmp/affinity-candidate.test ./cmd/tsgolint

# Use a disposable directory: this downloads projects, installs dependencies
# with lifecycle scripts disabled, and adjusts configs for TS7.
python3 benchmarks/checker-affinity/prepare.py /tmp/affinity-projects
```

The preparation script verifies exact project commits and keeps the existing
lockfiles. It uses npm 8.19.4 for TypeORM's historical optional-dependency lock
entries and pnpm 10.10.0 for Sentry. Current defaults prepare TypeScript, VS Code,
TypeORM, and Sentry; use `--names typescript vscode typeorm vue` to reproduce the
historical corpus (Vue uses pnpm 9.12.3). Sentry timings were paused and are not in
the committed reports. Node/npm version managers can select incompatible
versions from these old checkouts. Use `--node /absolute/path/to/node` and
`--npm-cli /absolute/path/to/npm-cli.js` to pin the executables;
`--legacy-node` optionally selects an older Node executable for TypeORM.
The recorded run used Node 26.5.0/npm 11.17.0, with Node 22.14.0/npm 8.19.4 for
TypeORM. `preparation.json` records revisions, lock hashes, and config patches.

`--generated-only` prepares just the synthetic workload and #1141 reproduction.
`--skip-install` reapplies configuration adjustments without installing packages.

## Run

```sh
python3 benchmarks/checker-affinity/run.py \
  --baseline /tmp/affinity-baseline.test \
  --candidate /tmp/affinity-candidate.test \
  --projects /tmp/affinity-projects \
  --output /tmp/affinity-results

python3 benchmarks/checker-affinity/summarize.py /tmp/affinity-results
```

Run on an otherwise idle machine. `--phase validate`, `--phase measure`, and
`--phase profile` separate correctness, uninstrumented timing, and profiling.
Do not run these phases concurrently. A measurement run can resume from complete
cases in `samples.csv`; an interrupted case restarts its full ABBA sequence.
`--resume-partial` instead continues a validated ABBA prefix. Use it when an
external load monitor has removed the affected suffix, retaining only samples
that finished before contention. Pauses preserve the original version/batch order.
Use a fresh output directory after changing binaries, dependencies, or manifests.
New samples also record their start timestamp and whole-process elapsed time so
that external load incidents can be matched to the affected runs. Process elapsed
time includes untimed setup and is not the reported benchmark duration.

Defaults are ten samples per implementation in five-sample A/B/B/A batches,
ten correctness repetitions, and worker/GOMAXPROCS settings 4, 1, and host default.
`--names synthetic` limits the corpus; `--workers 4` limits the concurrency matrix.
`--scenarios full-lint full-check` limits the measured operations for focused
scheduler experiments. Use fresh output directories for each comparison and
record what the runner's `baseline` and `candidate` labels mean.
For a comparison with more than two variants, repeat `--variant NAME=PATH`
instead of passing `--baseline` and `--candidate`, with `--phase measure` or
`--phase profile`. Versions run in forward then reverse order: four variants use
five-sample A/B/C/D/D/C/B/A batches for ten samples per version. `summarize.py`
also accepts these named variants. Partial resume checks the corresponding
ordered prefix.
All measured programs use the normal checker configuration, independently of the
worker setting. The pinned compiler normally creates four checkers, even with one
worker; each measurement records the checker count, configured worker limit, and
GOMAXPROCS. The hybrid starts at most one worker per selected file or checker.

## Measurement boundaries

Each sample executes exactly one operation in a fresh test-binary process using
`-test.run=^$ -test.bench=^BenchmarkCheckerAffinity$ -test.benchtime=1x`.
Fresh programs in the same process would still share the global parsed-source
cache, so repeated in-process iterations are deliberately rejected.

- **full-lint / full-check:** Create and bind a fresh program, then lint all
  eligible files, with semantic diagnostics disabled/enabled respectively.
- **one-cold / ten-cold:** Create and bind a fresh full program, then lint only
  the selected file or 10% subset.
- **one-warm / ten-warm:** Create and bind a fresh program and semantically check
  all eligible files outside the measurement, then time its first filtered lint
  pass. No earlier lint pass has populated additional checker caches.

Program creation, binding, selection lookup, and filesystem reads are included
in cold measurements. Warm preparation is excluded from both timing and allocated
bytes. Process startup, configuration-file loading by the harness, and rule-list
construction are excluded. OS filesystem caches are not flushed. These are Go
linter API measurements, not Oxlint CLI or incremental LSP measurements.

Real-project manifests contain project-owned non-declaration source files outside
node_modules; imported JSON is excluded. TSX and configured JavaScript are also
supported. The original recorded corpus contains only `.ts` files. Relative-path SHA-256
ranking chooses the one-file and 10% subsets, while retaining manifest order.
Every implementation uses the same saved paths. Synthetic modules contain 64
uniquely named copies of the issue's declaration pattern and run only
`no-unsafe-assignment`; real projects run all registered rules with default options.

Timing runs count diagnostics but do not serialize them or generate fixes or
suggestions. Separate correctness probes enable fixes and suggestions and compare
complete normalized diagnostic records. A performance result is not a correctness
result. Allocated bytes are not peak memory usage.

Profiles wrap rule setup and listener calls only, measuring rule time per checker,
selected file counts, and active checkers. The slowest checker's share of summed
rule time helps identify imbalance; it is not total workload wall time or CPU
utilization. Instrumentation overhead excludes these runs from performance tables.
The extended profiler also records stolen-file counts and rule time per file;
older saved profiles do not contain these fields.

## Hybrid scheduler

Each compiler-owned checker has a prefilled, closed file queue. Its worker consumes
that queue first. Initial checkers are reserved before launching workers, so a
fast worker cannot consume several empty workloads and leave its peers idle.
The active worker count is also capped by the number of selected files, keeping
a lone file on its warmed owner instead of stealing it without any parallel benefit.
After draining its queue, the worker claims any unstarted checker
workloads before stealing, preserving affinity when only one worker is available.
Once all workloads have been claimed, an idle worker takes one pending file at a
time from the queue with the most remaining files. Empty owner queues also provide
checkers that can help with a concentrated filtered selection.

Only files move between queues. Each checker remains exclusive to one workload,
so thieves use their own mutable caches. This can repeat type work for stolen
files; queue length is only a cheap estimate of pending cost. Files already being
processed cannot be stolen or split. The scheduler does not change the preceding
semantic-checking phase, and it does not fix the existing checker-state bug.

The Go helpers are opt-in via `TSGOLINT_AFFINITY_CASE`; ordinary test runs skip
the external-project probes. The Python runner supplies case JSON, launches each
process, saves selected paths and raw samples, and records complete diagnostic
output as compressed JSON. It deliberately records #1141 failures instead of
turning them into accepted snapshots.
