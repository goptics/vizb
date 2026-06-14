// Framework-free CSV builders for the stats panel export. Pure string in/out (no
// DOM, no Blob) so they're unit-testable; the panel wraps the result in a Blob
// download. Full numeric precision is preserved (raw String(v), not the panel's
// display formatter) and missing/non-finite numbers become empty cells.
import type { DescriptiveStats, SeriesProfile } from '../types'

// RFC-4180 quoting: wrap a field in double quotes and double any internal quote
// when it contains a comma, quote, newline, or edge whitespace. Non-finite
// numbers (NaN/±Inf) render as an empty cell so spreadsheets read them as blank,
// not as the literal text "NaN".
export function toCsvCell(v: string | number): string {
  if (typeof v === 'number') {
    if (!Number.isFinite(v)) return ''
    return String(v)
  }
  if (/[",\n]/.test(v) || v !== v.trim()) {
    return `"${v.replace(/"/g, '""')}"`
  }
  return v
}

const row = (cells: (string | number)[]): string => cells.map(toCsvCell).join(',')

// Descriptive table → CSV. Header is `Series` + each column label; one row per
// series with its raw stat values in column order.
export function descriptiveCsv(
  profiles: SeriesProfile[],
  columns: { key: keyof DescriptiveStats; label: string }[]
): string {
  const header = row(['Series', ...columns.map((c) => c.label)])
  const lines = profiles.map((p) => row([p.name, ...columns.map((c) => p.stats[c.key])]))
  return [header, ...lines].join('\n')
}

// Correlation matrix → CSV. Leading empty corner cell, then the series labels as
// the header; each row is a label followed by its matrix values.
export function correlationCsv(labels: string[], matrix: number[][]): string {
  const header = row(['', ...labels])
  const lines = labels.map((label, i) => row([label, ...(matrix[i] ?? [])]))
  return [header, ...lines].join('\n')
}
