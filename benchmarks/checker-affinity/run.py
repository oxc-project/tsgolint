#!/usr/bin/env python3
"""Run isolated checker-scheduling comparisons (Python stdlib only)."""

import argparse
import csv
import gzip
import hashlib
import json
import math
import os
import platform
import statistics
import subprocess
import time
from pathlib import Path

from prepare import HISTORICAL_PROJECTS, PROJECTS, write_json


def canonical(records):
    return sorted(records or [], key=lambda d: json.dumps(d, sort_keys=True))


def digest(value):
    return hashlib.sha256(json.dumps(value, sort_keys=True).encode()).hexdigest()


class Experiment:
    def __init__(self, args):
        self.args = args
        self.output = args.output.resolve()
        self.output.mkdir(parents=True, exist_ok=True)
        self.binaries = (dict(args.variant) if args.variant else
                         {"baseline": str(args.baseline.resolve()), "candidate": str(args.candidate.resolve())})
        self.root = (args.project_root or args.projects).resolve()

    def invoke(self, version, case, benchmark=False):
        case_path = self.output / "current-case.json"
        result_path = self.output / "current-output.json"
        write_json(case_path, case)
        result_path.unlink(missing_ok=True)
        env = dict(os.environ, TSGOLINT_AFFINITY_CASE=str(case_path), TSGOLINT_AFFINITY_OUTPUT=str(result_path))
        if case["workers"]:
            env["GOMAXPROCS"] = str(case["workers"])
        else:
            env.pop("GOMAXPROCS", None)
        flags = (["-test.run=^$", "-test.bench=^BenchmarkCheckerAffinity$", "-test.benchtime=1x", "-test.benchmem"]
                 if benchmark else ["-test.run=^TestCheckerAffinityProbe$", "-test.count=1"])
        result = subprocess.run([self.binaries[version], *flags], env=env, text=True,
                                stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=self.args.timeout)
        if result.returncode:
            raise RuntimeError(f"{version} {case}:\n{result.stdout}")
        if not benchmark:
            return json.loads(result_path.read_text())
        line = next(line for line in result.stdout.splitlines() if line.startswith("BenchmarkCheckerAffinity"))
        fields = line.split()
        if fields[1] != "1":
            raise RuntimeError("measurement was not a single operation")
        return {fields[i + 1]: float(fields[i]) for i in range(2, len(fields), 2)}

    def case(self, name, workers=4):
        if self.args.project_root:
            config = (self.root / self.args.config).resolve()
            return {"root": str(self.root), "config": str(config), "workers": workers,
                    "rules": [], "semantic": False, "warm": False}
        project = PROJECTS.get(name) or HISTORICAL_PROJECTS.get(name)
        config = project[3] if project else "tsconfig.json"
        root = self.root / name
        return {"root": str(root), "config": str(root / config), "workers": workers,
                "rules": ["no-unsafe-assignment"] if name == "synthetic" else [],
                "semantic": False, "warm": False}

    def manifests(self):
        manifests = {}
        corpus = {}
        for name in self.args.names:
            case = self.case(name)
            case["mode"] = "manifest"
            reference, *others = self.binaries
            files = self.invoke(reference, case)
            for version in others:
                if files != self.invoke(version, case):
                    raise RuntimeError(f"different file manifests for {name}: {reference}/{version}")
            ranked = sorted(files, key=lambda f: hashlib.sha256(f.encode()).digest())
            one, ten = set(ranked[:1]), set(ranked[:math.ceil(len(files) / 10)])
            manifests[name] = {"full": files, "one": [p for p in files if p in one],
                               "ten": [p for p in files if p in ten]}
            counts = {}
            contents = hashlib.sha256()
            for path in files:
                source = Path(case["root"]) / path
                data = source.read_bytes()
                contents.update(json.dumps([path, hashlib.sha256(data).hexdigest()]).encode())
                count = counts.setdefault(source.suffix, {"files": 0, "physical_lines": 0})
                count["files"] += 1
                count["physical_lines"] += len(data.splitlines())
            corpus[name] = {"files": len(files), "by_extension": counts,
                            "selected_sources_sha256": contents.hexdigest(),
                            "config_sha256": hashlib.sha256(Path(case["config"]).read_bytes()).hexdigest()}
        corpus_path = self.output / "corpus.json"
        if (self.output / "samples.csv").exists() and corpus_path.exists() and json.loads(corpus_path.read_text()) != corpus:
            raise RuntimeError("selected sources/config changed; use a fresh output directory")
        write_json(corpus_path, corpus)
        write_json(self.output / "manifests.json", manifests)
        return manifests

    def validate(self):
        records_dir = self.output / "diagnostics"
        records_dir.mkdir(exist_ok=True)
        summary = []
        fixtures = self.args.repo.resolve() / "e2e/fixtures/basic"
        fixture_files = sorted(p.relative_to(fixtures).as_posix() for p in fixtures.rglob("*")
                               if p.suffix in (".ts", ".tsx", ".mts", ".cts") and "node_modules" not in p.parts)
        for dataset in ("fixtures", "issue1141"):
            for workers in (1, 4):
                for semantic in (False, True):
                    for rules in ([[], ["no-unnecessary-type-assertion"]] if dataset == "issue1141" else [[]]):
                        label = f"{dataset}-w{workers}-semantic{int(semantic)}-{'all' if not rules else 'assertion'}"
                        case = self.case("issue1141", workers)
                        case.update(mode="diagnostics", semantic=semantic, rules=rules)
                        if dataset == "fixtures":
                            case.update(root=str(fixtures), config=str(fixtures / "tsconfig.json"), files=fixture_files)
                        row = {"case": label, "runs": []}
                        first = {}
                        for repeat in range(self.args.repeats):
                            for version in self.binaries:
                                records = canonical(self.invoke(version, case))
                                first.setdefault(version, records)
                                checksum = digest(records)
                                file = records_dir / f"{checksum}.json.gz"
                                if not file.exists():
                                    with gzip.open(file, "wt") as stream:
                                        json.dump(records, stream, sort_keys=True)
                                required_assertions = [d for d in records if d.get("rule") == "no-unnecessary-type-assertion"
                                                       and (d.get("file_path") or "").startswith("module")]
                                row["runs"].append({"version": version, "repeat": repeat, "sha256": checksum,
                                                    "count": len(records), "required_assertion_reports": required_assertions})
                        row["first_runs_equal"] = first["baseline"] == first["candidate"]
                        for version in self.binaries:
                            row[version + "_variants"] = len({r["sha256"] for r in row["runs"] if r["version"] == version})
                        if not row["first_runs_equal"]:
                            a = {json.dumps(d, sort_keys=True) for d in first["baseline"]}
                            b = {json.dumps(d, sort_keys=True) for d in first["candidate"]}
                            row["removed"] = [json.loads(d) for d in sorted(a - b)]
                            row["added"] = [json.loads(d) for d in sorted(b - a)]
                        summary.append(row)
                        write_json(self.output / "validation.json", summary)
                        print(f"{label}: parity={row['first_runs_equal']}, variants={row['baseline_variants']}/{row['candidate_variants']}", flush=True)

    def measurements(self):
        manifests = self.manifests()
        # ABBA for two versions; ABCDDCBA for four. Every version occupies
        # mirrored positions so a gradual machine-speed drift is balanced.
        order = [*self.binaries, *reversed(self.binaries)]
        scenarios = [("full-lint", "full", False, False), ("full-check", "full", True, False),
                     ("one-cold", "one", False, False), ("ten-cold", "ten", False, False),
                     ("one-warm", "one", False, True), ("ten-warm", "ten", False, True)]
        rows = []
        csv_path = self.output / "samples.csv"
        if csv_path.exists():
            with csv_path.open() as stream:
                rows = list(csv.DictReader(stream))
        completed = {(r["project"], r["scenario"], str(r["worker_setting"])) for r in rows
                     if sum(x["project"] == r["project"] and x["scenario"] == r["scenario"] and
                            str(x["worker_setting"]) == str(r["worker_setting"]) for x in rows) == self.args.samples * len(self.binaries)}
        for workers in self.args.workers:
            for name in self.args.names:
                for scenario, selection, semantic, warm in scenarios:
                    if self.args.scenarios and scenario not in self.args.scenarios:
                        continue
                    key = (name, scenario, str(workers))
                    if key in completed:
                        continue
                    previous = [r for r in rows if (r["project"], r["scenario"], str(r["worker_setting"])) == key]
                    if self.args.resume_partial:
                        expected = [(version, batch, sample)
                                    for batch, version in enumerate(order)
                                    for sample in range(self.args.samples // 2)]
                        actual = [(r["version"], int(r["batch"]), int(r["sample"])) for r in previous]
                        if actual != expected[:len(actual)]:
                            raise RuntimeError(f"partial case is not an ordered ABBA prefix: {key}")
                        resume_at = len(previous)
                    else:
                        # Restart partial cases unless a load monitor has retained a valid prefix.
                        rows = [r for r in rows if (r["project"], r["scenario"], str(r["worker_setting"])) != key]
                        resume_at = 0
                    case = self.case(name, workers)
                    case.update(files=manifests[name][selection], semantic=semantic, warm=warm)
                    for batch, version in enumerate(order):
                        for sample in range(self.args.samples // 2):
                            if batch * (self.args.samples // 2) + sample < resume_at:
                                continue
                            started = time.time()
                            metrics = self.invoke(version, case, benchmark=True)
                            rows.append({"project": name, "scenario": scenario, "worker_setting": workers,
                                         "version": version, "batch": batch, "sample": sample,
                                         "started_at": round(started, 3),
                                         "elapsed_seconds": round(time.time() - started, 3), **metrics})
                            temporary = csv_path.with_suffix(".tmp")
                            with temporary.open("w", newline="") as stream:
                                writer = csv.DictWriter(stream, fieldnames=list(rows[-1]))
                                writer.writeheader()
                                writer.writerows(rows)
                            temporary.replace(csv_path)
                    medians = {v: statistics.median(float(r["ns/op"]) for r in rows
                               if (r["project"], r["scenario"], str(r["worker_setting"])) == key and r["version"] == v)
                               for v in self.binaries}
                    times = ", ".join(f"{v}={value/1e6:.1f}" for v, value in medians.items())
                    print(f"{name} {scenario} workers={workers or 'host'}: {times} ms", flush=True)

    def profiles(self):
        manifests = self.manifests()
        results = []
        for name in self.args.names:
            for selection in ("full", "ten"):
                for semantic in (False, True):
                    case = self.case(name)
                    case.update(files=manifests[name][selection], semantic=semantic, mode="profile")
                    for version in self.binaries:
                        profiles = self.invoke(version, case)
                        total = sum(p["rule_ns"] for p in profiles)
                        results.append({"project": name, "selection": selection, "semantic": semantic,
                                        "version": version, "checkers": profiles,
                                        "active_checkers": sum(p["files"] > 0 for p in profiles),
                                        "slowest_share": max(p["rule_ns"] for p in profiles) / total if total else 0})
                        write_json(self.output / "profiles.json", results)
            print(f"profiled {name}", flush=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--baseline", type=Path)
    parser.add_argument("--candidate", type=Path)
    parser.add_argument("--variant", action="append", metavar="NAME=PATH",
                        help="named measurement/profile variant; repeat instead of --baseline/--candidate")
    projects = parser.add_mutually_exclusive_group(required=True)
    projects.add_argument("--projects", type=Path, help="prepared public corpus directory")
    projects.add_argument("--project-root", type=Path, help="existing local project; no downloads or installs")
    parser.add_argument("--config", type=Path, default=Path("tsconfig.json"), help="local config, relative to project root or absolute")
    parser.add_argument("--project-name", default="local", help="local project's label in the results")
    parser.add_argument("--binaries-dir", type=Path, help="directory produced by build.py; selects all four variants")
    parser.add_argument("--timeout", type=int, default=600, help="timeout in seconds per subprocess")
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--repo", type=Path, default=Path(__file__).resolve().parents[2])
    parser.add_argument("--phase", choices=("validate", "measure", "profile", "all"), default="all")
    parser.add_argument("--names", nargs="+", default=["synthetic", *PROJECTS])
    parser.add_argument("--workers", nargs="+", type=int, default=[4, 1, 0], help="0 uses the host default")
    parser.add_argument("--samples", type=int, default=10)
    parser.add_argument("--scenarios", nargs="+",
                        choices=("full-lint", "full-check", "one-cold", "ten-cold", "one-warm", "ten-warm"))
    parser.add_argument("--repeats", type=int, default=10)
    parser.add_argument("--resume-partial", action="store_true",
                        help="resume an ordered ABBA prefix retained by an external load monitor")
    args = parser.parse_args()
    if args.project_root:
        if args.phase not in ("measure", "profile"):
            parser.error("--project-root requires --phase measure/profile")
        if not (args.project_root / args.config).is_file():
            parser.error("local project's config does not exist")
        args.names = [args.project_name]
    if any(worker < 0 for worker in args.workers) or args.timeout <= 0:
        parser.error("workers must be nonnegative and timeout positive")
    if args.binaries_dir:
        if args.variant or args.baseline or args.candidate:
            parser.error("--binaries-dir cannot be mixed with other binary options")
        build_path = args.binaries_dir / "build-manifest.json"
        if not build_path.is_file():
            parser.error("missing build-manifest.json; finish build.py first")
        build = json.loads(build_path.read_text())
        args.variant = []
        for name in ("baseline", "affinity", "stealing", "sorted"):
            entry = build["variants"][name]
            binary = args.binaries_dir / entry["binary"]
            if hashlib.sha256(binary.read_bytes()).hexdigest() != entry["binary_sha256"]:
                parser.error(f"binary hash mismatch: {name}; rebuild all variants")
            args.variant.append(f"{name}={binary}")
    if args.variant:
        if args.baseline or args.candidate or args.phase not in ("measure", "profile"):
            parser.error("--variant requires --phase measure/profile and cannot be mixed with --baseline/--candidate")
        variants = []
        for value in args.variant:
            name, separator, path = value.partition("=")
            if not separator or not name or not path or name in dict(variants):
                parser.error("each --variant must have a unique NAME and a PATH")
            variants.append((name, str(Path(path).resolve())))
        if len(variants) < 2:
            parser.error("provide at least two --variant options")
        args.variant = variants
    elif not args.baseline or not args.candidate:
        parser.error("provide --baseline and --candidate, or repeat --variant")
    if args.samples < 2 or args.samples % 2:
        parser.error("samples must be a positive even number")
    experiment = Experiment(args)
    metadata = {"platform": platform.platform(), "machine": platform.machine(), "logical_cpus": os.cpu_count(),
                "workers": args.workers, "projects": args.names, "samples_per_version": args.samples,
                "scenarios": args.scenarios,
                "project_root": str(experiment.root), "config": str(args.config),
                "go": subprocess.check_output(["go", "version"], text=True).strip(),
                "started": time.strftime("%Y-%m-%dT%H:%M:%S%z"),
                "environment": {k: os.environ.get(k) for k in ("GOGC", "GOMEMLIMIT", "GOMAXPROCS")},
                "binaries": {k: hashlib.sha256(Path(v).read_bytes()).hexdigest() for k, v in experiment.binaries.items()}}
    if platform.system() == "Darwin":
        cpu = subprocess.run(["sysctl", "-n", "machdep.cpu.brand_string"], text=True, capture_output=True)
        metadata["cpu"] = cpu.stdout.strip() if cpu.returncode == 0 else "unavailable"
    environment_path = experiment.output / ("environment-" + args.phase + ".json")
    if args.phase == "measure" and (experiment.output / "samples.csv").exists() and environment_path.exists():
        previous = json.loads(environment_path.read_text())
        for key in ("binaries", "workers", "projects", "samples_per_version", "scenarios", "environment", "project_root", "config"):
            if key in previous and previous[key] != metadata[key]:
                parser.error(f"{key} changed since previous samples; use a fresh output directory")
    write_json(environment_path, metadata)
    if args.binaries_dir:
        write_json(experiment.output / "build-manifest.json", build)
    for phase, method in (("validate", experiment.validate), ("measure", experiment.measurements), ("profile", experiment.profiles)):
        if args.phase in (phase, "all"):
            method()


if __name__ == "__main__":
    main()
