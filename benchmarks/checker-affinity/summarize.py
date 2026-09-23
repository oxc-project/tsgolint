#!/usr/bin/env python3
"""Summarize raw isolated samples; diagnostic counts are reported independently."""

import argparse
import csv
import json
import statistics
from collections import defaultdict
from pathlib import Path


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("directory", type=Path)
    args = parser.parse_args()
    groups = defaultdict(lambda: defaultdict(list))
    variant_names = []
    with (args.directory / "samples.csv").open() as stream:
        for row in csv.DictReader(stream):
            if row["version"] not in variant_names:
                variant_names.append(row["version"])
            groups[(row["project"], row["scenario"], int(row["worker_setting"]))][row["version"]].append(row)
    paired = set(variant_names) == {"baseline", "candidate"}
    summaries = []
    lines = ["# Complete measurement matrix", "",
             "Medians of independent fresh-process samples. Positive deltas mean the candidate is slower or allocates more.", "",
             "`cold` includes program creation and binding; `warm` measures the first lint after full semantic checking.", "",
             "| Project | Scenario | Workers | Baseline ms | Candidate ms | Time delta | Baseline MiB | Candidate MiB | Allocation delta |",
             "| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |"]
    if not paired:
        lines = ["# Complete measurement matrix", "",
                 "Median milliseconds; bold marks the lowest observed median, not statistical significance.", "",
                 "| Project | Scenario | Workers | " + " | ".join(variant_names) + " |",
                 "| --- | --- | ---: | " + " | ".join("---:" for _ in variant_names) + " |"]
    for (project, scenario, workers), versions in sorted(groups.items()):
        row = {"project": project, "scenario": scenario, "worker_setting": workers}
        for version in variant_names:
            data = versions[version]
            if not data:
                raise SystemExit(f"missing {version}: {project} {scenario} {workers}")
            metrics = {}
            for metric in ("ns/op", "B/op", "allocs/op"):
                values = [float(r[metric]) for r in data]
                metrics[metric] = {"median": statistics.median(values), "min": min(values), "max": max(values),
                                   "stdev": statistics.stdev(values) if len(values) > 1 else 0,
                                   "relative_stdev_percent": statistics.stdev(values) / statistics.mean(values) * 100 if len(values) > 1 else 0}
            metrics.update(samples=len(data), diagnostics=sorted({int(float(r["diagnostics/op"])) for r in data}),
                           type_diagnostics=sorted({int(float(r["type-diagnostics/op"])) for r in data}),
                           checkers=sorted({int(float(r["checkers"])) for r in data}),
                           actual_workers=sorted({int(float(r["workers"])) for r in data}),
                           gomaxprocs=sorted({int(float(r["gomaxprocs"])) for r in data}))
            row[version] = metrics
        if not paired:
            best = min(row[v]["ns/op"]["median"] for v in variant_names)
            cells = []
            for v in variant_names:
                value = row[v]["ns/op"]["median"]
                formatted = f"{value/1e6:.2f}"
                cells.append(f"**{formatted}**" if value == best else formatted)
            lines.append(f"| {project} | {scenario} | {workers or 'host'} | " + " | ".join(cells) + " |")
            summaries.append(row)
            continue
        baseline, candidate = row["baseline"], row["candidate"]
        row["time_delta_percent"] = (candidate["ns/op"]["median"] / baseline["ns/op"]["median"] - 1) * 100
        row["bytes_delta_percent"] = (candidate["B/op"]["median"] / baseline["B/op"]["median"] - 1) * 100
        summaries.append(row)
        lines.append(f"| {project} | {scenario} | {workers or 'host'} | {baseline['ns/op']['median']/1e6:.2f} | "
                     f"{candidate['ns/op']['median']/1e6:.2f} | {row['time_delta_percent']:+.1f}% | "
                     f"{baseline['B/op']['median']/2**20:.2f} | {candidate['B/op']['median']/2**20:.2f} | "
                     f"{row['bytes_delta_percent']:+.1f}% |")
    (args.directory / "summary.json").write_text(json.dumps(summaries, indent=2) + "\n")
    (args.directory / "matrix.md").write_text("\n".join(lines) + "\n")


if __name__ == "__main__":
    main()
