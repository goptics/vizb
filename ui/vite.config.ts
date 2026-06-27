import { defineConfig, type ESBuildOptions } from 'vite'
import vue from '@vitejs/plugin-vue'
import { createHtmlPlugin } from 'vite-plugin-html'
import path from 'path'
import { createEmbedPlugins, embedBuildOptions } from './embed-build/index.ts'

const embedUi = process.env.EMBED_UI === 'True'
console.info('EMBED_UI env var:', process.env.EMBED_UI)

const esbuildOptions = {
  legalComments: 'none',
  pure: ['console.log', 'console.info', 'console.warn', 'console.debug', 'console.trace'],
} as ESBuildOptions

// https://vite.dev/config/
export default defineConfig(({ command }) => ({
  build: embedUi ? embedBuildOptions : undefined,
  plugins: [
    vue(),
    ...(embedUi ? createEmbedPlugins(__dirname) : []),
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
