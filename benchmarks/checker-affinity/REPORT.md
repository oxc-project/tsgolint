# Checker-affinity POC: performance results

The current implementation adds work stealing. See the [follow-up report](WORK-STEALING.md) for that approach; the results below preserve the original pure-affinity experiment.

With four workers, affinity improves combined checking and linting on three real projects. It slows lint-only on all four and also regresses TypeScript’s combined run, so these results do not support unconditional adoption. The existing [#1141](https://github.com/oxc-project/tsgolint/issues/1141) correctness bug is also exposed more consistently; this remains an experimental POC.

The timing results below come from the fresh rerun requested after concerns about machine noise. The initial timing samples are excluded.

Measured on Apple M2 Max (12 logical CPUs), macOS 27.0, Go 1.27.1. Each comparison uses ten independent fresh-process samples per implementation in five-sample A/B/B/A batches. Baseline: `acd88e1b1ef8f219e8956f804f2548dafd949705`.

## Primary comparison: four workers

Workers and GOMAXPROCS are both four; the compiler pool also contains four checkers. Times are medians in seconds. Positive deltas mean the affinity candidate is slower.

| Project    | Shared lint | Affinity lint |  Delta | Shared check + lint | Affinity check + lint |  Delta |
| ---------- | ----------: | ------------: | -----: | ------------------: | --------------------: | -----: |
| vscode     |       5.246 |         5.599 |  +6.7% |               7.350 |                 5.988 | -18.5% |
| typescript |       2.383 |         3.133 | +31.5% |               3.155 |                 3.461 |  +9.7% |
| typeorm    |       1.021 |         1.088 |  +6.5% |               1.333 |                 1.162 | -12.8% |
| vue        |       0.968 |         1.049 |  +8.4% |               1.452 |                 1.095 | -24.6% |

Affinity improves combined checking and linting on three of four real projects. Lint-only regresses on all four. TypeScript also regresses with semantic checking enabled, so these results do not support adopting unconditional checker affinity.

## Allocations during combined runs

Allocated bytes per operation, not peak memory:

| Project    | Shared MiB | Affinity MiB |  Delta |
| ---------- | ---------: | -----------: | -----: |
| vscode     |     6063.0 |       5093.6 | -16.0% |
| typescript |     2225.2 |       1969.2 | -11.5% |
| typeorm    |     1168.4 |        990.8 | -15.2% |
| vue        |     1276.9 |        973.8 | -23.7% |

## Filtered files

Ten percent of files, chosen using relative-path hashes and shared manifests. Cold timings include creating and binding the full program; warm timings measure its first filtered lint after full semantic checking, excluding preparation. Medians in milliseconds.

| Project    | Cold shared → affinity |  Delta | Warm shared → affinity |  Delta |
| ---------- | ---------------------: | -----: | ---------------------: | -----: |
| vscode     |        1366.8 → 1376.4 |  +0.7% |          366.8 → 257.3 | -29.9% |
| typescript |          636.7 → 981.0 | +54.1% |          435.5 → 606.6 | +39.3% |
| typeorm    |          375.4 → 377.0 |  +0.4% |            54.9 → 45.3 | -17.6% |
| vue        |          259.0 → 246.9 |  -4.7% |           105.4 → 23.2 | -77.9% |

## Host-default control

The host default is 12 workers and GOMAXPROCS 12, while the compiler pool still has four checkers. Full-project medians in seconds.

| Project    | Shared lint | Affinity lint |  Delta | Shared check + lint | Affinity check + lint |  Delta |
| ---------- | ----------: | ------------: | -----: | ------------------: | --------------------: | -----: |
| vscode     |       5.082 |         5.159 |  +1.5% |               7.874 |                 5.650 | -28.2% |
| typescript |       2.011 |         2.882 | +43.3% |               2.846 |                 3.073 |  +8.0% |
| typeorm    |       0.851 |         0.894 |  +5.1% |               1.091 |                 0.947 | -13.1% |
| vue        |       0.838 |         0.915 |  +9.1% |               1.281 |                 1.025 | -20.0% |

The TypeScript regressions persist at host-default concurrency. VS Code’s lint-only control has coefficients of variation of 7.1% and 7.9%, exceeding its median difference; that small difference is inconclusive. The four-worker matrix remains the primary comparison.

## One-worker control

The compiler still creates four checkers. The shared queue lets one checker consume all lint files; affinity visits four checker groups serially. Full-project medians in seconds.

| Project    | Shared lint | Affinity lint |  Delta | Shared check + lint | Affinity check + lint |  Delta |
| ---------- | ----------: | ------------: | -----: | ------------------: | --------------------: | -----: |
| vscode     |      20.216 |        21.666 |  +7.2% |              33.178 |                25.001 | -24.6% |
| typescript |       7.792 |         8.068 |  +3.5% |              10.152 |                 8.968 | -11.7% |
| typeorm    |       2.933 |         3.042 |  +3.7% |               4.014 |                 3.321 | -17.3% |
| vue        |       2.855 |         3.180 | +11.4% |               4.483 |                 3.320 | -26.0% |

Combined checking and linting improves on all four projects with one worker. TypeScript changes from an 11.7% improvement with one worker to a 9.7% regression with four, consistent with a parallel load-balancing cost. Its one-worker combined timing variation is only 0.6%/1.0%. The small one-worker lint-only difference is inconclusive: baseline variation is 10.2%.

Lint-only allocated bytes increase by 16.9% (VS Code), 11.9% (TypeScript), 6.3% (TypeORM), and 22.3% (Vue) with one worker. Keeping four checker caches active introduces extra work compared with a single checker consuming the shared queue.

## Checker workload balance

Separate instrumented runs measure rule setup and listener wall time per checker. These are not total workload duration or CPU utilization, and instrumentation is excluded from the timing tables. The following percentages are the busiest checker’s share of summed rule time.

| Project    | Shared lint | Affinity lint | Shared after checking | Affinity after checking |
| ---------- | ----------: | ------------: | --------------------: | ----------------------: |
| vscode     |       25.1% |         26.8% |                 25.2% |                   27.4% |
| typescript |       25.3% |         35.1% |                 25.1% |                   36.3% |
| typeorm    |       25.4% |         28.0% |                 25.5% |                   25.7% |
| vue        |       25.5% |         28.3% |                 26.1% |                   28.2% |

For TypeScript’s full lint, affinity assigns 152/150/150/148 files to the four checkers, yet one checker accounts for 35.1% of measured rule time. The shared queue balances time at roughly 25% per checker by processing very different file counts: 50/194/221/135. Equal file counts do not imply equal work.

For the cold 10% TypeScript subset, affinity assigns 16/15/14/15 files and the busiest checker accounts for 48.2% of rule time, compared with 29.5% for the shared queue. Together with the one-worker results, this supports load imbalance as the main explanation for the parallel regressions. A future design needs to balance cache reuse against uneven rule costs.

The full profiles match full-lint and full-check cases. Filtered profiles use the selected files with semantic checking disabled/enabled; they do not reproduce the full-program warm preparation used by the warm timing cases.

## Repeatability

The four-worker full-project runs have timing coefficients of variation of
0.9–6.2%. TypeScript's lint-only samples are 2.345–2.430 s for the shared queue
and 3.091–3.171 s for affinity; the ranges do not overlap. Its combined samples
also do not overlap (3.101–3.382 s versus 3.405–3.520 s). No individual outliers were removed from the reported cases.
All samples from those cases are retained in the raw CSV, including the slower observations.

The synthetic workload improves combined checking and linting by 22.3%
(93.6 → 72.8 ms) and reduces allocated bytes by 26.4%. Its lint-only time is
approximately unchanged (58.7 → 59.8 ms). This demonstrates the potential cache
reuse benefit, but the real-project regressions limit the general conclusion.

## Scope and validation

This POC keeps each file on the compiler-owned checker assigned by TypeScript's
checker pool. It retains the existing lint worker limit, emits only nonempty
per-checker queues, and preserves input order within each queue. It adds no public
API, compiler patch, or shim changes. Custom checker pools retain their existing
empty-enumeration behavior. A program's checkers must remain exclusive to the
linter until it returns.

Builds use the pinned compiler plus the repository's existing patch stack.

Repository lint passed with zero issues. Unit tests for all internal packages
passed, as did targeted race tests for
checker ownership, exactly-once processing, checker exclusivity, and worker
limits. End-to-end tests passed (22 passed, one skipped). Complete fixture
diagnostics, fixes, suggestions, and labeled ranges matched across 80 isolated
runs, spanning one/four workers and semantic checking off/on.

There is still a correctness blocker: on the separate #1141 reproduction,
semantic checking followed by affinity linting consistently reports all 20
required assertions as unnecessary. The shared queue also exhibits the existing
bug, but on fewer files and sometimes nondeterministically. Semantic-enabled
real-project diagnostic counts also differ between implementations. These timing
results therefore characterize the POC; they do not establish equivalent output
or readiness to ship. The checker-state correctness fix is outside this POC. Removing the 20 reported assertions produces 20 `TS2339` errors in the [follow-up check](results/assertion-removal-diagnostics.json).

## Measurement conditions and machine contention

There are 1,800 accepted samples: 90 cases with ten samples per implementation.
Every case retains the planned five-sample A/B/B/A order. The initial timing run
was discarded at the user's request. The full four-worker matrix finished before
the later build contention and was preserved in full.

A first control pass was discarded after unrelated Go builds saturated the
machine; another interrupted case was restarted after Rust build activity.
Subsequent controls used a temporary load monitor. It polled every five seconds,
waited for a 30-second quiet interval, and interrupted runs after two consecutive
observations of build processes or more than 500% aggregate CPU usage by other
processes. Initially whole interrupted cases were restarted. Later, uncontended
prefixes were retained and only the affected suffix was repeated, preserving
A/B/B/A order. A sample was excluded conservatively if its whole-process end
was within five seconds before the first busy observation or later. Rejection
was based on machine activity, not on whether a result favored either version.

This is a shared desktop, not an isolated benchmark host. Polling cannot detect
every short burst, and some controls have more variation than the primary runs.
Small control differences should be treated as inconclusive. The raw samples and
per-case variation are retained; no individual timing outliers were trimmed.

New control rows include start timestamps and whole-process elapsed time; these
fields are blank for the earlier four-worker rows. Whole-process elapsed time
includes untimed warm preparation and must not be substituted for `ns/op`.
Cold cases include program creation and binding. Warm cases time the first lint
after full semantic checking, excluding preparation from time and allocation
measurements. OS filesystem caches were not flushed. These are Go linter API
measurements, not Oxlint CLI or incremental LSP measurements. See the
[reproduction instructions](README.md) for exact boundaries and corpus setup.

## Data and reproduction

- [Complete matrix](results/matrix.md), [raw samples](results/samples.csv), and [variation, allocation counts, and diagnostic counts](results/summary.json).
- [Selected-file manifests](results/manifests.json), [pinned corpus preparation](results/preparation.json), [source/build provenance](results/source.json), and [binary build settings](results/binary-build-info.txt).
- [Four-worker environment](results/environment-measure-four-workers.json), [control environment](results/environment-measure.json), and [load-monitor events](results/control-load-events.json).
- [Checker profiles](results/profiles.json) and [profiling environment](results/environment-profile.json). Profiling uses separate binaries; its output-only serialization change is recorded in the build provenance.
- [Repeated correctness results](results/validation.json) and [complete diagnostic snapshots](results/diagnostics/).
