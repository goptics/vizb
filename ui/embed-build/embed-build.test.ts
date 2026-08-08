import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import fs from 'node:fs'
import os from 'node:os'
import path from 'path'
import zlib from 'node:zlib'
import * as childProcess from 'node:child_process'
import { appendVizbDataScriptTag } from './html.ts'
import { createEmbedPlugins, embedBuildOptions } from './index.ts'
import { embedUiPlugin } from './plugins/embed-ui.ts'
import { inlineFaviconPlugin } from './plugins/inline-favicon.ts'
import { detectChartRoots } from './go-codegen.ts'
import type { GoChunkArtifacts } from './types.ts'
import { CHART_ROOT_PREFIX } from './constants.ts'
import './types.ts'

vi.mock('node:child_process', async (importOriginal) => {
  const actual = await importOriginal<typeof import('node:child_process')>()
  return {
    ...actual,
    execSync: vi.fn(actual.execSync),
  }
})

type InlineFaviconPlugin = {
  name: string
  enforce?: string
  transformIndexHtml: (html: string) => string
}

type EmbedUiPlugin = {
  name: string
  apply: string
  closeBundle: () => void
}

describe('html.appendVizbDataScriptTag', () => {
  it('appends VIZB placeholders into head and keeps doctype', () => {
    const out = appendVizbDataScriptTag(
      '<!DOCTYPE html><html><head><title>t</title></head><body></body></html>'
    )
    expect(out.startsWith('<!DOCTYPE html>')).toBe(true)
    expect(out).toContain('window.VIZB_VERSION = [[VIZB .Version VIZB]]')
    expect(out).toContain('window.VIZB_DATA = [[VIZB .Data VIZB]]')
    expect(out).toContain('window.VIZB_DATA_URL = [[VIZB .DataURL VIZB]]')
    expect(out).toContain('window.VIZB_CHARTS = [[VIZB .ChartList VIZB]]')
  })
})

describe('index exports', () => {
  it('exposes embed build options and plugin factory', () => {
    expect(embedBuildOptions).toEqual({
      cssCodeSplit: false,
      modulePreload: false,
      assetsInlineLimit: 100_000_000,
    })
    const plugins = createEmbedPlugins('/tmp/root')
    expect(plugins).toHaveLength(2)
    const first = plugins[0] as InlineFaviconPlugin
    const second = plugins[1] as EmbedUiPlugin
    expect(first.name).toBe('inline-favicon')
    expect(second.name).toBe('embed-ui')
  })
})

describe('types + constants touch', () => {
  it('uses GoChunkArtifacts shape and chart root map', () => {
    const sample: GoChunkArtifacts = {
      chunks: { a: 'b' },
      imports: { a: [] },
      roots: { bar: 'vizb:ChartBar' },
      entryKey: 'vizb:index',
    }
    expect(sample.entryKey).toBe('vizb:index')
    expect(CHART_ROOT_PREFIX.ChartBar).toBe('bar')
    expect(CHART_ROOT_PREFIX.ChartSankey).toBe('sankey')
  })
})

describe('detectChartRoots branch gaps', () => {
  it('skips prefixes whose mapped chart name is falsy', () => {
    const roots = detectChartRoots(['ChartGhost-abc.js', 'ChartBar-x.js'], {
      ChartGhost: '',
      ChartBar: 'bar',
    })
    expect(roots).toEqual({ bar: 'vizb:ChartBar-x' })
  })
})

describe('inlineFaviconPlugin', () => {
  let tmp: string

  beforeEach(() => {
    tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'inline-fav-'))
  })

  afterEach(() => {
    fs.rmSync(tmp, { recursive: true, force: true })
    vi.restoreAllMocks()
  })

  const transform = (root: string, html: string) => {
    const plugin = inlineFaviconPlugin(root) as InlineFaviconPlugin
    return plugin.transformIndexHtml(html)
  }

  it('returns html unchanged when no icon link exists', () => {
    const html = '<html><head></head><body></body></html>'
    expect(transform(tmp, html)).toBe(html)
  })

  it('returns html unchanged when href is missing', () => {
    const html = '<html><head><link rel="icon" type="image/svg+xml"></head><body></body></html>'
    expect(transform(tmp, html)).toBe(html)
  })

  it('inlines favicon from project root path', () => {
    fs.writeFileSync(path.join(tmp, 'favicon.svg'), '<svg id="root"/>')
    const html =
      '<html><head><link rel="icon" type="image/svg+xml" href="/favicon.svg"></head><body></body></html>'
    const out = transform(tmp, html)
    expect(out.startsWith('<!DOCTYPE html>')).toBe(true)
    expect(out).toContain('data:image/svg+xml,')
    expect(out).toContain(encodeURIComponent('<svg id="root"/>'))
  })

  it('falls back to public/ when root file is missing', () => {
    fs.mkdirSync(path.join(tmp, 'public'))
    fs.writeFileSync(path.join(tmp, 'public', 'icon.svg'), '<svg id="pub"/>')
    const html =
      '<html><head><link rel="icon" type="image/svg+xml" href="icon.svg"></head><body></body></html>'
    const out = transform(tmp, html)
    expect(out).toContain(encodeURIComponent('<svg id="pub"/>'))
  })

  it('warns and returns original html when favicon cannot be found', () => {
    const warn = vi.spyOn(console, 'warn').mockImplementation(() => {})
    const html =
      '<html><head><link rel="icon" type="image/svg+xml" href="/missing.svg"></head><body></body></html>'
    expect(transform(tmp, html)).toBe(html)
    expect(warn).toHaveBeenCalledWith(
      expect.stringContaining('[inline-favicon] Could not locate favicon: /missing.svg')
    )
  })
})

describe('embedUiPlugin.closeBundle', () => {
  let workspace: string
  let root: string
  let dist: string
  let assets: string
  let goPath: string

  beforeEach(() => {
    workspace = fs.mkdtempSync(path.join(os.tmpdir(), 'embed-ui-'))
    root = path.join(workspace, 'ui')
    dist = path.join(root, 'dist')
    assets = path.join(dist, 'assets')
    fs.mkdirSync(assets, { recursive: true })
    // plugin writes ../pkg/template/vizb-ui.gen.go relative to rootDir
    goPath = path.join(workspace, 'pkg', 'template', 'vizb-ui.gen.go')
    fs.mkdirSync(path.dirname(goPath), { recursive: true })
    vi.mocked(childProcess.execSync).mockReset()
  })

  afterEach(() => {
    fs.rmSync(workspace, { recursive: true, force: true })
    vi.restoreAllMocks()
  })

  const runClose = (rootDir: string) => {
    const plugin = embedUiPlugin(rootDir) as EmbedUiPlugin
    expect(plugin.name).toBe('embed-ui')
    expect(plugin.apply).toBe('build')
    plugin.closeBundle()
  }

  it('no-ops when dist index.html is missing', () => {
    runClose(root)
    expect(fs.existsSync(goPath)).toBe(false)
  })

  it('warns and returns when no entry module script is present', () => {
    const warn = vi.spyOn(console, 'warn').mockImplementation(() => {})
    fs.writeFileSync(
      path.join(dist, 'index.html'),
      '<!DOCTYPE html><html><head></head><body></body></html>'
    )
    runClose(root)
    expect(warn).toHaveBeenCalledWith('[embed-ui] No entry module script found, skipping.')
    expect(fs.existsSync(goPath)).toBe(false)
  })

  it('inlines css, gzips chunks, rewrites imports, and emits go file', () => {
    const info = vi.spyOn(console, 'info').mockImplementation(() => {})
    const warn = vi.spyOn(console, 'warn').mockImplementation(() => {})
    fs.writeFileSync(path.join(assets, 'index.css'), 'body{color:red}')
    const entry = 'index-abc.js'
    const chart = 'ChartBar-xyz.js'
    const shared = 'shared-1.js'
    fs.writeFileSync(
      path.join(assets, entry),
      `import "./${shared}"; import("./${chart}"); console.log("entry")`
    )
    fs.writeFileSync(path.join(assets, chart), `import "./${shared}"; export const bar=1`)
    fs.writeFileSync(path.join(assets, shared), 'export const s=1')

    const html = `<!DOCTYPE html><html><head>
<link rel="stylesheet" href="/assets/index.css">
<link rel="stylesheet" href="">
<link rel="stylesheet" href="/assets/missing.css">
<script type="module" src="/assets/${entry}"></script>
</head><body></body></html>`
    fs.writeFileSync(path.join(dist, 'index.html'), html)

    // Cover formatGoFile catch path (gofmt unavailable / failing).
    vi.mocked(childProcess.execSync).mockImplementation(() => {
      throw new Error('no gofmt')
    })

    runClose(root)

    const outHtml = fs.readFileSync(path.join(dist, 'index.html'), 'utf-8')
    expect(outHtml).toContain('<style>body{color:red}</style>')
    expect(outHtml).not.toContain(`src="/assets/${entry}"`)
    expect(outHtml).toContain('[[VIZB .Chunks VIZB]]')
    expect(outHtml).toContain(`await import("vizb:${entry.replace(/\.js$/, '')}")`)

    expect(fs.existsSync(goPath)).toBe(true)
    const go = fs.readFileSync(goPath, 'utf-8')
    expect(go).toContain('package template')
    expect(go).toContain('var VizbChunks')
    expect(go).toContain('var VizbChunkImports')
    expect(go).toContain('var VizbChartRoots')
    expect(go).toContain('var VizbEntryKey')
    expect(go).toContain('const VizbHTMLTemplate')
    expect(go).toContain('window.VIZB_DATA')
    expect(go).toContain('"bar":')

    const b64Match = go.match(/"vizb:shared-1": "([A-Za-z0-9+/=]+)"/)
    expect(b64Match).toBeTruthy()
    const decoded = zlib.gunzipSync(Buffer.from(b64Match![1], 'base64')).toString('utf-8')
    expect(decoded).toContain('export const s=1')

    expect(info).toHaveBeenCalled()
    const report = String(info.mock.calls[0]?.[0])
    expect(report).toContain('[embed-ui] 3 chunks')

    const embeddedHtmlBytes = Buffer.byteLength(appendVizbDataScriptTag(outHtml), 'utf-8')
    const chunkKeys = ['index-abc', 'ChartBar-xyz', 'shared-1']
    const encodedChunks = chunkKeys.map((key) => {
      const match = go.match(new RegExp(`"vizb:${key}": "([A-Za-z0-9+/=]+)"`))
      expect(match, `missing encoded chunk ${key}`).toBeTruthy()
      return match![1]!
    })
    const rawTotalBytes =
      embeddedHtmlBytes +
      encodedChunks.reduce(
        (total, chunk) => total + zlib.gunzipSync(Buffer.from(chunk, 'base64')).length,
        0
      )
    const encodedTotalBytes =
      embeddedHtmlBytes + encodedChunks.reduce((total, chunk) => total + chunk.length, 0)

    expect(report).toContain(
      `HTML template (placeholder): ${(embeddedHtmlBytes / 1024).toFixed(2)} kB`
    )
    expect(report).toContain(
      `Total embedded payload (all chunks + HTML): actual: ${(rawTotalBytes / 1024).toFixed(2)} kB │ encoded: ${(encodedTotalBytes / 1024).toFixed(2)} kB`
    )
    expect(warn).toHaveBeenCalledWith(expect.stringContaining('[embed-ui] gofmt skipped'))
  })

  it('escapes backticks in html for Go raw string and invokes gofmt', () => {
    vi.spyOn(console, 'info').mockImplementation(() => {})
    const entry = 'index-bt.js'
    fs.writeFileSync(path.join(assets, entry), 'console.log("hi")')
    fs.writeFileSync(
      path.join(dist, 'index.html'),
      `<!DOCTYPE html><html><head><script type="module" src="/assets/${entry}"></script></head><body data-x="\`tick\`"></body></html>`
    )

    vi.mocked(childProcess.execSync).mockReturnValue(Buffer.from(''))

    runClose(root)
    const go = fs.readFileSync(goPath, 'utf-8')
    expect(go).toContain('` + "`" + `')
    expect(childProcess.execSync).toHaveBeenCalledWith(
      expect.stringContaining('gofmt -w'),
      expect.objectContaining({ stdio: 'pipe' })
    )
  })
})
