import type { LabelAffix } from '../types'

// Exact column-name match — no trimming or case folding (spec decision).
export function affixForColumn(affix: LabelAffix | undefined, column: string): string {
  return affix?.columns?.[column] ?? affix?.global ?? ''
}

/** Wrap a rendered value with its prefix and suffix: `prefix + value + suffix`. */
export function applyAffixes(value: string, prefix: string, suffix: string): string {
  return prefix + value + suffix
}
