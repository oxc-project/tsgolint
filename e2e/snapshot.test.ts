import { execFileSync } from 'node:child_process';
import fs from 'node:fs/promises';
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
  'no-unnecessary-template-expression',
  'no-unnecessary-type-arguments',
  'no-unnecessary-type-assertion',
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
  'prefer-includes',
  'prefer-optional-chain',
  'prefer-promise-reject-errors',
  'prefer-reduce-type-parameter',
  'prefer-return-this-type',
  'promise-function-async',
  'related-getter-setter-pairs',
  'require-array-sort-compare',
  'require-await',
  'restrict-plus-operands',
  'restrict-template-expressions',
  'return-await',
  'strict-boolean-expressions',
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

function generateConfig(
  files: string[],
  rules:
    readonly ((typeof ALL_RULES)[number] | { name: typeof ALL_RULES[number]; options: Record<string, unknown> })[] =
      ALL_RULES,
  options?: {
    reportSyntactic?: boolean;
    reportSemantic?: boolean;
  },
): string {
  // Headless payload format:
  // ```json
  // {
  //   "configs": [
  //     {
  //       "file_paths": ["/abs/path/a.ts", ...],
  //       "rules": [ { "name": "rule-a" }, { "name": "rule-b" } ]
  //     }
  //   ]
  // }
  // ```
  const config = {
    version: 2,
    configs: [
      {
        file_paths: files,
        rules: rules.map((r): {
          name: typeof ALL_RULES[number];
          options?: Record<string, unknown>;
        } => (typeof r === 'string' ? { name: r } : r)),
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

    expect(fileSystemRulesList.sort()).toEqual(
      [...ALL_RULES].sort(),
    );
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
    const env = { ...process.env, GOMAXPROCS: '1' };

    const output = execFileSync(TSGOLINT_BIN, ['headless'], {
      input: config,
      env,
    });

    let diagnostics = parseHeadlessOutput(output);
    diagnostics = sortDiagnostics(diagnostics);

    expect(diagnostics).toMatchSnapshot();
  });

  it('should correctly lint files not inside a project', async () => {
    const testFiles = await getTestFiles('with-unmatched-files');
    expect(testFiles.length).toBeGreaterThan(0);

    const config = generateConfig(testFiles, ['no-unsafe-argument']);

    const env = { ...process.env, GOMAXPROCS: '1' };

    const output = execFileSync(TSGOLINT_BIN, ['headless'], {
      input: config,
      env,
    });

    let diagnostics = parseHeadlessOutput(output);
    diagnostics = sortDiagnostics(diagnostics);

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

    function getDiagnostics(config: string): Diagnostic[] {
      let output: Buffer;
      output = execFileSync(TSGOLINT_BIN, ['headless'], {
        input: config,
        env: { ...process.env, GOMAXPROCS: '1' },
      });

      const diagnostics = parseHeadlessOutput(output);
      return sortDiagnostics(diagnostics);
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

    const env = { ...process.env, GOMAXPROCS: '1' };
    const output = execFileSync(TSGOLINT_BIN, ['headless'], {
      input: JSON.stringify(config),
      env,
    });

    let diagnostics = parseHeadlessOutput(output);
    diagnostics = sortDiagnostics(diagnostics);

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

    const env = { ...process.env, GOMAXPROCS: '1' };
    const output = execFileSync(TSGOLINT_BIN, ['headless'], {
      input: JSON.stringify(config),
      env,
    });

    const diagnostics = parseHeadlessOutput(output);

    expect(diagnostics.length).toBe(0);
  });

  it('should handle tsconfig diagnostics when TypeScript reports them', async () => {
    const testFiles = await getTestFiles('with-invalid-tsconfig-option');
    expect(testFiles.length).toBeGreaterThan(0);

    const config = generateConfig(testFiles, ['no-floating-promises']);

    const env = { ...process.env, GOMAXPROCS: '1' };

    const output = execFileSync(TSGOLINT_BIN, ['headless'], {
      input: config,
      env,
    });

    let diagnostics = parseHeadlessOutput(output);
    diagnostics = sortDiagnostics(diagnostics);

    expect(diagnostics).toMatchSnapshot();
  });

  it('should report an error if the tsconfig.json could not be parsed', async () => {
    const testFiles = await getTestFiles('with-invalid-tsconfig-json');
    expect(testFiles.length).toBeGreaterThan(0);

    const config = generateConfig(testFiles, ['no-floating-promises']);

    const env = { ...process.env, GOMAXPROCS: '1' };

    const output = execFileSync(TSGOLINT_BIN, ['headless'], {
      input: config,
      env,
    });

    let diagnostics = parseHeadlessOutput(output);
    diagnostics = sortDiagnostics(diagnostics);

    expect(diagnostics).toMatchSnapshot();
  });

  it('should work correctly with nested module namespaces and parent module searches (`ValueMatchesSomeSpecifier`) (issue #135)', async () => {
    const testFiles = await getTestFiles('issue-135');
    expect(testFiles.length).toBeGreaterThan(0);

    const config = generateConfig(testFiles, [{
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
    }]);

    const output = execFileSync(TSGOLINT_BIN, ['headless'], {
      input: config,
      env: { ...process.env, GOMAXPROCS: '1' },
    });

    let diagnostics = parseHeadlessOutput(output);
    diagnostics = sortDiagnostics(diagnostics);

    expect(diagnostics).toMatchSnapshot();
  });

  it('should report type errors', async () => {
    const testFiles = await getTestFiles('report-type-errors');
    expect(testFiles.length).toBeGreaterThan(0);

    const config = generateConfig(
      testFiles,
      ['no-floating-promises'],
      {
        reportSemantic: true,
      },
    );

    const output = execFileSync(TSGOLINT_BIN, ['headless'], {
      input: config,
      env: { ...process.env, GOMAXPROCS: '1' },
    });

    let diagnostics = parseHeadlessOutput(output);
    diagnostics = sortDiagnostics(diagnostics);

    expect(diagnostics).toMatchSnapshot();
  });
});
