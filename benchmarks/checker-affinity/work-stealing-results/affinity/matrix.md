# Complete measurement matrix

Medians of independent fresh-process samples. Positive deltas mean the candidate is slower or allocates more.

`cold` includes program creation and binding; `warm` measures the first lint after full semantic checking.

| Project | Scenario | Workers | Baseline ms | Candidate ms | Time delta | Baseline MiB | Candidate MiB | Allocation delta |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| typescript | full-check | 4 | 3310.44 | 2845.21 | -14.1% | 1969.15 | 2028.18 | +3.0% |
| typescript | full-lint | 4 | 3175.62 | 2423.23 | -23.7% | 1887.84 | 1891.53 | +0.2% |
| typescript | one-cold | 4 | 191.43 | 191.11 | -0.2% | 381.10 | 381.08 | -0.0% |
| typescript | one-warm | 4 | 4.41 | 4.77 | +8.3% | 1.23 | 1.23 | -0.3% |
| typescript | ten-cold | 4 | 938.31 | 775.09 | -17.4% | 716.54 | 716.74 | +0.0% |
| typescript | ten-warm | 4 | 588.06 | 483.11 | -17.8% | 169.46 | 186.04 | +9.8% |
