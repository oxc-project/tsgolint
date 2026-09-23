# Affinity with work stealing

The hybrid recovers TypeScript’s full-project balance while retaining part of the allocation benefit from checker affinity. It still regresses on TypeScript’s cold filtered selection. The existing [#1141](https://github.com/oxc-project/tsgolint/issues/1141) checker-state bug also remains, so this is a performance POC, not an unconditional adoption recommendation.

## Scheduler

Each compiler-owned checker first processes its assigned files. When its queue empties, it takes one pending file from the largest remaining queue, using its own checker. Checkers are never shared concurrently, and an in-progress file cannot be stolen or split.

Workers claim unstarted checker workloads before stealing, preserving affinity with one worker. Initial checkers are reserved before any worker starts, so a fast worker cannot consume several empty workloads and leave its peers without a checker. Active workers are capped by the number of checkers and selected files: a lone file stays on its warmed owner. The semantic-checking phase still runs beforehand with the compiler’s existing scheduler.

## Direct comparison with the original shared queue

Ten fresh-process samples per implementation and case, in five-sample A/B/B/A batches. Four configured workers, four compiler checkers, GOMAXPROCS=4. Times are medians in milliseconds; negative deltas favor the hybrid. Allocated bytes are not peak memory.

Several warm-subset cases have substantial timing variation; see the repeatability section before interpreting their exact percentages.

| Project    | Case       | Shared ms | Hybrid ms | Time delta | Bytes delta |
| ---------- | ---------- | --------: | --------: | ---------: | ----------: |
| typeorm    | full-check |    1320.0 |    1161.5 |     -12.0% |      -14.5% |
| typeorm    | full-lint  |    1027.7 |    1091.1 |      +6.2% |       +0.1% |
| typeorm    | ten-cold   |     366.1 |     373.9 |      +2.1% |       -0.2% |
| typeorm    | ten-warm   |      55.4 |      38.2 |     -31.0% |      -30.7% |
| typescript | full-check |    3066.2 |    2863.2 |      -6.6% |       -8.7% |
| typescript | full-lint  |    2352.5 |    2333.5 |      -0.8% |       +0.3% |
| typescript | ten-cold   |     643.5 |     785.3 |     +22.0% |       +1.2% |
| typescript | ten-warm   |     421.1 |     471.3 |     +11.9% |      -10.3% |
| vscode     | full-check |    7783.2 |    6058.9 |     -22.2% |      -15.6% |
| vscode     | full-lint  |    5400.1 |    5497.4 |      +1.8% |       -0.1% |
| vscode     | ten-cold   |    1436.1 |    1440.3 |      +0.3% |       +0.1% |
| vscode     | ten-warm   |     560.6 |     201.4 |     -64.1% |      -44.6% |
| vue        | full-check |    1464.8 |    1091.6 |     -25.5% |      -22.8% |
| vue        | full-lint  |     975.4 |     965.5 |      -1.0% |       +0.5% |
| vue        | ten-cold   |     261.5 |     226.2 |     -13.5% |       +1.6% |
| vue        | ten-warm   |     102.2 |      23.0 |     -77.4% |      -62.7% |

`full-lint` and `full-check` include fresh program creation and binding; the latter also emits semantic diagnostics. `ten-cold` selects a deterministic 10% subset but creates the full program. `ten-warm` times that subset’s first lint after full-program semantic checking outside the timer.

## Comparison with pure affinity

TypeScript only, six fresh-process samples per implementation in three-sample A/B/B/A batches. These are a separate comparison; do not combine absolute times from the two runs into a three-way timing table.

| Case       | Pure affinity ms | Hybrid ms | Time delta | Bytes delta |
| ---------- | ---------------: | --------: | ---------: | ----------: |
| full-check |          3310.44 |   2845.21 |     -14.1% |       +3.0% |
| full-lint  |          3175.62 |   2423.23 |     -23.7% |       +0.2% |
| one-cold   |           191.43 |    191.11 |      -0.2% |       -0.0% |
| one-warm   |             4.41 |      4.77 |      +8.3% |       -0.3% |
| ten-cold   |           938.31 |    775.09 |     -17.4% |       +0.0% |
| ten-warm   |           588.06 |    483.11 |     -17.8% |       +9.8% |

The single-file warm timing difference is inconclusive: its timing coefficients of variation are about 14%/12%. Allocations match the pure-affinity level, and a dedicated ownership test verifies that the lone file cannot be stolen.

## Why stealing helps, and where it stops helping

One separate instrumented run per implementation and condition measures rule setup and listener wall time, not total workload time or CPU utilization. Instrumentation is excluded from the timing tables.

| Selection | Semantic checking | Pure busiest share | Hybrid busiest share | Stolen files |
| --------- | ----------------- | -----------------: | -------------------: | -----------: |
| full      | disabled          |              35.2% |                25.2% |       95/600 |
| full      | enabled           |              36.3% |                25.7% |      158/600 |
| ten       | disabled          |              55.6% |                38.3% |        24/60 |
| ten       | enabled           |              63.7% |                36.7% |        25/60 |

The largest individual file cost in the full TypeScript run is `src/compiler/checker.ts`. The filtered selection concentrates work in files such as `tsbuildPublic.ts`, `parser.ts`, and `utilities.ts`. Stealing transfers queued work away from busy owners, but a late-starting expensive file can still leave a tail after other queues drain. It can also repeat type work in the thief’s cache. Starting likely expensive files earlier is a useful next experiment; it is not implemented here.

The filtered semantic profiles check only the selected files. They do not reproduce the full-program warm preparation used by `ten-warm` timing runs. Per-file timings show where instrumented rule time is spent; they should not be read as uninstrumented performance estimates.

## Repeatability and scope

Measured on the same Apple M2 Max / macOS 27 / Go 1.27.1 environment and pinned corpora as the [original experiment](REPORT.md). Final comparison binaries have identical dependency versions and build settings. Each sample executes one operation in a fresh process to avoid the global parsed-source cache. OS filesystem caches are not flushed.

A load monitor checked every five seconds. It required 30 seconds without compiler jobs and with other processes using at most 200% aggregate CPU before starting. Two consecutive busy observations stopped this task’s runner; samples overlapping the incident, with a five-second lookback, were quarantined. Retained samples preserve their A/B/B/A prefix. No samples were removed because of their measured performance. This remains a shared desktop experiment: short load spikes and thermal changes can escape periodic monitoring.

Small deltas should be interpreted using the distributions, not just medians. Cases in the direct comparison with timing CV above 8% in either implementation:

- typeorm / ten-warm: 33.5% shared, 11.5% hybrid.
- typescript / ten-warm: 11.2% shared, 11.1% hybrid.
- vscode / ten-cold: 20.2% shared, 5.8% hybrid.
- vscode / ten-warm: 25.1% shared, 12.5% hybrid.
- vue / ten-warm: 51.7% shared, 61.7% hybrid.

## Validation and correctness limit

Engine/CLI tests, targeted race tests (three repetitions), E2E (22 passed, one skipped), and lint checks for the changed Go packages pass. Scheduler tests cover exactly-once processing, checker exclusivity, worker limits, single-worker affinity, empty inputs, custom pools, single-file ownership, and four concurrent checkers serving files concentrated on one owner.

The concentrated-selection regression test fails on the first prototype and passes after reserving initial workloads. A second refinement caps workers by file count, eliminating pointless single-file steals observed in the warm benchmark.

All 24 ordinary fixture diagnostic runs match the original shared scheduler, including fixes and suggestions. The 48 separate #1141 characterization runs still expose the known bug: after semantic checking, the hybrid reports 14–20 of 20 required assertions as unnecessary with four workers, and all 20 with one worker. Those false positives do not occur without semantic checking. Real-project diagnostic counts can also differ; performance measurements do not establish semantic equivalence.

## Final handoff correction

The main timing matrix used the frozen hybrid binary before a final robustness correction for fewer workers than checkers. If a peer wins the last unstarted workload, an idle worker now retains its checker and falls back to stealing instead of exiting. The measured four-worker/four-checker cases have no unstarted workloads, and a one-file case has only one active worker, so neither can encounter that race. A deterministic regression test covers the lost-claim interleaving. Final-source checks and a separate comparison against the frozen binary are recorded alongside the primary measurements.

The supplemental comparison uses six fresh-process samples per implementation and case. TypeScript full-lint medians are 2483.4 → 2427.1 ms, and full-check medians are 2944.4 → 2899.0 ms. These small decreases are within the observed variation; no material slowdown was observed from the correction.

## Evidence and reproduction

- [Direct comparison: raw samples](work-stealing-results/shared/samples.csv), [distributions](work-stealing-results/shared/summary.json), [load events](work-stealing-results/shared/control-load-events.json).
- [Pure-affinity comparison](work-stealing-results/affinity/summary.json) and [raw samples](work-stealing-results/affinity/samples.csv).
- [Per-checker and per-file profiles](work-stealing-results/profiles/profiles.json).
- [Diagnostic comparisons](work-stealing-results/validation/validation.json) and [empty-queue regression log](work-stealing-results/empty-queue-regression.log).
- [Final handoff correction](work-stealing-results/final-handoff-fix.patch) and [supplemental final-binary comparison](work-stealing-results/final-check/summary.json).
- [Source and binary hashes](work-stealing-results/source.json), [build metadata](work-stealing-results/binary-build-info.txt), [pure-affinity patch](work-stealing-results/pure-affinity.patch), and [hybrid patch](work-stealing-results/work-stealing.patch).

To reconstruct implementations from the original baseline, apply `pure-affinity.patch` for pure affinity, then `work-stealing.patch` for the measured hybrid and `final-handoff-fix.patch` for the final two-worker handoff correction. Use the same current benchmark harness and pinned compiler in every build. The [runner documentation](README.md) describes corpus preparation and measurement boundaries; `--scenarios` selects focused cases. The `prototype-v1` and `prototype-v2` directories preserve earlier experiments and are excluded from the final timing tables.
