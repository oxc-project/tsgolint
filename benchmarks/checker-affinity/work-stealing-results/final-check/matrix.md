# Complete measurement matrix

Medians of independent fresh-process samples. Positive deltas mean the candidate is slower or allocates more.

`cold` includes program creation and binding; `warm` measures the first lint after full semantic checking.

| Project | Scenario | Workers | Baseline ms | Candidate ms | Time delta | Baseline MiB | Candidate MiB | Allocation delta |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| typescript | full-check | 4 | 2944.42 | 2899.00 | -1.5% | 2028.61 | 2029.49 | +0.0% |
| typescript | full-lint | 4 | 2483.40 | 2427.15 | -2.3% | 1892.22 | 1892.37 | +0.0% |
