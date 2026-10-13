import type { LabelAffix } from '../types'

// Exact column-name match — no trimming or case folding (spec decision).
export function affixForColumn(affix: LabelAffix | undefined, column: string): string {
  return affix?.columns?.[column] ?? affix?.global ?? ''
}

/** Both label affix configs (wire labelPrefix / labelSuffix). */
export interface Affixes {
  prefix?: LabelAffix
  suffix?: LabelAffix
}

/** Resolved prefix/suffix text for one column. */
export interface Wrap {
  prefix: string
  suffix: string
}

export const NO_WRAP: Wrap = { prefix: '', suffix: '' }

/** Affixes from a chart config's `labelPrefix` / `labelSuffix` refs. */
export const affixesOf = (c: {
  labelPrefix?: { value?: LabelAffix }
  labelSuffix?: { value?: LabelAffix }
}): Affixes => ({ prefix: c.labelPrefix?.value, suffix: c.labelSuffix?.value })

/** Resolve both affixes for a column; a per-side `fallback` applies when that side has no entry. */
export function wrapFor(a: Affixes | undefined, column: string, fallback: Wrap = NO_WRAP): Wrap {
  return {
    prefix: affixForColumn(a?.prefix, column) || fallback.prefix,
    suffix: affixForColumn(a?.suffix, column) || fallback.suffix,
  }
}

/** Wrap a rendered value with its prefix and suffix: `prefix + value + suffix`. */
export function applyAffixes(value: string, w: Wrap): string {
  return w.prefix + value + w.suffix
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

/** applyAffixes for tooltip HTML: both affixes are escaped. */
export function tooltipAffixes(value: string, w: Wrap): string {
  return escapeHtml(w.prefix) + value + escapeHtml(w.suffix)
}
