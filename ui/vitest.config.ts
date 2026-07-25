import path from 'path'
import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'

const alias = {
  '@': path.resolve(__dirname, './src'),
}

// Standalone from vite.config.ts so embed UI build plugins never load under test.
// Three projects share one config:
//   unit         — pure node (libs, workers, composables)
//   integration  — happy-dom + Vue SFC mounts
//   e2e smoke is Playwright (see playwright.config.ts), not Vitest.
export default defineConfig({
  resolve: { alias },
  test: {
    globals: false,
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
