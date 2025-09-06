import { execFileSync } from 'node:child_process';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, it, expect } from 'vitest';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const ROOT_DIR = resolve(__dirname, '..');
const E2E_DIR = __dirname;
const FIXTURES_DIR = join(E2E_DIR, 'fixtures');
const TSGOLINT_BIN = join(ROOT_DIR, `tsgolint${process.platform === 'win32' ? '.exe' : ''}`);

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

describe('TSConfig Path Aliases Integration Test', () => {
    it('should handle tsconfig extending another tsconfig with path aliases', async () => {
        // Test files using path aliases
        const pathAliasesDir = join(FIXTURES_DIR, 'path-aliases');
        const testFiles = [
            join(pathAliasesDir, 'src', 'index.ts'),
            join(pathAliasesDir, 'src', 'utils', 'api-client.ts'),
            join(pathAliasesDir, 'src', 'constants', 'api.ts'),
            join(pathAliasesDir, 'src', 'constants', 'app.ts')
        ];

        const config = {
            files: testFiles.map((filePath) => ({
                file_path: filePath,
                rules: ['no-floating-promises'], // Rule that should catch our intentional issues
            })),
        };

        const env = { ...process.env, GOMAXPROCS: '1' };

        // Run tsgolint in headless mode
        let output: Buffer;
        try {
            output = execFileSync(TSGOLINT_BIN, ['headless'], {
                input: JSON.stringify(config),
                env,
            });
        } catch (error: any) {
            // If there's stderr output, include it in the error message
            const stderr = error.stderr?.toString() || '';
            const stdout = error.stdout?.toString() || '';
            throw new Error(`tsgolint failed: ${error.message}\nstdout: ${stdout}\nstderr: ${stderr}`);
        }

        const diagnostics = parseHeadlessOutput(output);
        
        // We expect to find diagnostics for floating promises
        expect(diagnostics.length).toBeGreaterThan(0);
        
        // Verify that diagnostics are found in files using path aliases
        const foundFiles = new Set(diagnostics.map(d => d.file_path));
        
        // Should find issues in index.ts (main() and makeApiCall() calls)
        expect(Array.from(foundFiles).some(filePath => 
            filePath?.includes('index.ts')
        )).toBe(true);
        
        // Should find issues in api-client.ts (fetch() call)
        expect(Array.from(foundFiles).some(filePath => 
            filePath?.includes('api-client.ts')
        )).toBe(true);
        
        // All diagnostics should be for the no-floating-promises rule
        diagnostics.forEach(diagnostic => {
            expect(diagnostic.rule).toBe('no-floating-promises');
        });
        
        console.log(`Found ${diagnostics.length} diagnostics in path alias test`);
        console.log('Files with issues:', Array.from(foundFiles));
    });

    it('should resolve path aliases correctly without errors', async () => {
        // Test that the tsconfig setup itself doesn't cause errors
        const pathAliasesDir = join(FIXTURES_DIR, 'path-aliases');
        const testFiles = [
            join(pathAliasesDir, 'src', 'constants', 'api.ts'),
            join(pathAliasesDir, 'src', 'constants', 'app.ts')
        ];

        const config = {
            files: testFiles.map((filePath) => ({
                file_path: filePath,
                rules: ['await-thenable'], // A rule that shouldn't trigger on our constants
            })),
        };

        const env = { ...process.env, GOMAXPROCS: '1' };

        // This should not throw - the path aliases should resolve correctly
        expect(() => {
            execFileSync(TSGOLINT_BIN, ['headless'], {
                input: JSON.stringify(config),
                env,
            });
        }).not.toThrow();
    });
});