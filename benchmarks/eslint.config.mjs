import path from 'node:path';
import process from 'node:process';
import { fileURLToPath } from 'node:url';
import tseslint from 'typescript-eslint';

const { TSGOLINT_BENCHMARK_PROJECT } = process.env;

let files = [];
let project = '';

if (TSGOLINT_BENCHMARK_PROJECT === 'vscode') {
  files = ['src/**/*.ts'];
  project = './src/tsconfig.json';
} else if (TSGOLINT_BENCHMARK_PROJECT === 'typescript') {
  files = ['src/**/*.ts'];
  project = './src/tsconfig-eslint.json';
} else if (TSGOLINT_BENCHMARK_PROJECT === 'typeorm') {
  files = ['src/**/*.ts', 'sample/**/*.ts', 'test/**/*.ts'];
  project = './tsconfig.json';
} else if (TSGOLINT_BENCHMARK_PROJECT === 'vuejs') {
  files = [
    'packages/global.d.ts',
    'packages/*/src/**/*.ts',
    'packages/*/__tests__/**/*.ts',
    'packages/vue/jsx-runtime/**/*.ts',
    'packages/runtime-dom/types/jsx.d.ts',
    'scripts/*.ts',
  ];
  project = './tsconfig.json';
}

export default tseslint.config(
  {
    ignores: ['**/*.js', '**/*.mjs', '**/*.cjs'],
  },
  {
    linterOptions: {
      noInlineConfig: true,
      reportUnusedInlineConfigs: 'off',
      reportUnusedDisableDirectives: 'off',
    },
  },
  {
    files,
    languageOptions: {
      parser: tseslint.parser,
      parserOptions: {
        project,
        tsconfigRootDir: path.dirname(fileURLToPath(import.meta.url)),
      },
    },
    plugins: {
      '@typescript-eslint': tseslint.plugin,
    },
    rules: {
      '@typescript-eslint/await-thenable': 'error',
      '@typescript-eslint/consistent-return': 'error',
      '@typescript-eslint/consistent-type-exports': 'error',
      '@typescript-eslint/dot-notation': 'error',
      '@typescript-eslint/no-array-delete': 'error',
      '@typescript-eslint/no-base-to-string': 'error',
      '@typescript-eslint/no-confusing-void-expression': 'error',
      '@typescript-eslint/no-deprecated': 'error',
      '@typescript-eslint/no-duplicate-type-constituents': 'error',
      '@typescript-eslint/no-floating-promises': 'error',
      '@typescript-eslint/no-for-in-array': 'error',
      '@typescript-eslint/no-implied-eval': 'error',
      '@typescript-eslint/no-meaningless-void-operator': 'error',
      '@typescript-eslint/no-misused-promises': 'error',
      '@typescript-eslint/no-misused-spread': 'error',
      '@typescript-eslint/no-mixed-enums': 'error',
      '@typescript-eslint/no-redundant-type-constituents': 'error',
      '@typescript-eslint/no-unnecessary-boolean-literal-compare': 'error',
      '@typescript-eslint/no-unnecessary-condition': 'error',
      '@typescript-eslint/no-unnecessary-qualifier': 'error',
      '@typescript-eslint/no-unnecessary-template-expression': 'error',
      '@typescript-eslint/no-unnecessary-type-conversion': 'error',
      '@typescript-eslint/no-unnecessary-type-arguments': 'error',
      '@typescript-eslint/no-unnecessary-type-assertion': 'error',
      '@typescript-eslint/no-unnecessary-type-parameters': 'error',
      '@typescript-eslint/no-useless-default-assignment': 'error',
      '@typescript-eslint/no-unsafe-argument': 'error',
      '@typescript-eslint/no-unsafe-assignment': 'error',
      '@typescript-eslint/no-unsafe-call': 'error',
      '@typescript-eslint/no-unsafe-enum-comparison': 'error',
      '@typescript-eslint/no-unsafe-member-access': 'error',
      '@typescript-eslint/no-unsafe-return': 'error',
      '@typescript-eslint/no-unsafe-type-assertion': 'error',
      '@typescript-eslint/no-unsafe-unary-minus': 'error',
      '@typescript-eslint/non-nullable-type-assertion-style': 'error',
      '@typescript-eslint/only-throw-error': 'error',
      '@typescript-eslint/prefer-find': 'error',
      '@typescript-eslint/prefer-includes': 'error',
      '@typescript-eslint/prefer-nullish-coalescing': 'error',
      '@typescript-eslint/prefer-optional-chain': 'error',
      '@typescript-eslint/prefer-promise-reject-errors': 'error',
      '@typescript-eslint/prefer-readonly': 'error',
      '@typescript-eslint/prefer-readonly-parameter-types': 'error',
      '@typescript-eslint/prefer-reduce-type-parameter': 'error',
      '@typescript-eslint/prefer-regexp-exec': 'error',
      '@typescript-eslint/prefer-return-this-type': 'error',
      '@typescript-eslint/prefer-string-starts-ends-with': 'error',
      '@typescript-eslint/promise-function-async': 'error',
      '@typescript-eslint/related-getter-setter-pairs': 'error',
      '@typescript-eslint/require-array-sort-compare': 'error',
      '@typescript-eslint/require-await': 'error',
      '@typescript-eslint/restrict-plus-operands': 'error',
      '@typescript-eslint/restrict-template-expressions': 'error',
      '@typescript-eslint/return-await': 'error',
      '@typescript-eslint/strict-boolean-expressions': 'error',
      '@typescript-eslint/strict-void-return': 'error',
      '@typescript-eslint/switch-exhaustiveness-check': 'error',
      '@typescript-eslint/unbound-method': 'error',
      '@typescript-eslint/use-unknown-in-catch-callback-variable': 'error',
    },
  },
);
