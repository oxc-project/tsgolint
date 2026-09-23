# Node-count-based checker assignment

This experiment adds a fourth variant to the scheduling comparison:

1. **Baseline:** the original shared lint queue.
2. **Affinity:** checker-owned queues without stealing.
3. **Affinity + stealing:** owners process their own queue, then steal pending files.
4. **Affinity + stealing + node sort:** the same hybrid, with source files stably
   sorted by descending parser-recorded node count before round-robin checker
   assignment.

All four variants are rebuilt with the same benchmark harness, compiler base,
dependencies, and Go settings. Only the fourth includes the compiler change.
These results are independent of the earlier two-way experiments.

## Results

All values below are medians of ten fresh-process samples per variant, measured
in this experiment. Bold marks the lowest observed median, not statistical
significance. These are not the earlier historical timings.

### Lint only

Seconds; lower is faster.

| Repository |  Baseline | Affinity | Affinity + stealing | Affinity + stealing + node sort |
| ---------- | --------: | -------: | ------------------: | ------------------------------: |
| TypeORM    |     1.028 |    1.087 |           **1.020** |                           1.051 |
| TypeScript | **2.355** |    3.099 |               2.382 |                           2.472 |
| VS Code    |     5.802 |    6.024 |               5.447 |                       **5.277** |
| Vue        |     1.032 |    1.054 |               1.005 |                       **0.996** |

### Type checking + linting

Seconds; lower is faster.

| Repository | Baseline | Affinity | Affinity + stealing | Affinity + stealing + node sort |
| ---------- | -------: | -------: | ------------------: | ------------------------------: |
| TypeORM    |    1.376 |    1.172 |           **1.156** |                           1.166 |
| TypeScript |    3.326 |    3.576 |           **3.005** |                           3.029 |
| VS Code    |    8.173 |    6.596 |           **6.227** |                           6.307 |
| Vue        |    1.484 |    1.135 |           **1.131** |                           1.198 |

### Effect of adding node sorting

Relative to affinity + stealing. Positive values mean sorting is slower or
allocates more. CV is the timing standard deviation divided by the mean, shown
for stealing / sorted; it is not a confidence interval.

| Repository | Case         | Time delta | Allocated bytes delta |   Timing CV |
| ---------- | ------------ | ---------: | --------------------: | ----------: |
| TypeORM    | Lint         |      +3.0% |                 -0.0% | 5.9% / 6.7% |
| TypeORM    | Check + lint |      +0.9% |                 +0.3% | 3.4% / 4.5% |
| TypeScript | Lint         |      +3.8% |                 -0.9% | 2.0% / 3.7% |
| TypeScript | Check + lint |      +0.8% |                 +0.4% | 7.7% / 2.9% |
| VS Code    | Lint         |      -3.1% |                 -0.2% | 7.2% / 6.6% |
| VS Code    | Check + lint |      +1.3% |                 -0.7% | 2.4% / 6.7% |
| Vue        | Lint         |      -0.9% |                 -0.7% | 7.9% / 8.4% |
| Vue        | Check + lint |      +5.9% |                 +0.6% | 5.4% / 5.9% |

Node sorting does not show a consistent runtime benefit. Checking + linting is
0.8–5.9% slower in all four repositories, while lint-only results range from 3.1%
faster to 3.8% slower. Many differences are small relative to observed variation;
these data do not establish a reliable win for the extra heuristic. Allocated
bytes change by less than 1% in every case.

See the [complete raw timing summary](node-sort-results/full/matrix.md) and
[per-variant distributions](node-sort-results/full/summary.json).

## Initial balance and stealing

One separate instrumented full-lint run per repository and variant. Each arrow
compares affinity + stealing with affinity + stealing + node sort. A perfectly
even share across four checkers is 25%. Assigned nodes measure the initial owner;
rule time measures actual execution after stealing. These profiles are excluded
from the timing matrix.

| Repository | Largest selected-node share | Largest whole-program node share | Largest rule-time share | Stolen files |
| ---------- | --------------------------: | -------------------------------: | ----------------------: | -----------: |
| TypeORM    |               29.7% → 27.4% |                    29.1% → 27.4% |           25.4% → 25.5% |     206 → 83 |
| TypeScript |               35.6% → 38.9% |                    32.2% → 33.2% |           25.1% → 25.5% |    175 → 148 |
| VS Code    |               27.3% → 25.3% |                    27.0% → 25.9% |           25.2% → 25.2% |     125 → 55 |
| Vue        |               31.4% → 27.9% |                    30.1% → 30.4% |           25.4% → 25.8% |      38 → 24 |

Sorting improves selected-file node balance in three repositories, but worsens it
for TypeScript. Descending round robin does not guarantee smaller differences in
the accumulated node totals, and declaration-file assignments can make the full
program differ from the lint selection. Stealing already leaves rule wall time
close to 25% per checker in both variants. This is consistent with limited room
for this initial-assignment heuristic to improve total runtime; it is not proof
that equal node counts imply equal type-checking cost.

Fewer files are stolen in these profiles, but the timing matrix shows no consistent
speedup and less than 1% change in allocated bytes. Profile shares come from single
instrumented runs, not repeated timing measurements.

[Raw profiles](node-sort-results/profiles/profiles.json).

## Implementation

`checkerPool.createCheckers` sorts a copy of `program.files`, preserving the
original order for equal node counts. Sorting includes declaration files and
dependencies because those files also belong to the compiler's checker pool.
Program order and file traversal order remain unchanged. A single-checker pool
skips the copy and sort. No additional AST traversal is needed: the parser already
records `SourceFile.NodeCount`.

This is descending round robin, not assignment to the checker with the smallest
accumulated node count. A large individual file can still dominate one checker.
Node count estimates source size, not the difficulty of type operations, and the
selected lint files can have a different balance from the full program.

The permanent change and its regression test are in
[`0006-perf-distribute-checker-files-by-descending-node-count.patch`](../../patches/0006-perf-distribute-checker-files-by-descending-node-count.patch).
The submodule's existing patched HEAD is unchanged; its working tree has the patch
applied for development. No submodule pointer update is needed.

## Method

The full-project matrix uses TypeScript, VS Code, TypeORM, and Vue. Each variant
runs ten fresh-process samples of full lint and semantic checking plus lint,
with four compiler checkers, four configured lint workers, and GOMAXPROCS=4.
Five-sample batches run in A/B/C/D/D/C/B/A order for each case, giving every
variant mirrored positions in the sequence. The timed benchmark code is unchanged.

Both cases include program creation and binding. The node-count sort is included
in the measured operation. OS filesystem caches are not flushed. These are Go
linter API measurements, not Oxlint CLI or incremental LSP measurements.

The machine-load monitor samples every five seconds, waits for 30 seconds without
build jobs and at most 400% aggregate external CPU, and interrupts sustained
contention. Samples overlapping the incident, including a five-second lookback,
are quarantined; only a validated ordered prefix is resumed. Idle `go tool pprof`
HTTP viewers are not build jobs, but their CPU remains part of the load total.
The threshold was raised from 200% at the user’s request before any timing
samples started. After repeated build interruptions, the remaining TypeORM and Vue cases were
prioritized before resuming VS Code; every case retained its original variant
order. No samples are removed because of their measured performance. Short spikes and
thermal drift can still escape this shared-desktop protocol.

## Validation

All internal Go tests, CLI tests, compiler package tests, targeted race tests
(three repetitions), and E2E tests (22 passed, one skipped) pass with sorting.
The assignment regression test fails against the unsorted compiler for multiple
checkers and passes for the sorted compiler. It checks stable ties, unchanged
program order, one checker, and a requested checker count greater than the file
count. Two-way and four-way runner ordering and partial-prefix resumption were
also checked.

Separate diagnostic probes compare the hybrid with and without sorting, including
fixes and suggestions. All ordinary fixture runs match at one and four workers,
with and without semantic checking. The known #1141 reproduction still changes
diagnostics nondeterministically with four workers after semantic checking;
sorting does not fix checker-state reuse.

The existing checker-state issue [#1141](https://github.com/oxc-project/tsgolint/issues/1141)
still limits adoption of checker reuse. Passing tests and benchmark diagnostic
counts do not establish semantic equivalence.

## Reproduction and raw data

[`node-sort-results/source.json`](node-sort-results/source.json) records source
hashes, binary hashes, variant labels, and check results. The directory also
contains source snapshots for reconstructing the unsorted variants, binary build
information, raw measurements, load logs, and validation output.

Build each variant in a separate checkout or use Go's `-overlay` support to replace
the compiler and linter sources. Keep the benchmark harness identical in all four.
Then run the matrix with the named-variant interface:

```sh
python3 benchmarks/checker-affinity/run.py \
  --variant baseline=/path/to/baseline.test \
  --variant affinity=/path/to/affinity.test \
  --variant stealing=/path/to/stealing.test \
  --variant sorted=/path/to/sorted.test \
  --projects /path/to/prepared-projects \
  --output /path/to/new-results \
  --phase measure --workers 4 \
  --names typescript vscode typeorm vue \
  --scenarios full-lint full-check --samples 10

python3 benchmarks/checker-affinity/summarize.py /path/to/new-results
```
