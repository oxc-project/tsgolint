import { execFileSync } from 'node:child_process';
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
  'switch-exhaustiveness-check',
  'unbound-method',
  'use-unknown-in-catch-callback-variable',
] as const;

interface Diagnostic {
  file_path?: string;
  rule?: string;
  range?: {
    pos: number;
    end: number;
  };
  [key: string]: any;
}

interface TypeScriptDiagnostic {
  file_path?: string;
  code?: number;
  category?: string;
  message?: string;
  range?: {
    pos: number;
    end: number;
  };
  [key: string]: any;
}

function parseHeadlessOutput(data: Buffer): { diagnostics: Diagnostic[]; typeDiagnostics: TypeScriptDiagnostic[] } {
  let offset = 0;
  const diagnostics: Diagnostic[] = [];
  const typeDiagnostics: TypeScriptDiagnostic[] = [];

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

    // Process diagnostic messages (type 1) and TypeScript diagnostic messages (type 2)
    if (msgType === 1 || msgType === 2) {
      try {
        const diagnostic = JSON.parse(payload.toString('utf-8'));
        // Make file paths relative to fixtures/ for deterministic snapshots
        const filePath = diagnostic.file_path || '';
        if (filePath.includes('fixtures/')) {
          diagnostic.file_path = 'fixtures/' + filePath.split('fixtures/').pop();
        }

        if (msgType === 1) {
          diagnostics.push(diagnostic);
        } else if (msgType === 2) {
          typeDiagnostics.push(diagnostic);
        }
      } catch {
        continue;
      }
    }
  }

  return { diagnostics, typeDiagnostics };
}

function sortDiagnostics(diagnostics: Diagnostic[]): Diagnostic[] {
  diagnostics.sort((a, b) => {
    const aFilePath = a.file_path || '';
    const bFilePath = b.file_path || '';
    if (aFilePath !== bFilePath) return aFilePath.localeCompare(bFilePath);

    const aRule = a.rule || '';
    const bRule = b.rule || '';
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

function sortTypeScriptDiagnostics(diagnostics: TypeScriptDiagnostic[]): TypeScriptDiagnostic[] {
  diagnostics.sort((a, b) => {
    const aFilePath = a.file_path || '';
    const bFilePath = b.file_path || '';
    if (aFilePath !== bFilePath) return aFilePath.localeCompare(bFilePath);

    const aCode = a.code || 0;
    const bCode = b.code || 0;
    if (aCode !== bCode) return aCode - bCode;

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

function generateConfig(files: string[], rules: readonly (typeof ALL_RULES)[number][] = ALL_RULES): string {
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
        rules: rules.map((r) => ({ name: r })),
      },
    ],
  } as const;
  return JSON.stringify(config);
}

describe('TSGoLint E2E Snapshot Tests', () => {
  it('should generate consistent diagnostics snapshot', async () => {
    const testFiles = await getTestFiles('basic');
    expect(testFiles.length).toBeGreaterThan(0);

    const config = generateConfig(testFiles);

    // Run tsgolint in headless mode with single thread for deterministic results
    // Set GOMAXPROCS=1 for single-threaded execution
    const env = { ...process.env, GOMAXPROCS: '1' };

    let output: Buffer;
    output = execFileSync(TSGOLINT_BIN, ['headless'], {
      input: config,
      env,
    });

    const { diagnostics, typeDiagnostics } = parseHeadlessOutput(output);
    const sortedDiagnostics = sortDiagnostics(diagnostics);
    const sortedTypeDiagnostics = sortTypeScriptDiagnostics(typeDiagnostics);

    expect(sortedDiagnostics.length).toBeGreaterThan(0);

    expect({
      lintDiagnostics: sortedDiagnostics,
      typeDiagnostics: sortedTypeDiagnostics,
    }).toMatchSnapshot();
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

  it('should generate TypeScript type checking diagnostics', async () => {
    const testFiles = await getTestFiles('type-checking');
    expect(testFiles.length).toBeGreaterThan(0);

    const config = generateConfig(testFiles, [
      'require-await',
      'no-floating-promises',
      'no-unsafe-assignment',
      'no-unsafe-call',
      'no-unsafe-member-access',
    ]);

    const env = { ...process.env, GOMAXPROCS: '1' };

    const output = execFileSync(TSGOLINT_BIN, ['headless'], {
      input: config,
      env,
    });

    const { diagnostics, typeDiagnostics } = parseHeadlessOutput(output);
    const sortedDiagnostics = sortDiagnostics(diagnostics);
    const sortedTypeDiagnostics = sortTypeScriptDiagnostics(typeDiagnostics);

    // Should have both type and lint diagnostics
    expect(sortedTypeDiagnostics.length).toBeGreaterThan(0);
    expect(sortedDiagnostics.length).toBeGreaterThan(0);

    expect({
      lintDiagnostics: sortedDiagnostics,
      typeDiagnostics: sortedTypeDiagnostics,
    }).toMatchSnapshot();
  });

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

    const { diagnostics, typeDiagnostics } = parseHeadlessOutput(output);
    const sortedDiagnostics = sortDiagnostics(diagnostics);
    const sortedTypeDiagnostics = sortTypeScriptDiagnostics(typeDiagnostics);

    expect({
      lintDiagnostics: sortedDiagnostics,
      typeDiagnostics: sortedTypeDiagnostics,
    }).toMatchSnapshot();
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

    const { diagnostics, typeDiagnostics } = parseHeadlessOutput(output);
    const sortedDiagnostics = sortDiagnostics(diagnostics);
    const sortedTypeDiagnostics = sortTypeScriptDiagnostics(typeDiagnostics);

    expect({
      lintDiagnostics: sortedDiagnostics,
      typeDiagnostics: sortedTypeDiagnostics,
    }).toMatchSnapshot();
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

    function getDiagnostics(
      config: string,
    ): { lintDiagnostics: Diagnostic[]; typeDiagnostics: TypeScriptDiagnostic[] } {
      let output: Buffer;
      output = execFileSync(TSGOLINT_BIN, ['headless'], {
        input: config,
        env: { ...process.env, GOMAXPROCS: '1' },
      });

      const { diagnostics, typeDiagnostics } = parseHeadlessOutput(output);
      return {
        lintDiagnostics: sortDiagnostics(diagnostics),
        typeDiagnostics: sortTypeScriptDiagnostics(typeDiagnostics),
      };
    }

    const testFiles = await getTestFiles('basic');
    expect(testFiles.length).toBeGreaterThan(0);

    const v1Config = generateV1HeadlessPayload(testFiles);
    const v1Result = getDiagnostics(v1Config);

    const v2Config = generateConfig(testFiles);
    const v2Result = getDiagnostics(v2Config);

    // For backward compatibility test, we only compare lint diagnostics
    // since v1 format doesn't support type diagnostics
    expect(v1Result.lintDiagnostics).toStrictEqual(v2Result.lintDiagnostics);
  });
});
