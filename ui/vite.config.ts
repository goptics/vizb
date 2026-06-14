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
  script.textContent = `window.VIZB_VERSION = [[VIZB .Version VIZB]]; window.VIZB_DATA = [[VIZB .Data VIZB]]; window.VIZB_DATA_URL = [[VIZB .DataURL VIZB]]; window.VIZB_CHARTS = [[VIZB .ChartList VIZB]];`

  document.head.appendChild(script)

  return '<!DOCTYPE html>' + document.documentElement.outerHTML
}

// Manifest captured by compressBundlePlugin and consumed by the Go-wrapper
// plugin so the Go side can prune unreachable chunks per --charts at generation
// time. `imports` is the gzipped-chunk reference graph (keyed by import-map key),
// `roots` maps each logical chart to its renderer chunk key, `entryKey` is the
// always-present entry chunk.
type GoChunkArtifacts = {
  chunks: Record<string, string>
  imports: Record<string, string[]>
  roots: Record<string, string>
  entryKey: string
}
let goChunkArtifacts: GoChunkArtifacts | null = null

// Dynamic-import chunk filename prefix → logical chart name. Rollup names a
// dynamic-import chunk after its module, so "ChartBar-<hash>.js" → "bar". These
// are the only chunks the Go pruner gates; everything else (shared echarts core,
// vendor) is always kept when reachable.
const CHART_ROOT_PREFIX: Record<string, string> = {
  ChartBar: 'bar',
  ChartLine: 'line',
  ChartPie: 'pie',
  ChartHeatmap: 'heatmap',
  Chart3D: '3d',
}

// Emit a Go `map[string]string{…}` body from a JS object. base64 chunk blobs and
// import-map keys are ASCII with no Go-string-breaking chars, so JSON.stringify
// yields valid Go double-quoted string literals.
const goStringMap = (m: Record<string, string>): string =>
  '{\n' +
  Object.entries(m)
    .map(([k, v]) => `\t${JSON.stringify(k)}: ${JSON.stringify(v)},`)
    .join('\n') +
  '\n}'

// Emit a Go `map[string][]string{…}` body from a JS object of string arrays.
const goStringSliceMap = (m: Record<string, string[]>): string =>
  '{\n' +
  Object.entries(m)
    .map(([k, arr]) => `\t${JSON.stringify(k)}: {${arr.map((s) => JSON.stringify(s)).join(', ')}},`)
    .join('\n') +
  '\n}'

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
      const chunkImports: Record<string, string[]> = {}
      const chartRoots: Record<string, string> = {}
      const chunkSizes: Array<{ file: string; raw: number; b64: number }> = []
      for (const file of chunkFiles) {
        const fileKey = keyOf(file)
        // Tag gated renderer chunks (ChartBar-<hash>.js → "bar", …) so the Go
        // pruner knows which chunk each logical chart resolves to.
        for (const prefix of Object.keys(CHART_ROOT_PREFIX)) {
          if (file.startsWith(`${prefix}-`) || file === `${prefix}.js`) {
            chartRoots[CHART_ROOT_PREFIX[prefix]] = fileKey
          }
        }
        let code = fs.readFileSync(path.resolve(assetsDir, file), 'utf-8')
        const refs: string[] = []
        for (const other of chunkFiles) {
          if (other === file) continue
          if (code.includes(`./${other}`)) {
            refs.push(keyOf(other))
            code = code.split(`./${other}`).join(keyOf(other))
          }
        }
        chunkImports[fileKey] = refs
        const rawBuf = Buffer.from(code, 'utf-8')
        const gzBuf = zlib.gzipSync(rawBuf, { level: 9 })
        const b64 = gzBuf.toString('base64')
        chunkData[fileKey] = b64
        chunkSizes.push({ file, raw: rawBuf.length, b64: b64.length })
      }

      // Classic (non-module) bootstrap: dynamically registering an import map is
      // only permitted before the first module script is prepared, and classic
      // scripts don't lock import-map acquisition. So we decode every chunk to a
      // blob URL, register the map, then dynamically import the entry (which —
      // and its lazy chunks — resolve through the map).
      // The chunk map is injected by Go at generation time ([[VIZB .Chunks VIZB]])
      // so it can ship only the chunks the selected charts + data can reach. The
      // placeholder survives because this bootstrap is appended in closeBundle,
      // after HTML minification, exactly like the [[VIZB .Data VIZB]] data script.
      const bootstrap =
        `(async()=>{` +
        `console.time("parse");` +
        `const C=[[VIZB .Chunks VIZB]],m={};` +
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

      // Hand the chunk blobs + reference graph to the Go-wrapper plugin, which
      // emits them as Go vars next to the template so Go can prune at gen time.
      goChunkArtifacts = {
        chunks: chunkData,
        imports: chunkImports,
        roots: chartRoots,
        entryKey: keyOf(entryFile),
      }

      const nameW = Math.max(...chunkSizes.map(({ file }) => file.length))
      const rawW = Math.max(...chunkSizes.map(({ raw }) => (raw / 1024).toFixed(2).length))
      const b64W = Math.max(...chunkSizes.map(({ b64 }) => (b64 / 1024).toFixed(2).length))
      const rows = chunkSizes
        .sort((a, b) => b.raw - a.raw)
        .map(({ file, raw, b64 }) => {
          const name = `dist/assets/${file}`.padEnd(nameW + 'dist/assets/'.length)
          const rawStr = (raw / 1024).toFixed(2).padStart(rawW)
          const b64Str = (b64 / 1024).toFixed(2).padStart(b64W)
          return `  ${name}  ${rawStr} kB │ encoded: ${b64Str} kB`
        })
        .join('\n')
      const templateKB = (Buffer.byteLength(out, 'utf-8') / 1024).toFixed(2)
      console.info(`\n[compress-bundle] ${chunkFiles.length} chunks\n${rows}\n\n  HTML template (placeholder): ${templateKB} kB\n`)
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

      if (!goChunkArtifacts) {
        console.warn(
          '[benchmark-ui-go-wrapper] No chunk artifacts captured; the template references [[VIZB .Chunks VIZB]] but no chunk vars will be emitted.'
        )
      }
      const { chunks, imports, roots, entryKey } = goChunkArtifacts ?? {
        chunks: {},
        imports: {},
        roots: {},
        entryKey: '',
      }

      const goFileContent = `package template

// Code generated by vite.config.ts; DO NOT EDIT.

// VizbChunks maps each import-map key to its gzip+base64 chunk blob.
var VizbChunks = map[string]string${goStringMap(chunks)}

// VizbChunkImports is the chunk reference graph (key → keys it imports).
var VizbChunkImports = map[string][]string${goStringSliceMap(imports)}

// VizbChartRoots maps each logical chart (bar/line/pie/heatmap/3d) to its
// renderer chunk key. These are the only chunks the pruner gates.
var VizbChartRoots = map[string]string${goStringMap(roots)}

// VizbEntryKey is the entry chunk key, always shipped.
var VizbEntryKey = ${JSON.stringify(entryKey)}

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
