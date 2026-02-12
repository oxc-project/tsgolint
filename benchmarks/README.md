# Benchmarks

[`eslint.config.mjs`](./eslint.config.mjs) includes only those rules implemented in **tsgolint**.

> [!NOTE]
> Biome and `deno lint` are not considered in this benchmark because they [do not support typed linting](https://www.joshuakgoldberg.com/blog/why-typed-linting-needs-typescript-today/).

## Results

> On AMD Ryzen 7 5800H (8 cores, 16 threads)

| Repository                                                      | ESLint + typescript-eslint | tsgolint | Speedup |
| --------------------------------------------------------------- | -------------------------- | -------- | ------- |
| [microsoft/vscode](https://github.com/microsoft/vscode)         | 167.8s                     | 4.89s    | **34x** |
| [microsoft/typescript](https://github.com/microsoft/typescript) | 47.4s                      | 2.10s    | **23x** |
| [typeorm/typeorm](https://github.com/typeorm/typeorm)           | 27.3s                      | 0.93s    | **29x** |
| [vuejs/core](https://github.com/vuejs/core)                     | 20.7s                      | 0.95s    | **22x** |

<details>

<summary>Detailed report</summary>

| microsoft/vscode |
| ---------------- |

```plaintext
Benchmark 1: eslint
  Time (mean ± σ):     167.840 s ±  1.233 s    [User: 211.852 s, System: 10.952 s]
  Range (min … max):   164.816 s … 169.410 s    10 runs

  Warning: Ignoring non-zero exit code.

Benchmark 2: tsgolint
  Time (mean ± σ):      4.897 s ±  0.153 s    [User: 64.479 s, System: 4.911 s]
  Range (min … max):    4.736 s …  5.183 s    10 runs

Summary
  tsgolint ran
   34.27 ± 1.10 times faster than eslint
```

| microsoft/typescript |
| -------------------- |

```plaintext
Benchmark 1: eslint
  Time (mean ± σ):     47.465 s ±  0.669 s    [User: 70.492 s, System: 4.250 s]
  Range (min … max):   46.636 s … 48.685 s    10 runs

  Warning: Ignoring non-zero exit code.

Benchmark 2: tsgolint
  Time (mean ± σ):      2.100 s ±  0.023 s    [User: 18.254 s, System: 1.448 s]
  Range (min … max):    2.068 s …  2.138 s    10 runs

Summary
  tsgolint ran
   22.60 ± 0.40 times faster than eslint
```

| typeorm/typeorm |
| --------------- |

```plaintext
Benchmark 1: eslint
  Time (mean ± σ):     27.294 s ±  0.504 s    [User: 42.467 s, System: 2.522 s]
  Range (min … max):   26.614 s … 28.522 s    10 runs

  Warning: Ignoring non-zero exit code.

Benchmark 2: tsgolint
  Time (mean ± σ):     930.0 ms ±   5.2 ms    [User: 9575.1 ms, System: 1041.6 ms]
  Range (min … max):   921.4 ms … 941.6 ms    10 runs

Summary
  tsgolint ran
   29.35 ± 0.57 times faster than eslint
```

| vuejs/core |
| ---------- |

```plaintext
Benchmark 1: eslint
  Time (mean ± σ):     20.680 s ±  0.364 s    [User: 35.617 s, System: 2.117 s]
  Range (min … max):   20.412 s … 21.604 s    10 runs

  Warning: Ignoring non-zero exit code.

Benchmark 2: tsgolint
  Time (mean ± σ):     955.5 ms ±  31.4 ms    [User: 11528.6 ms, System: 1001.1 ms]
  Range (min … max):   909.7 ms … 993.4 ms    10 runs

Summary
  tsgolint ran
   21.64 ± 0.81 times faster than eslint
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
