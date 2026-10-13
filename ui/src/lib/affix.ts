import type { LabelAffix } from '../types'

// Exact column-name match — no trimming or case folding (spec decision).
export function affixForColumn(affix: LabelAffix | undefined, column: string): string {
  return affix?.columns?.[column] ?? affix?.global ?? ''
}

// Affixes are user text; ECharts renders tooltip formatter strings as HTML.
// Escape before interpolating into tooltip markup (canvas labels/axes stay raw).
export function escapeHtml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}
