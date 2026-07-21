# Benchmarks

[`eslint.config.mjs`](./eslint.config.mjs) includes only those rules implemented in **tsgolint**.

> [!NOTE]
> Biome and `deno lint` are not considered in this benchmark because they [do not support typed linting](https://www.joshuakgoldberg.com/blog/why-typed-linting-needs-typescript-today/).

## Results

> Measured July 20, 2026 on Apple M4 Pro (12 cores: 8 performance, 4 efficiency)

TypeORM and Vue were benchmarked with equivalent TypeScript 7-compatible configurations: TypeORM's removed `downlevelIteration` option was omitted and Node 16 module resolution was used, while Vue's removed `baseUrl` option was omitted and its path targets were made explicitly relative. The same adjusted configurations were used by both linters.

| Repository                                                      | ESLint + typescript-eslint | tsgolint | Speedup |
| --------------------------------------------------------------- | -------------------------- | -------- | ------- |
| [microsoft/vscode](https://github.com/microsoft/vscode)         | 83.2s                      | 6.96s    | **12x** |
| [microsoft/typescript](https://github.com/microsoft/typescript) | 27.2s                      | 1.94s    | **14x** |
| [typeorm/typeorm](https://github.com/typeorm/typeorm)           | 13.2s                      | 0.75s    | **18x** |
| [vuejs/core](https://github.com/vuejs/core)                     | 12.3s                      | 0.95s    | **13x** |

<details>

<summary>Detailed report</summary>

| microsoft/vscode |
| ---------------- |

```plaintext
Benchmark 1: eslint
  Time (mean ± σ):     83.239 s ±  2.889 s    [User: 109.053 s, System: 5.656 s]
  Range (min … max):   79.239 s … 87.427 s    10 runs

  Warning: Ignoring non-zero exit code.

Benchmark 2: tsgolint
  Time (mean ± σ):      6.964 s ±  0.241 s    [User: 32.351 s, System: 1.517 s]
  Range (min … max):    6.821 s …  7.593 s    10 runs

Summary
  tsgolint ran
   11.95 ± 0.59 times faster than eslint
```

| microsoft/typescript |
| -------------------- |

```plaintext
Benchmark 1: eslint
  Time (mean ± σ):     27.170 s ±  0.971 s    [User: 38.137 s, System: 1.305 s]
  Range (min … max):   26.383 s … 29.403 s    10 runs

  Warning: Ignoring non-zero exit code.

Benchmark 2: tsgolint
  Time (mean ± σ):      1.942 s ±  0.034 s    [User: 9.884 s, System: 0.513 s]
  Range (min … max):    1.911 s …  2.007 s    10 runs

Summary
  tsgolint ran
   13.99 ± 0.56 times faster than eslint
```

| typeorm/typeorm |
| --------------- |

```plaintext
Benchmark 1: eslint
  Time (mean ± σ):     13.168 s ±  0.419 s    [User: 19.208 s, System: 1.057 s]
  Range (min … max):   12.746 s … 13.862 s    10 runs

  Warning: Ignoring non-zero exit code.

Benchmark 2: tsgolint
  Time (mean ± σ):     748.9 ms ±  14.0 ms    [User: 3366.7 ms, System: 440.8 ms]
  Range (min … max):   723.2 ms … 773.0 ms    10 runs

Summary
  tsgolint ran
   17.58 ± 0.65 times faster than eslint
```

| vuejs/core |
| ---------- |

```plaintext
Benchmark 1: eslint
  Time (mean ± σ):     12.302 s ±  0.791 s    [User: 22.701 s, System: 0.620 s]
  Range (min … max):   11.394 s … 14.198 s    10 runs

  Warning: Ignoring non-zero exit code.

Benchmark 2: tsgolint
  Time (mean ± σ):     952.4 ms ±  53.1 ms    [User: 4270.1 ms, System: 247.8 ms]
  Range (min … max):   866.4 ms … 1038.5 ms    10 runs

Summary
  tsgolint ran
   12.92 ± 1.10 times faster than eslint
```

</details>

## How to run benchmarks

### Running in Docker/Podman

Prerequisites:

- Built `tsgolint` binary. See [README.md](../README.md) for how to build it.
- Docker/Podman

```shell
docker build --file ./Containerfile --progress plain ..

# or

podman build --file ./Containerfile --progress plain ..
```

### Running locally

Prerequisites:

- Built `tsgolint` binary. See [README.md](../README.md) for how to build it.
- Node.js & Corepack
- [`hyperfine`](https://github.com/sharkdp/hyperfine)

1. Clone the repositories
   ```bash
   ./clone-projects.sh
   ```
2. Install deps and setup ESLint configs
   ```bash
   ./setup.sh
   ```
3. Run benchmarks
   ```bash
   ./bench.sh
   ```
