import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    include: ['**/*.test.ts'],
    testTimeout: 60000,
    pool: 'forks',
    poolOptions: {
      forks: {
        singleFork: true
      }
    }
  }
});