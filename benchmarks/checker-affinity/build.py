#!/usr/bin/env python3
"""Build all four checker scheduling variants without changing checkout sources."""

import argparse
import hashlib
import json
import os
import shutil
import subprocess
from pathlib import Path

HERE = Path(__file__).resolve().parent
REPO = HERE.parents[1]
VARIANTS = {
    "baseline": ("baseline", "unsorted"),
    "affinity": ("affinity", "unsorted"),
    "stealing": ("stealing", "unsorted"),
    "sorted": ("stealing", "sorted"),
}


def run(*command, cwd=REPO):
    subprocess.run(command, cwd=cwd, check=True)


def capture(*command, cwd=REPO):
    return subprocess.check_output(command, cwd=cwd, text=True).strip()


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def initialize():
    compiler = REPO / "typescript-go"
    if not (compiler / ".git").exists():
        run("git", "submodule", "update", "--init", "--", "typescript-go")
    expected = capture("git", "rev-parse", "HEAD:typescript-go")
    actual = capture("git", "rev-parse", "HEAD", cwd=compiler)
    if actual != expected or capture("git", "status", "--porcelain", cwd=compiler):
        raise SystemExit("--init requires a clean submodule at the recorded upstream commit. "
                         "Use a fresh clone, or omit --init if patches were already applied. "
                         "No existing checkout has been reset.")
    run("git", "-c", "user.name=Benchmark Setup", "-c", "user.email=benchmark@localhost",
        "am", "--3way", "--no-gpg-sign", *map(str, sorted((REPO / "patches").glob("*.patch"))), cwd=compiler)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--init", action="store_true", help="initialize and patch a fresh checkout's submodule once")
    parser.add_argument("--output", type=Path, default=HERE / "local/bin")
    parser.add_argument("--jobs", type=int, default=2, help="Go build parallelism (builds are sequential)")
    args = parser.parse_args()
    if args.jobs < 1:
        parser.error("--jobs must be positive")
    if args.init:
        initialize()
    collections = REPO / "typescript-go/internal/collections"
    if not collections.is_dir():
        parser.error("initialize a fresh checkout with --init first")
    destination = REPO / "internal/collections"
    destination.mkdir(exist_ok=True)
    for path in collections.glob("*.go"):
        if not path.name.endswith("_test.go"):
            shutil.copyfile(path, destination / path.name)
    output = args.output.resolve()
    output.mkdir(parents=True, exist_ok=True)
    manifest_path = output / "build-manifest.json"
    manifest_path.unlink(missing_ok=True)  # Never leave old provenance beside a partial build.
    manifest = {"commit": capture("git", "rev-parse", "HEAD"),
                "go": capture("go", "version"),
                "harness_sha256": sha(REPO / "cmd/tsgolint/checker_affinity_bench_test.go"),
                "variants": {}}
    for name, (linter, checker) in VARIANTS.items():
        sources = {"internal/linter/linter.go": HERE / f"variants/{linter}-linter.go.txt",
                   "typescript-go/internal/compiler/checkerpool.go": HERE / f"variants/{checker}-checkerpool.go.txt"}
        replacements = {str(REPO / target): str(source) for target, source in sources.items()}
        # Scheduler unit tests refer to implementation-specific private helpers.
        replacements[str(REPO / "internal/linter/checker_affinity_test.go")] = ""
        overlay = output / f"{name}-overlay.json"
        overlay.write_text(json.dumps({"Replace": replacements}, indent=2) + "\n")
        binary = output / (name + (".test.exe" if os.name == "nt" else ".test"))
        print(f"Building {name}", flush=True)
        run("go", "test", "-p", str(args.jobs), "-c", "-overlay", str(overlay), "-o", str(binary), "./cmd/tsgolint")
        build_info = capture("go", "version", "-m", str(binary))
        (output / f"{name}-build-info.txt").write_text(build_info + "\n")
        manifest["variants"][name] = {"binary": binary.name, "binary_sha256": sha(binary),
                                      "sources": {target: sha(source) for target, source in sources.items()}}
    manifest_path.write_text(json.dumps(manifest, indent=2) + "\n")
    print(f"Built all four variants in {output}")


if __name__ == "__main__":
    main()
