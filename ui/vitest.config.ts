import path from 'path'
import { defineConfig } from 'vitest/config'

// Standalone from vite.config.ts so the embed UI build plugins (favicon
// inlining, bundle compression, Go wrapper, HTML minify) never load under test.
// The stats/csv units are pure functions — a plain node environment is enough.
export default defineConfig({
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  test: {
    environment: 'node',
    include: ['src/**/*.test.ts', 'embed-build/**/*.test.ts'],
    globals: false,
  },
})
