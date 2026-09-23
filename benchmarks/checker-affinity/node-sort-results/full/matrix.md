# Complete measurement matrix

Median milliseconds; bold marks the lowest observed median, not statistical significance.

| Project | Scenario | Workers | baseline | affinity | stealing | sorted |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| typeorm | full-check | 4 | 1376.36 | 1171.80 | **1155.75** | 1165.60 |
| typeorm | full-lint | 4 | 1028.34 | 1086.54 | **1020.07** | 1050.57 |
| typescript | full-check | 4 | 3325.85 | 3575.62 | **3004.96** | 3029.18 |
| typescript | full-lint | 4 | **2354.80** | 3099.47 | 2381.72 | 2471.86 |
| vscode | full-check | 4 | 8173.00 | 6595.63 | **6227.04** | 6307.28 |
| vscode | full-lint | 4 | 5802.32 | 6023.89 | 5446.84 | **5276.63** |
| vue | full-check | 4 | 1483.80 | 1134.99 | **1131.06** | 1197.72 |
| vue | full-lint | 4 | 1032.12 | 1053.87 | 1004.78 | **996.23** |
