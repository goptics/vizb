import { defineConfig, type ESBuildOptions } from 'vite'
import vue from '@vitejs/plugin-vue'
import { createHtmlPlugin } from 'vite-plugin-html'
import path from 'path'
import { createEmbedPlugins, embedBuildOptions } from './embed-build/index.ts'

const esbuildOptions = {
  legalComments: 'none',
  pure: ['console.log', 'console.info', 'console.warn', 'console.debug', 'console.trace'],
} as ESBuildOptions

// https://vite.dev/config/
export default defineConfig(({ command, mode }) => ({
  build: mode === 'embed' ? embedBuildOptions : undefined,
  plugins: [
    vue(),
    ...(mode === 'embed' ? createEmbedPlugins(__dirname) : []),
    createHtmlPlugin({
      minify: {
        removeComments: true,
        collapseWhitespace: true,
        removeAttributeQuotes: true,
        collapseBooleanAttributes: true,
        removeEmptyAttributes: true,
        minifyCSS: true,
        minifyJS: true,
      },
    }),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  esbuild: esbuildOptions,
  define: {
    'process.env.NODE_ENV': command === 'serve' ? '"development"' : '"production"',
  },
}))
