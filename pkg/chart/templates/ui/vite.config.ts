import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import { viteSingleFile } from "vite-plugin-singlefile";
import { createHtmlPlugin } from 'vite-plugin-html';
import path from "path";

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(), 
    viteSingleFile({
      removeViteModuleLoader: true,
      useRecommendedBuildConfig: true,
    }),
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
      "@": path.resolve(__dirname, "./src"),
    },
  },
  esbuild: {
    legalComments: 'none',
    pure: ['console.log', 'console.info', 'console.warn', 'console.debug', 'console.trace'],
  },
  define: {
    'process.env.NODE_ENV': '"production"',
  },
});
