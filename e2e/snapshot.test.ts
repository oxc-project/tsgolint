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
];

interface Diagnostic {
    file_path?: string;
    rule?: string;
    range?: {
        pos: number;
        end: number;
    };
    [key: string]: any;
}

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

async function getTestFiles(): Promise<string[]> {
    const patterns = ['**/*.ts', '**/*.tsx', '**/*.mts', '**/*.cts'];
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

function generateConfig(files: string[]): string {
    const config = {
        files: files.map((filePath) => ({
            file_path: filePath,
            rules: ALL_RULES,
        })),
    };
    return JSON.stringify(config);
}

describe('TSGoLint E2E Snapshot Tests', () => {
    it('should generate consistent diagnostics snapshot', async () => {
        const testFiles = await getTestFiles();
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

        let diagnostics = parseHeadlessOutput(output);
        diagnostics = sortDiagnostics(diagnostics);

        expect(diagnostics.length).toBeGreaterThan(0);
        expect(diagnostics).toMatchSnapshot();
    });
});
