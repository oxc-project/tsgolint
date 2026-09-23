#!/usr/bin/env python3
"""Prepare disposable, pinned projects for the checker-affinity experiment."""

import argparse
import hashlib
import json
import os
import subprocess
from pathlib import Path


PROJECTS = {
    "vscode": ("microsoft/vscode", "1.99.0", "4437686ffebaf200fa4a6e6e67f735f3edf24ada", "src/tsconfig.json"),
    "typescript": ("microsoft/typescript", "v5.8.2", "beb69e4cdd61b1a0fd9ae21ae58bd4bd409d7217", "src/tsconfig-eslint.json"),
    "typeorm": ("typeorm/typeorm", "0.3.22", "6c5668bd82233301642593a83236cc4ae315d6fc", "tsconfig.json"),
    # Exact typed-ESLint revision linked from tsgolint issue #116.
    "sentry": ("getsentry/sentry", "9ed4ed8c78e6163eb443023a05ef37b3739cb82a",
               "9ed4ed8c78e6163eb443023a05ef37b3739cb82a", "tsconfig.json"),
}

HISTORICAL_PROJECTS = {
    "vue": ("vuejs/core", "v3.5.13", "6eb29d345aa73746207f80c89ee8b37ff7b949c9", "tsconfig.json"),
}


def run(args, cwd=None):
    return subprocess.check_output(args, cwd=cwd, text=True).strip()


def write_json(path, value):
    path.write_text(json.dumps(value, indent=2) + "\n")


def generated_projects(root):
    config = {"compilerOptions": {"strict": True, "target": "esnext", "types": [], "skipLibCheck": True}, "include": ["*.ts"]}
    synthetic = root / "synthetic"
    synthetic.mkdir(exist_ok=True)
    write_json(synthetic / "tsconfig.json", config)
    for module in range(64):
        # Unique names prevent declaration merging even if module detection changes.
        code = []
        for i in range(64):
            suffix = f"{module}_{i}"
            code.extend([
                f"type Copy{suffix}<T> = {{ [K in keyof T]: T[K] }};",
                f"declare function copy{suffix}<T>(value: T): Copy{suffix}<T>;",
                f"export const value{suffix} = copy{suffix}({{ count: 0, name: 'value', nested: {{ enabled: true }} }});",
            ])
        (synthetic / f"module{module:02}.ts").write_text("\n".join(code) + "\n")

    issue = root / "issue1141"
    issue.mkdir(exist_ok=True)
    write_json(issue / "tsconfig.json", config)
    (issue / "declarations.d.ts").write_text(
        "export interface Base { id: string }\n"
        "export interface Derived extends Base { extra: number }\n"
        "export declare function query<T extends Base = Base>(key: string): T;\n"
    )
    for i in range(20):
        (issue / f"module{i:02}.ts").write_text(
            "import { query, type Derived } from './declarations';\n"
            "export const run = (): number => {\n"
            "  const v = query('k') as Derived;\n"
            "  return v.extra;\n"
            "};\n"
        )
    (issue / "controls.ts").write_text(
        "declare function identity<T>(value: T): T;\n"
        "export const inferred = identity({ extra: 1 }) as { extra: number };\n"
        "export const redundant = 'hello' as string;\n"
    )


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("directory", type=Path, help="disposable corpus directory")
    parser.add_argument("--skip-install", action="store_true")
    parser.add_argument("--generated-only", action="store_true")
    parser.add_argument("--names", nargs="+", choices=[*PROJECTS, *HISTORICAL_PROJECTS], default=list(PROJECTS),
                        help="projects to prepare; Vue is available only for historical reproductions")
    parser.add_argument("--node", type=Path, help="absolute Node executable, bypassing directory-dependent version managers")
    parser.add_argument("--npm-cli", type=Path, help="npm-cli.js to run with --node")
    parser.add_argument("--legacy-node", type=Path, help="optional Node executable for TypeORM's older npm 8 lockfile tooling")
    args = parser.parse_args()
    root = args.directory.resolve()
    root.mkdir(parents=True, exist_ok=True)
    generated_projects(root)
    metadata_path = root / "preparation.json"
    metadata = json.loads(metadata_path.read_text()) if metadata_path.exists() else {}
    projects = {**HISTORICAL_PROJECTS, **PROJECTS}
    for name in args.names:
        if args.generated_only:
            break
        repo, tag, revision, config = projects[name]
        directory = root / name
        if not directory.exists():
            if tag == revision:
                directory.mkdir()
                run(["git", "init", "-q"], directory)
                run(["git", "remote", "add", "origin", "https://github.com/" + repo], directory)
                run(["git", "fetch", "--depth", "1", "origin", revision], directory)
                run(["git", "checkout", "--detach", "FETCH_HEAD"], directory)
            else:
                subprocess.run(["git", "clone", "--depth", "1", "--single-branch", "--branch", tag,
                                "https://github.com/" + repo, str(directory)], check=True)
        if run(["git", "rev-parse", "HEAD"], directory) != revision:
            raise SystemExit(f"{directory} is not at the pinned revision; use a new disposable directory")
        if name in ("typeorm", "vue"):
            original = run(["git", "show", "HEAD:tsconfig.json"], directory)
            parsed = json.loads(original)
            options = parsed["compilerOptions"]
            if name == "typeorm":
                options.pop("downlevelIteration")
                options.update(module="node16", moduleResolution="node16")
            else:
                assert options.pop("baseUrl") == "."
                options["paths"] = {k: ["./" + p for p in paths] for k, paths in options["paths"].items()}
            write_json(directory / "tsconfig.json", parsed)
        if name == "sentry":
            # TS7 rejects these legacy options. ES2022 uses native iteration;
            # alwaysStrict=false is no longer supported by the compiler.
            original = run(["git", "show", "HEAD:config/tsconfig.base.json"], directory) + "\n"
            for removed in ('    "downlevelIteration": true,\n', '    "alwaysStrict": false,\n'):
                assert original.count(removed) == 1
                original = original.replace(removed, "")
            (directory / "config/tsconfig.base.json").write_text(original)
        npm = ["npm"]
        npx = ["npx"]
        env = dict(os.environ)
        if args.node and args.npm_cli:
            npm = [str(args.node), str(args.npm_cli)]
            npx = [str(args.node), str(args.npm_cli.with_name("npx-cli.js"))]
            env["PATH"] = str(args.node.parent) + os.pathsep + env["PATH"]
        if name in ("vue", "sentry"):
            manager = "pnpm@10.10.0" if name == "sentry" else "pnpm@9.12.3"
            if name == "sentry":
                assert json.loads((directory / "package.json").read_text())["packageManager"] == manager
            install = npx + ["--yes", manager, "install", "--ignore-scripts", "--frozen-lockfile"]
        elif name == "typeorm":
            # New npm versions reject this historical lock's optional fsevents
            # entries. npm 8 consumes it without rewriting the pinned lockfile.
            if args.legacy_node and args.npm_cli:
                npx = [str(args.legacy_node), str(args.npm_cli.with_name("npx-cli.js"))]
                env["PATH"] = str(args.legacy_node.parent) + os.pathsep + env["PATH"]
            install = npx + ["--yes", "npm@8.19.4", "ci", "--ignore-scripts", "--no-audit", "--no-fund"]
        else:
            install = npm + ["ci", "--ignore-scripts", "--no-audit", "--no-fund"]
        if not args.skip_install:
            print(f"Installing {name} dependencies", flush=True)
            with (root / f"{name}-install.log").open("w") as log:
                subprocess.run(install, cwd=directory, env=env, stdout=log, stderr=subprocess.STDOUT, check=True)
        lock = directory / ("pnpm-lock.yaml" if name in ("vue", "sentry") else "package-lock.json")
        if run(["git", "diff", "--", lock.name], directory):
            raise SystemExit(f"dependency install changed the pinned lockfile: {lock}")
        metadata[name] = {
            "repo": repo, "ref": tag, "revision": revision, "config": config,
            "install": install, "lock_sha256": hashlib.sha256(lock.read_bytes()).hexdigest(),
            "node_version": run([str(args.node) if args.node else "node", "--version"]),
            "config_diff": run(["git", "diff", "--", "tsconfig.json", "config/tsconfig.base.json"], directory),
        }
    write_json(root / "preparation.json", metadata)


if __name__ == "__main__":
    main()
