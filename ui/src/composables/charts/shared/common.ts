import { computed } from 'vue'
import type { Ref } from 'vue'
import type { SortOrder, ScaleType, ChartData, SeriesData, Sort } from '../../../types'
import type { Point3D } from '../../../types'
import { hasYAxis } from '../../../lib/utils'

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
export const adjustForLogScaleLine = (value: number, scale: ScaleType): number | null => {
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

// Compute effective scale and min value — log is downgraded to linear when max < 1.
// Single pass (no `Math.min(...arr)` spread — that both allocates an intermediate
// array and throws RangeError once the arg count is large, e.g. 100k points).
export function getEffectiveScale(
  series: SeriesData[],
  scale: ScaleType
): { minValue: number | undefined; effectiveScale: ScaleType } {
  let minValue: number | undefined
  let maxValue = 0
  let any = false
  for (const s of series) {
    for (const v of s.values) {
      if (!any) {
        maxValue = v
        any = true
      } else if (v > maxValue) {
        maxValue = v
      }
      if (v > 0 && (minValue === undefined || v < minValue)) minValue = v
    }
  }
  return { minValue, effectiveScale: scale === 'log' && maxValue < 1 ? 'linear' : scale }
}

// Sum each named series' data values — used for tooltip axis-sum display. Data
// items are plain numbers (or null for log gaps); also tolerates the legacy
// `{ value }` item shape.
type SeriesDatum = number | null | { value?: number }
export function computeSeriesTotals(
  series: Array<{ name: string; data: SeriesDatum[] }>
): Map<string, number> {
  const valueOf = (d: SeriesDatum): number => (typeof d === 'number' ? d : (d?.value ?? 0))
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
