import type { LabelAffix } from '../types'

// Exact column-name match — no trimming or case folding (spec decision).
export function affixForColumn(affix: LabelAffix | undefined, column: string): string {
  return affix?.columns?.[column] ?? affix?.global ?? ''
}
