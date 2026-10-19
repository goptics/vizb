import type { Better } from '@/types'
import { LARGE_DATA_THRESHOLD } from './chartConfig'

/** Glyph that marks the best value on a chart. */
export const BEST_MARK = '★'
/** Glyph prefixed to the category axis label of the best group (2D bar). */
export const WINNER_MARK = '🏆'

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
  visible: boolean[]
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

/** Rows + legend visibility, or null when the chart can't be marked. */
function markable(
  seriesList: MarkableSeries[],
  better: Better | undefined,
  selected: Record<string, boolean>
): { rows: (number | null)[][]; visible: boolean[]; better: Better } | null {
  if (!better || !seriesList.length) return null
  // ponytail: ECharts `large` mode ignores per-item styles past this size anyway.
  if (seriesList.some((s) => s.data.length > LARGE_DATA_THRESHOLD)) return null
  const plain = seriesList.every((s) => s.data.every((d) => d == null || typeof d === 'number'))
  if (!plain) return null
  const rows = seriesList.map((s) => s.data as (number | null)[])
  const visible = seriesList.map((s) => selected[s.name ?? ''] !== false)
  return { rows, visible, better }
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
  const m = markable(seriesList, better, selected)
  if (!m) return
  const mask = bestMask(m.rows, m.better, m.visible)
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
            emphasis: {
              label: { formatter: (p: LabelParams) => `${BEST_MARK} ${base?.(p) ?? p.value}` },
            },
          }
        : d
    )
  })
}

/**
 * Category axis-label fragment that prefixes the best group with 🏆: the
 * category whose visible series total is lowest (`low`) or highest (`high`).
 * Needs two visible series (one series already gets a ★ on its best category).
 * Categories missing a visible value are skipped, so a gap can't win on `low`.
 * Tied categories all get it, unless every category ties. Call before
 * applyBestMarker (it needs plain data). Always returns a formatter: ECharts
 * merges option updates, so omitting it would keep a stale trophy.
 */
export function trophyAxisLabel(
  seriesList: MarkableSeries[],
  better: Better | undefined,
  selected: Record<string, boolean>
): { formatter: (value: string, index: number) => string } {
  const winners = bestCategories(seriesList, better, selected)
  return { formatter: (v, i) => (winners.has(i) ? `${WINNER_MARK} ${v}` : v) }
}

/**
 * Returns `axes` with trophyAxisLabel merged into the category axis (`key`)
 * label. Callers skip numeric-x (value axis) charts.
 */
export function withCategoryTrophy<A extends Record<'xAxis' | 'yAxis', any>>(
  axes: A,
  key: 'xAxis' | 'yAxis',
  seriesList: MarkableSeries[],
  better: Better | undefined,
  selected: Record<string, boolean>
): A {
  const trophy = trophyAxisLabel(seriesList, better, selected)
  return { ...axes, [key]: { ...axes[key], axisLabel: { ...axes[key].axisLabel, ...trophy } } }
}

function bestCategories(
  seriesList: MarkableSeries[],
  better: Better | undefined,
  selected: Record<string, boolean>
): Set<number> {
  const m = markable(seriesList, better, selected)
  const live = m ? m.rows.filter((_, i) => m.visible[i]) : []
  if (live.length < 2) return new Set()
  const totals = live[0]!.map((_, j) => {
    const vals = live.map((r) => r[j])
    return vals.every(finite) ? vals.reduce((a, b) => a + b, 0) : NaN
  })
  const eligible = totals.filter(Number.isFinite)
  const best = m!.better === 'low' ? Math.min(...eligible) : Math.max(...eligible)
  const winners = totals.flatMap((t, j) => (t === best ? [j] : []))
  return winners.length === eligible.length ? new Set() : new Set(winners)
}
