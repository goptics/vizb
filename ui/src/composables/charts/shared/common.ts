import { computed } from 'vue'
import type { Ref } from 'vue'
import type { SortOrder, ScaleType, ChartData, Sort } from '@/types'
import type { Point3D } from '@/types'
import { hasYAxis } from '@/lib/utils'

export const fontSize = 12

export const sortBy =
  <K extends string>(key: K) =>
  <T extends Record<K, number>>(sortOrder: SortOrder) => {
    if (sortOrder === 'asc') {
      return (a: T, b: T) => a[key] - b[key]
    }

    return (a: T, b: T) => b[key] - a[key]
  }

export const sortByTotal = sortBy('total')

export const sortByValue = sortBy('value')

// For line charts: use null for zero values to create gaps instead of dropping below axis
export const adjustForLogScaleLine = (value: number | null, scale: ScaleType): number | null => {
  if (value === null) return null
  if (scale !== 'log') return value
  return value <= 0 ? null : value
}

// Series order for bar/line charts. The sort now happens in the transform worker
// (see `sortSeriesByTotal` in lib/transform.ts), so the series arrive in final
// order — this just derives the x-axis category list and the has-y flag. `sort`
// is kept in the signature (callers pass it) but is no longer re-sorted here;
// removing the per-recompute O(n log n) over up to 100k series off the main
// thread is the point.
export function useSortedSeriesData(chartData: Ref<ChartData>, _sort: Ref<Sort>) {
  return computed(() => ({
    series: chartData.value.series,
    xAxisData: chartData.value.series.map((s) => s.xAxis),
    hasYAxis: hasYAxis(chartData),
  }))
}

/** Log needs a positive domain; fall back to linear when none exists. */
export function resolveLogScale(scale: ScaleType, values: Iterable<number | null>): ScaleType {
  if (scale !== 'log') return scale
  for (const v of values) {
    if (v !== null && v > 0) return 'log'
  }
  return 'linear'
}

// Sum each named series' data values — used for tooltip axis-sum display. Data
// items are plain numbers (or null for log gaps), coerced `[x, y]` pairs, or
// the legacy `{ value }` item shape.
type SeriesDatum = number | null | [number, number | null] | { value?: number }
export function computeSeriesTotals(
  series: Array<{ name: string; data: SeriesDatum[] }>
): Map<string, number> {
  const valueOf = (d: SeriesDatum): number => {
    if (typeof d === 'number') return d
    if (Array.isArray(d)) return typeof d[1] === 'number' ? d[1] : 0
    return d?.value ?? 0
  }
  return new Map(
    series.map((s) => [s.name, s.data.reduce<number>((sum, d) => sum + valueOf(d), 0)])
  )
}

// Sort category values by their total across all points on the given axis
export function sortByAxisTotal(
  values: string[],
  key: 'xAxis' | 'yAxis' | 'zAxis',
  points: Point3D[],
  order: SortOrder
): string[] {
  const totals = new Map<string, number>()
  for (const p of points) totals.set(p[key], (totals.get(p[key]) ?? 0) + p.value)
  return [...values].sort((a, b) => {
    const diff = (totals.get(a) ?? 0) - (totals.get(b) ?? 0)
    return order === 'asc' ? diff : -diff
  })
}
