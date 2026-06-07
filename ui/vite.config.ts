import { defineConfig, type PluginOption } from 'vite'
import vue from '@vitejs/plugin-vue'
import { createHtmlPlugin } from 'vite-plugin-html'
import fs from 'node:fs'
import zlib from 'node:zlib'
import path from 'path'
import { parseHTML } from 'linkedom'

const inlineFaviconPlugin = (): PluginOption => {
  return {
    name: 'inline-favicon',
    enforce: 'pre',
    transformIndexHtml(html: string) {
      const { document } = parseHTML(html)
      const link = document.querySelector('link[rel="icon"]')

      if (!link) {
        return html
      }

      const href = link.getAttribute('href')
      if (!href) return html

      // Strip the leading "/" so the href resolves relative to the project
      // (path.resolve treats a leading slash as an absolute path and discards
      // the base), checking the public directory if needed.
      const relHref = href.replace(/^\//, '')
      let faviconPath = path.resolve(__dirname, relHref)

      if (!fs.existsSync(faviconPath)) {
        faviconPath = path.resolve(__dirname, 'public', relHref)
      }

      if (!fs.existsSync(faviconPath)) {
        console.warn(`[inline-favicon] Could not locate favicon: ${href}`)
        return html
      }

      const source = fs.readFileSync(faviconPath, 'utf-8')
      link.setAttribute('href', `data:${link.getAttribute('type')},${encodeURIComponent(source)}`)

      return '<!DOCTYPE html>' + document.documentElement.outerHTML
    },
  }
}

const appendVizbDataScriptTag = (html: string): string => {
  const { document } = parseHTML(html)

  const script = document.createElement('script')
  script.type = 'text/javascript'
  // Custom Go-template delimiters: echarts-gl's clay.gl GLSL shaders use {{ }}
  // for loop unrolling, which would otherwise collide with html/template parsing.
  script.textContent = `window.VIZB_VERSION = [[VIZB .Version VIZB]]; window.VIZB_DATA = [[VIZB .Data VIZB]]; window.VIZB_DATA_URL = [[VIZB .DataURL VIZB]];`

  document.head.appendChild(script)

  return '<!DOCTYPE html>' + document.documentElement.outerHTML
}

// echarts-gl bundles the clay.gl WebGL engine, dominating the JS payload. The
// build emits multiple ES module chunks (a 2D entry plus an async echarts-gl
// chunk reached only when a 3D chart renders). This plugin inlines every chunk
// as its own gzip+base64 blob and wires them together with a runtime import
// map, so the browser parses/compiles only the entry on load and defers the gl
// chunk's parse until its dynamic import() actually fires — while the output
// stays a single self-contained HTML file.
const compressBundlePlugin = (): PluginOption => {
  const distDir = path.resolve(__dirname, 'dist')
  const distHtmlPath = path.resolve(distDir, 'index.html')
  const keyOf = (file: string) => `vizb:${file.replace(/\.js$/, '')}`

  return {
    name: 'compress-bundle',
    apply: 'build',
    closeBundle() {
      if (!fs.existsSync(distHtmlPath)) return

      const html = fs.readFileSync(distHtmlPath, 'utf-8')
      const { document } = parseHTML(html)

      // Inline the (single, cssCodeSplit:false) stylesheet, drop the <link>.
      for (const link of Array.from(document.querySelectorAll('link[rel="stylesheet"]')) as any[]) {
        const href = link.getAttribute('href')
        if (!href) continue
        const cssPath = path.resolve(distDir, href.replace(/^\//, ''))
        if (fs.existsSync(cssPath)) {
          const style = document.createElement('style')
          style.textContent = fs.readFileSync(cssPath, 'utf-8')
          link.replaceWith(style)
        }
      }

      const entryScript = document.querySelector('script[type="module"][src]')
      if (!entryScript) {
        console.warn('[compress-bundle] No entry module script found, skipping.')
        return
      }
      const entrySrc = entryScript.getAttribute('src')!.replace(/^\//, '')
      const entryFile = path.basename(entrySrc)
      const assetsDir = path.resolve(distDir, path.dirname(entrySrc))
      const chunkFiles = fs.readdirSync(assetsDir).filter((f) => f.endsWith('.js'))

      // gzip each chunk, rewriting rollup's relative sibling/dynamic-import
      // specifiers ("./<chunk>.js") to stable import-map keys so they resolve
      // through the runtime map regardless of the blob URLs assigned later.
      const chunkData: Record<string, string> = {}
      for (const file of chunkFiles) {
        let code = fs.readFileSync(path.resolve(assetsDir, file), 'utf-8')
        for (const other of chunkFiles) {
          code = code.split(`./${other}`).join(keyOf(other))
        }
        chunkData[keyOf(file)] = zlib
          .gzipSync(Buffer.from(code, 'utf-8'), { level: 9 })
          .toString('base64')
      }

      // Classic (non-module) bootstrap: dynamically registering an import map is
      // only permitted before the first module script is prepared, and classic
      // scripts don't lock import-map acquisition. So we decode every chunk to a
      // blob URL, register the map, then dynamically import the entry (which —
      // and its lazy chunks — resolve through the map).
      const bootstrap =
        `(async()=>{` +
        `console.time("parse");` +
        `const C=${JSON.stringify(chunkData)},m={};` +
        `await Promise.all(Object.keys(C).map(async k=>{` +
        `const b=new Uint8Array(await(await fetch("data:application/octet-stream;base64,"+C[k])).arrayBuffer());` +
        `const s=new Blob([b]).stream().pipeThrough(new DecompressionStream("gzip"));` +
        `const code=await new Response(s).text();` +
        `m[k]=URL.createObjectURL(new Blob([code],{type:"text/javascript"}))` +
        `}));` +
        `const im=document.createElement("script");im.type="importmap";` +
        `im.textContent=JSON.stringify({imports:m});document.head.appendChild(im);` +
        `await import(${JSON.stringify(keyOf(entryFile))});` +
        `console.timeEnd("parse")` +
        `})()`

      entryScript.remove()
      const boot = document.createElement('script')
      boot.textContent = bootstrap
      document.body.appendChild(boot)

      const out = '<!DOCTYPE html>' + document.documentElement.outerHTML
      fs.writeFileSync(distHtmlPath, out, 'utf-8')

      const sizes = chunkFiles
        .map((f) => `${f} ${(chunkData[keyOf(f)].length / 1024).toFixed(0)}KB(b64)`)
        .join(', ')
      const compressedKB = (Buffer.byteLength(out, 'utf-8') / 1024).toFixed(1)
      console.info(
        `[compress-bundle] ${chunkFiles.length} chunk(s) [${sizes}] → ${compressedKB} KB HTML`
      )
    },
  }
}

const benchmarkUiGoWrapperPlugin = (): PluginOption => {
  const distHtmlPath = path.resolve(__dirname, 'dist/index.html')
  const goFilePath = path.resolve(__dirname, '..', 'pkg', 'template', 'vizb-ui.gen.go')

  return {
    name: 'benchmark-ui-go-wrapper',
    apply: 'build',
    closeBundle() {
      if (!fs.existsSync(distHtmlPath)) {
        console.warn(
          `[benchmark-ui-go-wrapper] Unable to find ${distHtmlPath}, skipping Go wrapper generation.`
        )
        return
      }

      const htmlContent = fs.readFileSync(distHtmlPath, 'utf-8')
      const htmlWithState = appendVizbDataScriptTag(htmlContent)

      // Escape backticks for Go raw string literal
      const goRawString = htmlWithState.split('`').join('` + "`" + `')

      const goFileContent = `package template

// Code generated by vite.config.ts; DO NOT EDIT.

const VizbHTMLTemplate = \`${goRawString}\`
`

      fs.writeFileSync(goFilePath, goFileContent, 'utf-8')
    },
  }
}

const singleFile = process.env.SINGLEFILE === 'True'
console.info('SINGLEFILE env var:', process.env.SINGLEFILE)

const plugins: PluginOption[] = [vue()]

if (singleFile) {
  plugins.push(inlineFaviconPlugin(), compressBundlePlugin(), benchmarkUiGoWrapperPlugin())
}

// In single-file mode we own the inlining (compressBundlePlugin): keep CSS in
// one file to inline, drop modulepreload (its dep arrays would defy the import
// map), and inline any static assets so nothing external is referenced. Dynamic
// imports are intentionally NOT inlined — that's what keeps echarts-gl in its
// own lazily-parsed chunk.
const singleFileBuild = singleFile
  ? {
      cssCodeSplit: false,
      modulePreload: false as const,
      assetsInlineLimit: 100_000_000,
    }
  : undefined

// https://vite.dev/config/
export default defineConfig({
  build: singleFileBuild,
  plugins: [
    ...plugins,
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
  esbuild: {
    legalComments: 'none',
    pure: ['console.log', 'console.info', 'console.warn', 'console.debug', 'console.trace'],
  },
  define: {
    'process.env.NODE_ENV': '"production"',
  },
})
