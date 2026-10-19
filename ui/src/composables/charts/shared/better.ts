import type { Better } from '@/types'
import { LARGE_DATA_THRESHOLD } from './chartConfig'

/** Glyph that marks the best value on a chart. */
export const BEST_MARK = '★'

/** Human hint for a metric direction; empty when unknown. */
export function betterHint(better?: Better): string {
  if (better === 'low') return 'Lower is better'
  if (better === 'high') return 'Higher is better'
  return ''
}

const finite = (v: number | null | undefined): v is number =>
  typeof v === 'number' && Number.isFinite(v)

/**
 * Marks the best value per category across visible rows (one row per chart
 * series, one column per category). With a single visible row the best
 * category is marked instead. Ties are all marked; nothing is marked unless at
 * least two finite values compete.
 */
export function bestMask(
  rows: (number | null | undefined)[][],
  better: Better,
  visible: boolean[] = rows.map(() => true)
): boolean[][] {
  const mask = rows.map((r) => r.map(() => false))
  const pick = (vals: number[]) => (better === 'low' ? Math.min(...vals) : Math.max(...vals))
  const live = rows.map((_, i) => i).filter((i) => visible[i])

  if (live.length === 1) {
    const i = live[0]!
    const vals = rows[i]!.filter(finite)
    if (vals.length < 2) return mask
    const best = pick(vals)
    rows[i]!.forEach((v, j) => (mask[i]![j] = v === best))
    return mask
  }

  const cols = Math.max(0, ...rows.map((r) => r.length))
  for (let j = 0; j < cols; j++) {
    const vals = live.map((i) => rows[i]![j]).filter(finite)
    if (vals.length < 2) continue
    const best = pick(vals)
    for (const i of live) if (rows[i]![j] === best) mask[i]![j] = true
  }
  return mask
}

type LabelParams = { value?: number | (number | null)[] | null }
type LabelFormatter = (params: LabelParams) => string
type MarkableSeries = {
  name?: string
  data: unknown[]
  label?: { formatter?: LabelFormatter }
}

/**
 * Replaces the best data points of bar/line series with styled items: a
 * contrasting border and a ★ label (prefixed to the value label when labels
 * are on). Series whose data is not plain numbers (e.g. [x, y] tuples) are
 * left alone. `selected` is the legend state; missing names count as visible.
 */
export function applyBestMarker(
  seriesList: MarkableSeries[],
  better: Better | undefined,
  selected: Record<string, boolean>,
  opts: { isDark: boolean; showLabels: boolean }
): void {
  if (!better || !seriesList.length) return
  // ponytail: ECharts `large` mode ignores per-item styles past this size anyway.
  if (seriesList.some((s) => s.data.length > LARGE_DATA_THRESHOLD)) return
  const plain = seriesList.every((s) => s.data.every((d) => d == null || typeof d === 'number'))
  if (!plain) return

  const rows = seriesList.map((s) => s.data as (number | null)[])
  const visible = seriesList.map((s) => selected[s.name ?? ''] !== false)
  const mask = bestMask(rows, better, visible)
  // Gold reads as "winner"; deeper in light mode for contrast on white, and a
  // glow keeps the outline visible on orange/yellow series bars.
  const gold = opts.isDark ? '#fbbf24' : '#d97706'
  const glow = opts.isDark ? 'rgba(251,191,36,0.55)' : 'rgba(217,119,6,0.45)'

  seriesList.forEach((s, i) => {
    const base = s.label?.formatter
    s.data = s.data.map((d, j) =>
      mask[i]![j]
        ? {
            value: d,
            itemStyle: { borderColor: gold, borderWidth: 2, shadowColor: glow, shadowBlur: 8 },
            label: {
              show: true,
              color: gold,
              fontWeight: 'bold',
              fontSize: 14,
              // Dark outline keeps the gold star legible on orange/yellow bars.
              textBorderColor: 'rgba(0,0,0,0.55)',
              textBorderWidth: 2,
              formatter: (p: LabelParams) =>
                opts.showLabels && base ? `${BEST_MARK} ${base(p)}` : BEST_MARK,
            },
            // Hover labels (labels off) still show the value next to the star.
            ...(base
              ? {
                  emphasis: { label: { formatter: (p: LabelParams) => `${BEST_MARK} ${base(p)}` } },
                }
              : {}),
          }
        : d
    )
  })
}
