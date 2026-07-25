import path from 'path'
import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'

const alias = {
  '@': path.resolve(__dirname, './src'),
}

// Standalone from vite.config.ts so embed UI build plugins never load under test.
// Projects:
//   unit         — pure node (libs, workers, composables)
//   integration  — happy-dom + Vue SFC mounts
//   e2e smoke is Playwright (see playwright.config.ts), not Vitest.
export default defineConfig({
  resolve: { alias },
  test: {
    globals: false,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'text-summary', 'html', 'json-summary'],
      reportsDirectory: './coverage',
      // All production UI source. Types/d.ts and test helpers stay out.
      include: ['src/**/*.{ts,vue}', 'embed-build/**/*.ts'],
      exclude: [
        'src/**/*.test.ts',
        'src/**/*.integration.test.ts',
        'src/test-utils/**',
        'src/workers/__test-utils__/**',
        'src/**/*.d.ts',
        'src/types/**',
        'src/data/**',
        // Pure re-export / type-only modules (0 instrumentable statements).
        'src/components/ui/index.ts',
        'src/components/ui/combobox.ts',
        'src/composables/charts/index.ts',
        'src/composables/charts/shared/index.ts',
        'embed-build/types.ts',
        'embed-build/**/*.test.ts',
        'e2e/**',
      ],
      // 100% on statements/branches/functions/lines for included files.
      thresholds: {
        statements: 100,
        branches: 100,
        functions: 100,
        lines: 100,
      },
    },
    projects: [
      {
        resolve: { alias },
        test: {
          name: 'unit',
          environment: 'node',
          include: ['src/**/*.test.ts', 'embed-build/**/*.test.ts'],
          exclude: ['src/**/*.integration.test.ts', 'e2e/**'],
        },
      },
      {
        plugins: [vue()],
        resolve: { alias },
        test: {
          name: 'integration',
          environment: 'happy-dom',
          include: ['src/**/*.integration.test.ts'],
          exclude: ['e2e/**'],
        },
      },
    ],
  },
})
