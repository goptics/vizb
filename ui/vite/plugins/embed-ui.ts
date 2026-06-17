import fs from 'node:fs'
import zlib from 'node:zlib'
import path from 'path'
import { parseHTML } from 'linkedom'
import type { PluginOption } from 'vite'
import { detectChartRoots, chunkKeyOf, goStringMap, goStringSliceMap } from '../go-codegen.ts'
import { appendVizbDataScriptTag } from '../html.ts'
import type { GoChunkArtifacts } from '../types.ts'

// echarts-gl bundles the clay.gl WebGL engine, dominating the JS payload. The
// build emits multiple ES module chunks (a 2D entry plus an async echarts-gl
// chunk reached only when a 3D chart renders). This plugin inlines every chunk
// as its own gzip+base64 blob and wires them together with a runtime import
// map, so the browser parses/compiles only the entry on load and defers the gl
// chunk's parse until its dynamic import() actually fires — while the output
// stays a self-contained HTML template embedded in the Go binary. It then emits
// pkg/template/vizb-ui.gen.go.
export const embedUiPlugin = (rootDir: string): PluginOption => {
  const distDir = path.resolve(rootDir, 'dist')
  const distHtmlPath = path.resolve(distDir, 'index.html')
  const goFilePath = path.resolve(rootDir, '..', 'pkg', 'template', 'vizb-ui.gen.go')

  return {
    name: 'embed-ui',
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
        console.warn('[embed-ui] No entry module script found, skipping.')
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
      const chartRoots = detectChartRoots(chunkFiles)
      const chunkSizes: Array<{ file: string; raw: number; b64: number }> = []
      for (const file of chunkFiles) {
        const fileKey = chunkKeyOf(file)
        let code = fs.readFileSync(path.resolve(assetsDir, file), 'utf-8')
        const refs: string[] = []
        for (const other of chunkFiles) {
          if (other === file) continue
          if (code.includes(`./${other}`)) {
            refs.push(chunkKeyOf(other))
            code = code.split(`./${other}`).join(chunkKeyOf(other))
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
        `await import(${JSON.stringify(chunkKeyOf(entryFile))});` +
        `console.timeEnd("parse")` +
        `})()`

      entryScript.remove()
      const boot = document.createElement('script')
      boot.textContent = bootstrap
      document.body.appendChild(boot)

      const out = '<!DOCTYPE html>' + document.documentElement.outerHTML
      fs.writeFileSync(distHtmlPath, out, 'utf-8')

      const artifacts: GoChunkArtifacts = {
        chunks: chunkData,
        imports: chunkImports,
        roots: chartRoots,
        entryKey: chunkKeyOf(entryFile),
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
      console.info(
        `[embed-ui] ${chunkFiles.length} chunks\n${rows}\n\n  HTML template (placeholder): ${templateKB} kB\n`
      )

      const htmlWithState = appendVizbDataScriptTag(out)

      // Escape backticks for Go raw string literal
      const goRawString = htmlWithState.split('`').join('` + "`" + `')
      const { chunks, imports, roots, entryKey } = artifacts

      const goFileContent = `package template

// Code generated by ui/vite/plugins/embed-ui.ts; DO NOT EDIT.

// VizbChunks maps each import-map key to its gzip+base64 chunk blob.
var VizbChunks = map[string]string${goStringMap(chunks)}

// VizbChunkImports is the chunk reference graph (key → keys it imports).
var VizbChunkImports = map[string][]string${goStringSliceMap(imports)}

// VizbChartRoots maps each logical chart (bar/line/pie/heatmap/radar/3d) to its
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
