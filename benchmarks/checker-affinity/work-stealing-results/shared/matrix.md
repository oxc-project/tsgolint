# Complete measurement matrix

Medians of independent fresh-process samples. Positive deltas mean the candidate is slower or allocates more.

`cold` includes program creation and binding; `warm` measures the first lint after full semantic checking.

| Project | Scenario | Workers | Baseline ms | Candidate ms | Time delta | Baseline MiB | Candidate MiB | Allocation delta |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| typeorm | full-check | 4 | 1320.01 | 1161.51 | -12.0% | 1170.07 | 1000.54 | -14.5% |
| typeorm | full-lint | 4 | 1027.68 | 1091.08 | +6.2% | 942.57 | 943.50 | +0.1% |
| typeorm | ten-cold | 4 | 366.12 | 373.89 | +2.1% | 391.81 | 391.11 | -0.2% |
| typeorm | ten-warm | 4 | 55.37 | 38.21 | -31.0% | 56.67 | 39.29 | -30.7% |
| typescript | full-check | 4 | 3066.19 | 2863.16 | -6.6% | 2225.56 | 2031.59 | -8.7% |
| typescript | full-lint | 4 | 2352.55 | 2333.50 | -0.8% | 1888.66 | 1894.20 | +0.3% |
| typescript | ten-cold | 4 | 643.51 | 785.28 | +22.0% | 703.72 | 712.27 | +1.2% |
| typescript | ten-warm | 4 | 421.08 | 471.25 | +11.9% | 207.25 | 185.97 | -10.3% |
| vscode | full-check | 4 | 7783.18 | 6058.91 | -22.2% | 6054.51 | 5111.51 | -15.6% |
| vscode | full-lint | 4 | 5400.10 | 5497.41 | +1.8% | 4880.40 | 4873.61 | -0.1% |
| vscode | ten-cold | 4 | 1436.05 | 1440.33 | +0.3% | 1943.28 | 1945.48 | +0.1% |
| vscode | ten-warm | 4 | 560.56 | 201.42 | -64.1% | 240.82 | 133.39 | -44.6% |
| vue | full-check | 4 | 1464.85 | 1091.61 | -25.5% | 1279.65 | 987.99 | -22.8% |
| vue | full-lint | 4 | 975.45 | 965.46 | -1.0% | 938.37 | 942.97 | +0.5% |
| vue | ten-cold | 4 | 261.48 | 226.16 | -13.5% | 305.67 | 310.43 | +1.6% |
| vue | ten-warm | 4 | 102.21 | 23.05 | -77.4% | 48.42 | 18.04 | -62.7% |
