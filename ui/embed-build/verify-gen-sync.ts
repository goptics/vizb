import fs from 'node:fs'
import zlib from 'node:zlib'

const usage = 'Usage: node embed-build/verify-gen-sync.ts <committed.go> <rebuilt.go>'

const read = (path: string): string => {
  if (!fs.existsSync(path)) {
    console.error(`missing file: ${path}`)
    process.exit(1)
  }
  return fs.readFileSync(path, 'utf-8')
}

const extractScalar = (source: string, name: string): string => {
  const match = source.match(new RegExp(`var ${name} = ("[^"]*")`))
  if (!match) throw new Error(`could not parse ${name}`)
  return JSON.parse(match[1].replace(/\\"/g, '"'))
}

const extractStringMap = (source: string, name: string): Record<string, string> => {
  const start = source.indexOf(`var ${name} = map[string]string{`)
  if (start < 0) throw new Error(`could not parse ${name}`)
  const bodyStart = source.indexOf('{', start) + 1
  const bodyEnd = source.indexOf('\n}', bodyStart)
  const body = source.slice(bodyStart, bodyEnd)
  const entries: Record<string, string> = {}
  const re = /^\t("(?:\\.|[^"\\])+"):\s*("(?:\\.|[^"\\])+"),$/gm
  for (const match of body.matchAll(re)) {
    entries[JSON.parse(match[1])] = JSON.parse(match[2])
  }
  if (Object.keys(entries).length === 0) throw new Error(`no entries parsed for ${name}`)
  return entries
}

const extractSliceMap = (source: string, name: string): Record<string, string[]> => {
  const start = source.indexOf(`var ${name} = map[string][]string{`)
  if (start < 0) throw new Error(`could not parse ${name}`)
  const bodyStart = source.indexOf('{', start) + 1
  const bodyEnd = source.indexOf('\n}', bodyStart)
  const body = source.slice(bodyStart, bodyEnd)
  const entries: Record<string, string[]> = {}
  const re = /^\t("(?:\\.|[^"\\])+"):\s*\{([^}]*)\},$/gm
  for (const match of body.matchAll(re)) {
    const key = JSON.parse(match[1]) as string
    const refs = match[2]
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
      .map((s) => JSON.parse(s) as string)
    entries[key] = refs
  }
  if (Object.keys(entries).length === 0) throw new Error(`no entries parsed for ${name}`)
  return entries
}

const extractHTMLTemplate = (source: string): string => {
  const marker = 'const VizbHTMLTemplate = `'
  const start = source.indexOf(marker)
  if (start < 0) throw new Error('could not parse VizbHTMLTemplate')
  const from = start + marker.length
  const end = source.lastIndexOf('`')
  if (end <= from) throw new Error('could not find VizbHTMLTemplate terminator')
  return source.slice(from, end)
}

const gunzipB64 = (b64: string): string => {
  try {
    const buf = Buffer.from(b64, 'base64')
    return zlib.gunzipSync(buf).toString('utf-8')
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err)
    throw new Error(`invalid gzip chunk payload: ${message}`)
  }
}

const sortedKeys = (m: Record<string, unknown>): string[] => Object.keys(m).sort()

const compareMaps = <T>(
  label: string,
  left: Record<string, T>,
  right: Record<string, T>,
  eq: (a: T, b: T) => boolean
): string[] => {
  const errors: string[] = []
  const leftKeys = sortedKeys(left)
  const rightKeys = sortedKeys(right)
  if (leftKeys.join('\0') !== rightKeys.join('\0')) {
    errors.push(`${label}: key sets differ`)
    return errors
  }
  for (const key of leftKeys) {
    if (!eq(left[key], right[key])) errors.push(`${label}: ${key} differs`)
  }
  return errors
}

const main = (): void => {
  const [committedPath, rebuiltPath] = process.argv.slice(2)
  if (!committedPath || !rebuiltPath) {
    console.error(usage)
    process.exit(1)
  }

  const committed = read(committedPath)
  const rebuilt = read(rebuiltPath)
  const errors: string[] = []

  if (extractScalar(committed, 'VizbEntryKey') !== extractScalar(rebuilt, 'VizbEntryKey')) {
    errors.push('VizbEntryKey differs')
  }

  errors.push(
    ...compareMaps(
      'VizbChartRoots',
      extractStringMap(committed, 'VizbChartRoots'),
      extractStringMap(rebuilt, 'VizbChartRoots'),
      (a, b) => a === b
    )
  )

  errors.push(
    ...compareMaps(
      'VizbChunkImports',
      extractSliceMap(committed, 'VizbChunkImports'),
      extractSliceMap(rebuilt, 'VizbChunkImports'),
      (a, b) => a.join('\0') === b.join('\0')
    )
  )

  errors.push(
    ...compareMaps(
      'VizbChunks',
      extractStringMap(committed, 'VizbChunks'),
      extractStringMap(rebuilt, 'VizbChunks'),
      (a, b) => gunzipB64(a) === gunzipB64(b)
    )
  )

  if (extractHTMLTemplate(committed) !== extractHTMLTemplate(rebuilt)) {
    errors.push('VizbHTMLTemplate differs')
  }

  if (errors.length > 0) {
    console.error('pkg/template/vizb-ui.gen.go is out of sync with ui sources.')
    for (const err of errors) console.error(`  - ${err}`)
    console.error('Run: cd ui && EMBED_UI=True pnpm build && git add pkg/template/vizb-ui.gen.go')
    process.exit(1)
  }

  console.info('vizb-ui.gen.go chunk graph and template are in sync')
}

main()
