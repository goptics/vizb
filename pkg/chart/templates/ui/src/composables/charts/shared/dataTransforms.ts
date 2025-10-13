import type { ChartData, SortOrder, SeriesData } from '../../../types/benchmark'
import { sortByTotal } from './common'

export type SeriesWithTotal = SeriesData & {
  total?: number
}

export interface SortedData {
  series: SeriesWithTotal[]
  xAxisData: string[]
}

/**
 * Transforms and sorts chart data based on sort order
 * Handles both single workload and multiple workload scenarios
 */
export function transformAndSortData(
  chartData: ChartData,
  sortOrder: SortOrder
): SortedData {
  if (sortOrder === "") {
    return {
      series: chartData.series,
      xAxisData: chartData.workloads,
    }
  }

  const hasSubjectTotals = chartData.subjectTotals !== undefined

  if (hasSubjectTotals) {
    return sortBySubjectTotals(chartData, sortOrder)
  }

  return sortBySeriesTotals(chartData, sortOrder)
}

/**
 * Sorts data when we have subject totals (multiple workloads case)
 */
function sortBySubjectTotals(chartData: ChartData, sortOrder: SortOrder): SortedData {
  const sortedSubjects = chartData.workloads
    .map((subject: string) => ({
      subject,
      total: chartData.subjectTotals![subject] || 0,
    }))
    .sort(sortByTotal(sortOrder))
    .map((item) => item.subject)

  const subjectIndexMap = new Map(
    chartData.workloads.map((subject: string, idx: number) => [subject, idx])
  )

  const sortedSeries = chartData.series.map((series: SeriesData) => ({
    ...series,
    values: sortedSubjects.map((subject: string) => {
      const idx = subjectIndexMap.get(subject)
      return idx !== undefined ? (series.values[idx] ?? 0) : 0
    }),
  }))

  return {
    series: sortedSeries,
    xAxisData: sortedSubjects,
  }
}

/**
 * Sorts data when we have single workload (sort series by their totals)
 */
function sortBySeriesTotals(chartData: ChartData, sortOrder: SortOrder): SortedData {
  const seriesWithTotals = chartData.series.map((series: SeriesData) => ({
    ...series,
    total: series.values.reduce((sum: number, val: number) => sum + val, 0),
  }))
    .sort(sortByTotal(sortOrder))

  return {
    series: seriesWithTotals,
    xAxisData: chartData.workloads,
  }
}
