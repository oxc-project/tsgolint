# Complete measurement matrix

Medians of independent fresh-process samples. Positive deltas mean the candidate is slower or allocates more.

`cold` includes program creation and binding; `warm` measures the first lint after full semantic checking.

| Project | Scenario | Workers | Baseline ms | Candidate ms | Time delta | Baseline MiB | Candidate MiB | Allocation delta |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| typescript | full-check | 4 | 3436.15 | 2958.32 | -13.9% | 1969.02 | 2031.98 | +3.2% |
| typescript | full-lint | 4 | 3130.86 | 2403.19 | -23.2% | 1887.75 | 1892.41 | +0.2% |
| typescript | one-cold | 4 | 197.24 | 189.04 | -4.2% | 381.14 | 381.22 | +0.0% |
| typescript | one-warm | 4 | 5.03 | 7.29 | +44.8% | 1.23 | 1.58 | +28.4% |
| typescript | ten-cold | 4 | 965.26 | 791.31 | -18.0% | 716.77 | 710.52 | -0.9% |
| typescript | ten-warm | 4 | 596.88 | 488.31 | -18.2% | 169.81 | 185.94 | +9.5% |
