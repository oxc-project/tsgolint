# Complete measurement matrix

Medians of independent fresh-process samples. Positive deltas mean the candidate is slower or allocates more.

`cold` includes program creation and binding; `warm` measures the first lint after full semantic checking.

| Project | Scenario | Workers | Baseline ms | Candidate ms | Time delta | Baseline MiB | Candidate MiB | Allocation delta |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| typescript | full-check | 4 | 3867.83 | 3027.29 | -21.7% | 1969.21 | 2028.96 | +3.0% |
| typescript | full-lint | 4 | 3313.54 | 2425.14 | -26.8% | 1887.48 | 1899.04 | +0.6% |
| typescript | one-cold | 4 | 196.52 | 204.70 | +4.2% | 381.18 | 381.19 | +0.0% |
| typescript | one-warm | 4 | 4.89 | 4.97 | +1.5% | 1.23 | 1.23 | +0.1% |
| typescript | ten-cold | 4 | 995.01 | 830.55 | -16.5% | 716.38 | 714.20 | -0.3% |
| typescript | ten-warm | 4 | 681.16 | 539.34 | -20.8% | 169.61 | 186.07 | +9.7% |
