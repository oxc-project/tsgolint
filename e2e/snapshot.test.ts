import { execFileSync } from 'node:child_process';
import fs from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

import { glob } from 'fast-glob';
import { describe, expect, it } from 'vitest';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const ROOT_DIR = resolve(__dirname, '..');
const E2E_DIR = __dirname;
const FIXTURES_DIR = join(E2E_DIR, 'fixtures');
const TSGOLINT_BIN = join(ROOT_DIR, `tsgolint${process.platform === 'win32' ? '.exe' : ''}`);

const ALL_RULES = [
  'await-thenable',
  'consistent-return',
  'consistent-type-exports',
  'dot-notation',
  'no-array-delete',
  'no-base-to-string',
  'no-confusing-void-expression',
  'no-deprecated',
  'no-duplicate-type-constituents',
  'no-floating-promises',
  'no-for-in-array',
  'no-implied-eval',
  'no-meaningless-void-operator',
  'no-misused-promises',
  'no-misused-spread',
  'no-mixed-enums',
  'no-redundant-type-constituents',
  'no-unnecessary-boolean-literal-compare',
  'no-unnecessary-condition',
  'no-unnecessary-qualifier',
  'no-unnecessary-template-expression',
  'no-unnecessary-type-conversion',
  'no-unnecessary-type-arguments',
  'no-unnecessary-type-parameters',
  'no-unnecessary-type-assertion',
  'no-useless-default-assignment',
  'no-unsafe-argument',
  'no-unsafe-assignment',
  'no-unsafe-call',
  'no-unsafe-enum-comparison',
  'no-unsafe-member-access',
  'no-unsafe-return',
  'no-unsafe-type-assertion',
  'no-unsafe-unary-minus',
  'non-nullable-type-assertion-style',
  'only-throw-error',
  'prefer-find',
  'prefer-includes',
  'prefer-nullish-coalescing',
  'prefer-optional-chain',
  'prefer-promise-reject-errors',
  'prefer-regexp-exec',
  'prefer-readonly',
  'prefer-readonly-parameter-types',
  'prefer-reduce-type-parameter',
  'prefer-return-this-type',
  'prefer-string-starts-ends-with',
  'promise-function-async',
  'related-getter-setter-pairs',
  'require-array-sort-compare',
  'require-await',
  'restrict-plus-operands',
  'restrict-template-expressions',
  'return-await',
  'strict-boolean-expressions',
  'strict-void-return',
  'switch-exhaustiveness-check',
  'unbound-method',
  'use-unknown-in-catch-callback-variable',
] as const;

enum DiagnosticKind {
  Rule,
  Internal,
}

interface BaseDiagnostic {
  kind: DiagnosticKind;
  file_path: string;
  range: {
    pos: number;
    end: number;
  };
  message: {
    id: string;
    description: string;
    help?: string;
  };
}

interface RuleDiagnostic extends BaseDiagnostic {
  kind: DiagnosticKind.Rule;
  rule: string;
  fixes?: Array<{
    text: string;
    range: { pos: number; end: number };
  }>;
  suggestions?: {
    message: { id: string; description: string; help?: string };
    fixes: Array<{
      text: string;
      range: { pos: number; end: number };
    }>;
  }[];
}

interface InternalDiagnostic extends BaseDiagnostic {
  kind: DiagnosticKind.Internal;
}

type Diagnostic = RuleDiagnostic | InternalDiagnostic;

function parseHeadlessOutput(data: Buffer): Diagnostic[] {
  let offset = 0;
  const diagnostics: Diagnostic[] = [];

  while (offset < data.length) {
    if (offset + 5 > data.length) {
      break;
    }

    // Read header: 4 bytes length + 1 byte message type
    const length = data.readUInt32LE(offset);
    const msgType = data[offset + 4];
    offset += 5;

    if (offset + length > data.length) {
      break;
    }

    // Read payload
    const payload = data.subarray(offset, offset + length);
    offset += length;

    // Only process diagnostic messages (type 1)
    if (msgType === 1) {
      try {
        const diagnostic = JSON.parse(payload.toString('utf-8'));
        // Make file paths relative to fixtures/ for deterministic snapshots
        const filePath = diagnostic.file_path || '';
        if (filePath.includes('fixtures/')) {
          diagnostic.file_path = 'fixtures/' + filePath.split('fixtures/').pop();
        }
        diagnostics.push(diagnostic);
      } catch {
        continue;
      }
    }
  }

  return diagnostics;
}

function sortDiagnostics(diagnostics: Diagnostic[]): Diagnostic[] {
  diagnostics.sort((a, b) => {
    const aFilePath = a.file_path || '';
    const bFilePath = b.file_path || '';
    if (aFilePath !== bFilePath) return aFilePath.localeCompare(bFilePath);

    const aRule = 'rule' in a ? a.rule || '' : '';
    const bRule = 'rule' in b ? b.rule || '' : '';
    if (aRule !== bRule) return aRule.localeCompare(bRule);

    const aPos = (a.range && a.range.pos) || 0;
    const bPos = (b.range && b.range.pos) || 0;
    if (aPos !== bPos) return aPos - bPos;

    const aEnd = (a.range && a.range.end) || 0;
    const bEnd = (b.range && b.range.end) || 0;
    return aEnd - bEnd;
  });

  return diagnostics;
}

async function getTestFiles(testPath: string): Promise<string[]> {
  const patterns = [`${testPath}/**/*.ts`, `${testPath}/**/*.tsx`, `${testPath}/**/*.mts`, `${testPath}/**/*.cts`];
  const allFiles: string[] = [];

  for (const pattern of patterns) {
    const files = await glob(pattern, {
      cwd: FIXTURES_DIR,
      absolute: true,
      ignore: ['**/node_modules/**', '**/*.json'],
    });
    allFiles.push(...files);
  }

  return allFiles;
}

function resolveTestFilePath(relativePath: string): string {
  return join(FIXTURES_DIR, relativePath);
}

function getDiagnostics(config: string): Diagnostic[] {
  const output = execFileSync(TSGOLINT_BIN, ['headless'], {
    input: config,
    env: { ...process.env, GOMAXPROCS: '1' },
  });

  return sortDiagnostics(parseHeadlessOutput(output));
}

function generateConfig(
  files: string[],
  rules:
    readonly ((typeof ALL_RULES)[number] | { name: (typeof ALL_RULES)[number]; options: Record<string, unknown> })[] =
      ALL_RULES,
  options?: {
    reportSyntactic?: boolean;
    reportSemantic?: boolean;
  },
): string {
  // See the headless payload schema in `cmd/tsgolint/payload.go`.
  const config = {
    version: 2,
    configs: [
      {
        file_paths: files,
        rules: rules.map(
          (
            r,
          ): {
            name: (typeof ALL_RULES)[number];
            options?: Record<string, unknown>;
          } => (typeof r === 'string' ? { name: r } : r),
        ),
      },
    ],
    ...(options?.reportSyntactic !== undefined && { report_syntactic: options.reportSyntactic }),
    ...(options?.reportSemantic !== undefined && { report_semantic: options.reportSemantic }),
  } as const;
  return JSON.stringify(config);
}

describe('TSGoLint E2E Snapshot Tests', () => {
  it('(`ALL_RULES`) should include every available rule ', async () => {
    const rulesDir = join(import.meta.dirname, '..', 'internal', 'rules');
    const fileSystemRulesList = [];

    // read all folders in the directory
    for (const entry of await fs.readdir(rulesDir)) {
      if (entry.startsWith('.') || entry === 'fixtures') {
        continue;
      }
      const entryPath = join(rulesDir, entry);
      const stat = await fs.stat(entryPath);
      if (!stat.isDirectory()) continue;
      const ruleFileStat = await fs.stat(join(entryPath, `${entry}.go`)).catch(() => null);
      if (ruleFileStat?.isFile()) {
        fileSystemRulesList.push(entry.replace(/_/g, '-'));
      }
    }

    expect(fileSystemRulesList.sort()).toEqual([...ALL_RULES].sort());
  });

  it('should generate consistent diagnostics snapshot', async () => {
    const testFiles = await getTestFiles('basic');
    expect(testFiles.length).toBeGreaterThan(0);

    const config = generateConfig(testFiles);

    // Run tsgolint in headless mode with single thread for deterministic results
    // Set GOMAXPROCS=1 for single-threaded execution
    const env = { ...process.env, GOMAXPROCS: '1' };

    let output: Buffer;
    output = execFileSync(TSGOLINT_BIN, ['headless', '-fix', '-fix-suggestions'], {
      input: config,
      env,
    });

    let diagnostics = parseHeadlessOutput(output);
    diagnostics = sortDiagnostics(diagnostics);

    expect(diagnostics.length).toBeGreaterThan(0);

    expect(diagnostics).toMatchSnapshot();
  });

  it('supports passing rule config', async () => {
    const testFile = resolveTestFilePath('basic/rules/no-floating-promises/void.ts');
    const config = (ignoreVoid: boolean) => ({
      version: 2,
      configs: [
        {
          file_paths: [testFile],
          rules: [
            {
              name: 'no-floating-promises',
              options: { ignoreVoid },
            },
          ],
        },
      ],
    });

    let output: Buffer;
    output = execFileSync(TSGOLINT_BIN, ['headless'], {
      input: JSON.stringify(config(false)),
    });

    let diagnostics = parseHeadlessOutput(output);
    diagnostics = sortDiagnostics(diagnostics);

    expect(diagnostics.length).toBeGreaterThan(0);
    expect(diagnostics).toMatchSnapshot();

    // Re-run with ignoreVoid=true, should have no diagnostics
    output = execFileSync(TSGOLINT_BIN, ['headless'], {
      input: JSON.stringify(config(true)),
    });

    diagnostics = parseHeadlessOutput(output);
    expect(diagnostics.length).toBe(0);
  });

  it.runIf(process.platform === 'win32')(
    'should not panic with mixed forward/backslash paths from Rust (issue #143)',
    async () => {
      // Regression test for https://github.com/oxc-project/tsgolint/issues/143
      // This test reproduces the issue where Rust sends paths with backslashes
      // but TypeScript program has forward slashes, causing:
      // "panic: Expected file 'E:\oxc\...\index.ts' to be in program"

      const testFile = join(FIXTURES_DIR, 'basic', 'rules', 'no-floating-promises', 'index.ts');

      // On Windows, convert forward slashes to backslashes to simulate Rust input
      const rustStylePath = testFile.replace(/\//g, '\\');

      expect(() => {
        execFileSync(TSGOLINT_BIN, ['headless'], {
          input: generateConfig([rustStylePath], ['no-floating-promises']),
          env: { ...process.env, GOMAXPROCS: '1' },
        });
      }).not.toThrow();
    },
  );

  it('should correctly evaluate project references', async () => {
    const testFiles = await getTestFiles('project-references');
    expect(testFiles.length).toBeGreaterThan(0);

    const config = generateConfig(testFiles, ['no-unsafe-argument']);

    // Run tsgolint in headless mode with single thread for deterministic results
    // Set GOMAXPROCS=1 for single-threaded execution
    let diagnostics = getDiagnostics(config);

    expect(diagnostics).toMatchSnapshot();
  });

  it('should correctly lint files not inside a project', async () => {
    const testFiles = await getTestFiles('with-unmatched-files');
    expect(testFiles.length).toBeGreaterThan(0);

    const config = generateConfig(testFiles, ['no-unsafe-argument']);

    let diagnostics = getDiagnostics(config);

    expect(diagnostics).toMatchSnapshot();
  });

  it('should work with the old version of the headless payload', async () => {
    function generateV1HeadlessPayload(
      files: string[],
      rules: readonly (typeof ALL_RULES)[number][] = ALL_RULES,
    ): string {
      const config = {
        files: files.map((filePath) => ({
          file_path: filePath,
          rules,
        })),
      };
      return JSON.stringify(config);
    }

    const testFiles = await getTestFiles('basic');
    expect(testFiles.length).toBeGreaterThan(0);

    const v1Config = generateV1HeadlessPayload(testFiles);
    const v1Diagnostics = getDiagnostics(v1Config);

    const v2Config = generateConfig(testFiles);
    const v2Diagnostics = getDiagnostics(v2Config);

    expect(v1Diagnostics).toStrictEqual(v2Diagnostics);
  });

  it('should use source overrides instead of reading from disk', async () => {
    const testFiles = await getTestFiles('source-overrides');
    expect(testFiles.length).toBeGreaterThan(0);
    const testFile = testFiles[0];

    const overriddenContent = `const promise = new Promise((resolve, _reject) => resolve("value"));
promise;
`;

    const config = {
      version: 2,
      configs: [
        {
          file_paths: [testFile],
          rules: [{ name: 'no-floating-promises' }],
        },
      ],
      source_overrides: {
        [testFile]: overriddenContent,
      },
    };

    let diagnostics = getDiagnostics(JSON.stringify(config));

    expect(diagnostics.length).toBe(1);
    expect(diagnostics[0].kind == DiagnosticKind.Rule && diagnostics[0].rule).toBe('no-floating-promises');
    expect(diagnostics[0].file_path).toContain('original.ts');
  });

  it('should not report errors when source override is valid', async () => {
    const testFiles = await getTestFiles('source-overrides');
    expect(testFiles.length).toBeGreaterThan(0);
    const testFile = testFiles[0];

    const validOverride = `// Valid code with no errors
const x: number = 42;
console.log(x);
`;

    const config = {
      version: 2,
      configs: [
        {
          file_paths: [testFile],
          rules: [{ name: 'no-floating-promises' }, { name: 'no-unsafe-assignment' }],
        },
      ],
      source_overrides: {
        [testFile]: validOverride,
      },
    };

    const diagnostics = getDiagnostics(JSON.stringify(config));

    expect(diagnostics.length).toBe(0);
  });

  it.for([false, true])('should lint differently cased editor paths (inferred: %s)', async (inferred, context) => {
    const directory = await fs.mkdtemp(join(tmpdir(), 'tsgolint-casing-'));
    try {
      const file = join(directory, 'original.ts');
      const editorFile = join(directory, 'Original.ts');
      await fs.writeFile(file, 'const value = 1;');
      if (!(await fs.stat(editorFile).catch(() => undefined))) {
        context.skip(); // This regression requires a case-insensitive filesystem.
        return;
      }
      if (inferred) {
        // An unsaved editor file exists only in the source overrides.
        await fs.rm(file);
      } else {
        await fs.writeFile(join(directory, 'tsconfig.json'), JSON.stringify({ files: ['original.ts'] }));
      }

      const output = execFileSync(TSGOLINT_BIN, ['headless'], {
        cwd: directory,
        input: JSON.stringify({
          version: 2,
          configs: [{ file_paths: [editorFile], rules: [{ name: 'no-unsafe-assignment' }] }],
          source_overrides: { [file]: 'declare const value: any; const result = value;' },
        }),
      });
      const diagnostics = parseHeadlessOutput(output);
      expect(diagnostics).toHaveLength(1);
      expect(diagnostics[0].kind === DiagnosticKind.Rule && diagnostics[0].rule).toBe('no-unsafe-assignment');
    } finally {
      await fs.rm(directory, { recursive: true, force: true });
    }
  });

  it('should handle tsconfig diagnostics when TypeScript reports them', async () => {
    const testFiles = await getTestFiles('with-invalid-tsconfig-option');
    expect(testFiles.length).toBeGreaterThan(0);

    const config = generateConfig(testFiles, ['no-floating-promises']);

    let diagnostics = getDiagnostics(config);

    expect(diagnostics).toMatchSnapshot();
  });

  it('should report an error if the tsconfig.json could not be parsed', async () => {
    const testFiles = await getTestFiles('with-invalid-tsconfig-json');
    expect(testFiles.length).toBeGreaterThan(0);

    const config = generateConfig(testFiles, ['no-floating-promises']);

    let diagnostics = getDiagnostics(config);

    expect(diagnostics).toMatchSnapshot();
  });

  it('should suppress tsconfig-error diagnostics when OXLINT_TSGOLINT_DANGEROUSLY_SUPPRESS_PROGRAM_DIAGNOSTICS env var is true', async () => {
    const testFiles = await getTestFiles('with-invalid-tsconfig-option');
    expect(testFiles.length).toBeGreaterThan(0);

    const config = generateConfig(testFiles, ['no-floating-promises']);

    const env = { ...process.env, GOMAXPROCS: '1', OXLINT_TSGOLINT_DANGEROUSLY_SUPPRESS_PROGRAM_DIAGNOSTICS: 'true' };

    const output = execFileSync(TSGOLINT_BIN, ['headless'], {
      input: config,
      env,
    });

    const diagnostics = parseHeadlessOutput(output);

    expect(diagnostics.length).toBe(0);
  });

  it('should still report lint rule diagnostics when OXLINT_TSGOLINT_DANGEROUSLY_SUPPRESS_PROGRAM_DIAGNOSTICS env var is true', async () => {
    const testFiles = await getTestFiles('with-invalid-tsconfig-option-and-lint-errors');
    expect(testFiles.length).toBeGreaterThan(0);

    const config = generateConfig(testFiles, ['no-floating-promises']);

    const env = { ...process.env, GOMAXPROCS: '1', OXLINT_TSGOLINT_DANGEROUSLY_SUPPRESS_PROGRAM_DIAGNOSTICS: 'true' };

    const output = execFileSync(TSGOLINT_BIN, ['headless'], {
      input: config,
      env,
    });

    const diagnostics = parseHeadlessOutput(output);

    // tsconfig-error diagnostics should be suppressed, but lint rules should still fire
    const tsconfigErrors = diagnostics.filter((d: any) => d.kind === DiagnosticKind.Internal);
    const ruleDiagnostics = diagnostics.filter((d: any) => d.kind === DiagnosticKind.Rule);

    expect(tsconfigErrors.length).toBe(0);
    expect(ruleDiagnostics).toMatchSnapshot();
  });

  it('should work correctly with nested module namespaces and parent module searches (`ValueMatchesSomeSpecifier`) (issue #135)', async () => {
    const testFiles = await getTestFiles('issue-135');
    expect(testFiles.length).toBeGreaterThan(0);

    const config = generateConfig(testFiles, [
      {
        name: 'no-floating-promises',
        options: {
          allowForKnownSafeCalls: [
            {
              from: 'package',
              name: ['test', 'it', 'suite', 'describe'],
              package: 'node2:test',
            },
          ],
        },
      },
    ]);

    let diagnostics = getDiagnostics(config);

    expect(diagnostics).toMatchSnapshot();
  });

  it('should keep no-floating-promises suggestion messages distinct (issue #534)', async () => {
    const testFiles = await getTestFiles('issue-534');
    expect(testFiles.length).toBeGreaterThan(0);

    const config = generateConfig(testFiles, [
      {
        name: 'no-floating-promises',
        options: { ignoreVoid: true },
      },
    ]);

    const output = execFileSync(TSGOLINT_BIN, ['headless', '-fix-suggestions'], {
      input: config,
      env: { ...process.env, GOMAXPROCS: '1' },
    });

    let diagnostics = parseHeadlessOutput(output);
    diagnostics = sortDiagnostics(diagnostics);

    expect(diagnostics).toHaveLength(1);
    expect((diagnostics[0] as RuleDiagnostic).suggestions).toStrictEqual(
      [
        {
          fixes: [{ range: expect.any(Object), text: 'void ' }],
          message: { description: 'Add void operator to ignore.', id: 'floatingFixVoid' },
        },
        {
          fixes: [{ range: expect.any(Object), text: 'await ' }],
          message: { description: 'Add await operator.', id: 'floatingFixAwait' },
        },
      ],
    );
  });

  it('should use strict defaults when strict is omitted (issue #762)', async () => {
    const testFiles = await getTestFiles('issue-762');
    expect(testFiles.length).toBeGreaterThan(0);

    const config = generateConfig(testFiles, ['no-unnecessary-boolean-literal-compare']);

    let diagnostics = getDiagnostics(config);

    expect(diagnostics).toHaveLength(0);
  });

  it('should use ancestor tsconfig when nearest tsconfig excludes the file (voidzero-dev/vite-plus#1443)', async () => {
    const testFile = resolveTestFilePath('issue-vp-1443/packages/cli/src/utils/terminal.ts');
    const config = generateConfig([testFile], ['no-floating-promises'], {
      reportSemantic: true,
    });

    const diagnostics = getDiagnostics(config);

    expect(diagnostics).toStrictEqual([]);
  });

  it('should report type errors', async () => {
    const testFiles = await getTestFiles('report-type-errors');
    expect(testFiles.length).toBeGreaterThan(0);

    const config = generateConfig(testFiles, ['no-floating-promises'], {
      reportSemantic: true,
    });

    let diagnostics = getDiagnostics(config);

    expect(diagnostics).toMatchSnapshot();
  });

  it('should handle circular project references (issue #297)', async () => {
    // Regression test for https://github.com/oxc-project/tsgolint/issues/297
    // This test reproduces the issue where circular tsconfig references
    // (project1 -> project2 -> project1) caused a panic:
    // "panic: Expected file '...project2/src/demo/index.ts' to be in program '...project2/tsconfig.json'"
    const testFiles = await getTestFiles('circular-project-references');
    expect(testFiles.length).toBeGreaterThan(0);

    const config = generateConfig(testFiles, ALL_RULES, {
      reportSemantic: true,
      reportSyntactic: true,
    });

    let diagnostics = getDiagnostics(config);

    expect(diagnostics).toMatchSnapshot();
  });

  it('should report syntactic issues', async () => {
    const testFiles = await getTestFiles('semantic-syntactic-diagnostics');
    expect(testFiles.length).toBeGreaterThan(0);

    const config = generateConfig(testFiles, ALL_RULES, {
      reportSemantic: true,
      reportSyntactic: true,
    });

    let diagnostics = getDiagnostics(config);

    expect(diagnostics).toMatchSnapshot();
  });

  describe('per-config `type_check`', () => {
    const SYNTACTIC_FILE = 'semantic-syntactic-diagnostics/src/a.ts';
    const SEMANTIC_FILE = 'semantic-syntactic-diagnostics/src/b.mts';
    const TYPE_ERROR_FILES = [`fixtures/${SYNTACTIC_FILE}`, `fixtures/${SEMANTIC_FILE}`];

    // Both files belong to the same tsconfig, so they are linted by a single program.
    // The options are keyed by fixture file, not by kind of diagnostic to report.
    function runWithTypeCheck(typeCheckByFile: { 'src/a.ts'?: boolean; 'src/b.mts'?: boolean }): Diagnostic[] {
      const configFor = (file: string, typeCheck: boolean | undefined) => ({
        file_paths: [resolveTestFilePath(file)],
        rules: [{ name: 'no-floating-promises' }],
        ...(typeCheck !== undefined && { type_check: typeCheck }),
      });

      return getDiagnostics(JSON.stringify({
        version: 2,
        configs: [
          configFor(SYNTACTIC_FILE, typeCheckByFile['src/a.ts']),
          configFor(SEMANTIC_FILE, typeCheckByFile['src/b.mts']),
        ],
        report_semantic: true,
        report_syntactic: true,
      }));
    }

    // Diagnostics of the fixture's two source files: internal diagnostics of
    // other files are tsconfig diagnostics, which `type_check` doesn't cover
    // and which can carry a TS id and a range too.
    function typeErrorFiles(diagnostics: Diagnostic[]): string[] {
      const files = diagnostics
        .filter((d) => d.kind === DiagnosticKind.Internal && TYPE_ERROR_FILES.includes(d.file_path))
        .map((d) => d.file_path);
      return [...new Set(files)].sort();
    }

    // Rule 1 of the `type_check` semantics documented in cmd/tsgolint/payload.go.
    it('reports type errors for every file when no config group carries `type_check`', () => {
      expect(typeErrorFiles(runWithTypeCheck({}))).toStrictEqual([...TYPE_ERROR_FILES].sort());
    });

    it('skips the files of a config group setting `type_check: false`', () => {
      expect(typeErrorFiles(runWithTypeCheck({ 'src/b.mts': false }))).toStrictEqual([`fixtures/${SYNTACTIC_FILE}`]);
    });

    it('reports only the opted-in config group when every group carries the flag', () => {
      expect(typeErrorFiles(runWithTypeCheck({ 'src/a.ts': false, 'src/b.mts': true }))).toStrictEqual([
        `fixtures/${SEMANTIC_FILE}`,
      ]);
    });

    it('reports no type errors when every config group sets `type_check: false`', () => {
      expect(typeErrorFiles(runWithTypeCheck({ 'src/a.ts': false, 'src/b.mts': false }))).toStrictEqual([]);
    });

    // Rule 3 of the `type_check` semantics documented in cmd/tsgolint/payload.go:
    // `false` excludes every TypeScript diagnostic attached to the file, while
    // the type-aware lint rules still run on it.
    it('rule 3: a `type_check: false` file reports no type error but still runs the rules', () => {
      const testFile = resolveTestFilePath('type-check-and-rules/index.ts');
      const runOn = (typeCheck: boolean) =>
        getDiagnostics(JSON.stringify({
          version: 2,
          configs: [
            {
              file_paths: [testFile],
              rules: [{ name: 'no-floating-promises' }],
              type_check: typeCheck,
            },
          ],
          report_semantic: true,
        }));

      // The fixture has both a type error and a floating promise.
      const reported = runOn(true);
      expect(reported.filter((d) => d.kind === DiagnosticKind.Rule).map((d) => (d as RuleDiagnostic).rule))
        .toStrictEqual(['no-floating-promises']);
      expect(reported.filter((d) => d.kind === DiagnosticKind.Internal).map((d) => d.message.id))
        .toStrictEqual(['TS2322']);

      // Opting out only drops the type error.
      const skipped = runOn(false);
      expect(skipped.filter((d) => d.kind === DiagnosticKind.Rule).map((d) => (d as RuleDiagnostic).rule))
        .toStrictEqual(['no-floating-promises']);
      expect(skipped.filter((d) => d.kind === DiagnosticKind.Internal)).toStrictEqual([]);
    });

    // Rule 4 of the `type_check` semantics documented in cmd/tsgolint/payload.go:
    // `report_syntactic`/`report_semantic` stay the run-wide kind selectors, so
    // opting a file in reports nothing more without them.
    it('rule 4: `type_check: true` is a no-op without a kind selector', () => {
      const testFile = resolveTestFilePath('type-check-and-rules/index.ts');
      const diagnostics = getDiagnostics(JSON.stringify({
        version: 2,
        configs: [
          {
            file_paths: [testFile],
            rules: [{ name: 'no-floating-promises' }],
            type_check: true,
          },
        ],
      }));

      expect(diagnostics.filter((d) => d.kind === DiagnosticKind.Rule).map((d) => (d as RuleDiagnostic).rule))
        .toStrictEqual(['no-floating-promises']);
      expect(diagnostics.filter((d) => d.kind === DiagnosticKind.Internal)).toStrictEqual([]);
    });

    // Rule 2 of the `type_check` semantics documented in cmd/tsgolint/payload.go,
    // rules half: the last config group listing a file defines its rules too.
    it("rule 2: a file listed twice is linted with the last config group's rules", () => {
      const testFile = resolveTestFilePath('type-check-and-rules/index.ts');
      const rulesReportedFor = (rules: string[][]) =>
        getDiagnostics(JSON.stringify({
          version: 2,
          configs: rules.map((names) => ({
            file_paths: [testFile],
            rules: names.map((name) => ({ name })),
          })),
        })).filter((d) => d.kind === DiagnosticKind.Rule).map((d) => (d as RuleDiagnostic).rule);

      // The fixture violates both rules, so only the last group's rule reports.
      expect(rulesReportedFor([['no-unsafe-argument'], ['no-floating-promises']]))
        .toStrictEqual(['no-floating-promises']);
      expect(rulesReportedFor([['no-floating-promises'], ['no-unsafe-argument']]))
        .toStrictEqual(['no-unsafe-argument']);
    });

    // Rule 3 of the `type_check` semantics documented in cmd/tsgolint/payload.go:
    // the program-level tsconfig diagnostics belong to no file in particular, so
    // opting a file out doesn't silence them.
    it('rule 3: a `type_check: false` file still reports its tsconfig diagnostics', () => {
      const testFile = resolveTestFilePath('with-invalid-tsconfig-option/index.ts');
      const diagnostics = getDiagnostics(JSON.stringify({
        version: 2,
        configs: [
          {
            file_paths: [testFile],
            rules: [{ name: 'no-floating-promises' }],
            type_check: false,
          },
        ],
        report_semantic: true,
      }));

      expect(diagnostics.filter((d) => d.kind === DiagnosticKind.Internal).map((d) => d.message.id))
        .toStrictEqual(['tsconfig-error']);
    });

    // The files that match no tsconfig are linted through an inferred program,
    // which `type_check` has to scope like any other.
    it('skips the type errors of a `type_check: false` file outside any project', () => {
      const testFile = resolveTestFilePath('with-unmatched-files/test.ts');
      const typeErrorsOf = (typeCheck: boolean) =>
        getDiagnostics(JSON.stringify({
          version: 2,
          configs: [
            {
              file_paths: [testFile],
              rules: [{ name: 'no-unsafe-argument' }],
              type_check: typeCheck,
            },
          ],
          report_semantic: true,
        })).filter((d) => d.kind === DiagnosticKind.Internal).map((d) => d.message.id);

      expect(typeErrorsOf(true)).toStrictEqual(['TS2322']);
      expect(typeErrorsOf(false)).toStrictEqual([]);
    });
  });

  it('lints a file listed with a path relative to the working directory', () => {
    // The tsconfig lives in the subdirectory the file is in, so a relative path
    // that reached the linter would be resolved against the wrong directory and
    // miss the program.
    const directory = resolveTestFilePath('relative-file-paths');

    const output = execFileSync(TSGOLINT_BIN, ['headless'], {
      cwd: directory,
      input: JSON.stringify({
        version: 2,
        configs: [{ file_paths: ['src/index.ts'], rules: [{ name: 'no-floating-promises' }] }],
      }),
      env: { ...process.env, GOMAXPROCS: '1' },
    });

    const diagnostics = sortDiagnostics(parseHeadlessOutput(output));

    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0].kind === DiagnosticKind.Rule && diagnostics[0].rule).toBe('no-floating-promises');
    expect(diagnostics[0].file_path).toBe('fixtures/relative-file-paths/src/index.ts');
  });

  it('should attach the tsconfig path to tsconfig-error diagnostics without a TypeScript diagnostic file (oxc-project/oxc#2200)', async () => {
    const testFiles = await getTestFiles('issue-oxc-2200');
    expect(testFiles.length).toBeGreaterThan(0);

    const config = generateConfig(testFiles, ['no-floating-promises']);

    let diagnostics = getDiagnostics(config);

    expect(diagnostics).toMatchSnapshot();
  });
});
