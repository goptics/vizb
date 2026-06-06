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

// Shared sorted-series computation for bar/line charts (identical logic in both)
export function useSortedSeriesData(chartData: Ref<ChartData>, sort: Ref<Sort>) {
  return computed(() => {
    if (!sort.value.enabled) {
      return {
        series: chartData.value.series,
        xAxisData: chartData.value.series.map((s) => s.xAxis),
        hasYAxis: hasYAxis(chartData),
      }
    }

    const seriesWithTotals = chartData.value.series.map((series) => ({
      ...series,
      total: series.values.reduce((sum, val) => sum + val, 0),
    }))

    seriesWithTotals.sort(sortByTotal(sort.value.order))

    return {
      series: seriesWithTotals,
      xAxisData: seriesWithTotals.map((s) => s.xAxis),
      hasYAxis: hasYAxis(chartData),
    }
  })
}

// Compute effective scale and min value — log is downgraded to linear when max < 1
export function getEffectiveScale(
  series: SeriesData[],
  scale: ScaleType
): { minValue: number | undefined; effectiveScale: ScaleType } {
  const allValues = series.flatMap((s) => s.values)
  const nonZeroValues = allValues.filter((v) => v > 0)
  const minValue = nonZeroValues.length > 0 ? Math.min(...nonZeroValues) : undefined
  const maxValue = allValues.length > 0 ? Math.max(...allValues) : 0
  return { minValue, effectiveScale: scale === 'log' && maxValue < 1 ? 'linear' : scale }
}

// Sum each named series' data values — used for tooltip axis-sum display
export function computeSeriesTotals(
  series: Array<{ name: string; data: Array<{ value?: number } | null> }>
): Map<string, number> {
  return new Map(
    series.map((s) => [
      s.name,
      s.data.reduce((sum, d) => sum + (d?.value ?? 0), 0),
    ])
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
