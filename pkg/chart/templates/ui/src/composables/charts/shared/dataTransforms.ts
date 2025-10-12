import type { ChartData, SortOrder, SeriesData } from '../../../types/benchmark'

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
    .sort((a: { subject: string; total: number }, b: { subject: string; total: number }) => {
      if (sortOrder === "asc") return a.total - b.total
      return b.total - a.total
    })
    .map((item: { subject: string; total: number }) => item.subject)

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

  seriesWithTotals.sort((a: SeriesWithTotal, b: SeriesWithTotal) => {
    if (sortOrder === "asc") return (a.total || 0) - (b.total || 0)
    return (b.total || 0) - (a.total || 0)
  })

  return {
    series: seriesWithTotals,
    xAxisData: chartData.workloads,
  }
}

/**
 * Creates pie chart data from series with totals
 */
export function createPieData(
  series: SeriesWithTotal[],
  sortOrder: SortOrder
): Array<{ name: string; value: number; total: number }> {
  const pieData = series.map((seriesData) => ({
    name: seriesData.subject,
    value: seriesData.total || 0,
    total: seriesData.total || 0,
  }))

  if (sortOrder !== "") {
    pieData.sort((a, b) => {
      if (sortOrder === "asc") return a.total - b.total
      return b.total - a.total
    })
  }

  return pieData
}

/**
 * Creates pie chart data from subject totals
 */
export function createSubjectPieData(
  subjectTotals: Record<string, number>,
  sortOrder: SortOrder
): Array<{ name: string; value: number; total: number }> {
  const pieData = Object.entries(subjectTotals).map(([subject, total]) => ({
    name: subject,
    value: total,
    total,
  }))

  if (sortOrder !== "") {
    pieData.sort((a, b) => {
      if (sortOrder === "asc") return a.total - b.total
      return b.total - a.total
    })
  }

  return pieData
}
