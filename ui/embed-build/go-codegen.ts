import { CHART_ROOT_PREFIX } from './constants.ts'

export const chunkKeyOf = (file: string): string => `vizb:${file.replace(/\.js$/, '')}`

// Emit a Go `map[string]string{…}` body from a JS object. base64 chunk blobs and
// import-map keys are ASCII with no Go-string-breaking chars, so JSON.stringify
// yields valid Go double-quoted string literals.
const sortedEntries = <T>(m: Record<string, T>): Array<[string, T]> =>
  Object.entries(m).sort(([a], [b]) => a.localeCompare(b))

export const goStringMap = (m: Record<string, string>): string =>
  '{\n' +
  sortedEntries(m)
    .map(([k, v]) => `\t${JSON.stringify(k)}: ${JSON.stringify(v)},`)
    .join('\n') +
  '\n}'

// Emit a Go `map[string][]string{…}` body from a JS object of string arrays.
export const goStringSliceMap = (m: Record<string, string[]>): string =>
  '{\n' +
  sortedEntries(m)
    .map(([k, arr]) => {
      const refs = [...arr].sort((a, b) => a.localeCompare(b))
      return `\t${JSON.stringify(k)}: {${refs.map((s) => JSON.stringify(s)).join(', ')}},`
    })
    .join('\n') +
  '\n}'

// Tag gated renderer chunks (ChartBar-<hash>.js → "bar", …) so the Go pruner knows
// which chunk each logical chart resolves to.
export const detectChartRoots = (
  files: string[],
  prefixMap: Record<string, string> = CHART_ROOT_PREFIX
): Record<string, string> => {
  const roots: Record<string, string> = {}
  for (const file of files) {
    const fileKey = chunkKeyOf(file)
    for (const prefix of Object.keys(prefixMap)) {
      const chartName = prefixMap[prefix]
      if (!chartName) continue
      if (file.startsWith(`${prefix}-`) || file === `${prefix}.js`) {
        roots[chartName] = fileKey
      }
    }
  }
  return roots
}
