import { computed, type Ref } from 'vue'
import type { BenchmarkResult, ChartData, SeriesData } from '../types/benchmark'

/**
 * Composable for processing benchmark results into chart data
 * Groups results by stat type to create consolidated charts
 */
export function useChartData(results: Ref<BenchmarkResult[]> | BenchmarkResult[]) {
  /**
   * Process results into chart data
   * Creates one chart per stat type with all subjects as series
   */
  const chartData = computed<ChartData[]>(() => {
    const resultsList = Array.isArray(results) ? results : results.value

    if (!resultsList || resultsList.length === 0) return []

    // Get all unique stat types from first result
    const firstResult = resultsList[0]
    if (!firstResult || !firstResult.stats) return []

    const charts: ChartData[] = []

    // Create a chart for each stat type
    firstResult.stats.forEach((stat, statIndex) => {
      // Group data by workload and subject
      const dataMap = new Map<string, Map<string, number>>()
      const workloadsSet = new Set<string>()
      const subjectsSet = new Set<string>()

      resultsList.forEach(result => {
        const workload = result.workload || 'Default'
        const subject = result.subject
        const value = result.stats[statIndex]?.value || 0

        workloadsSet.add(workload)
        subjectsSet.add(subject)

        if (!dataMap.has(workload)) {
          dataMap.set(workload, new Map())
        }
        dataMap.get(workload)!.set(subject, value)
      })

      const workloads = Array.from(workloadsSet)
      const subjects = Array.from(subjectsSet)

      // Create series for each subject
      const series: SeriesData[] = subjects.map(subject => {
        const values = workloads.map(workload => {
          return dataMap.get(workload)?.get(subject) || 0
        })
        return { subject, values }
      })

      // Create chart title
      const title = stat.unit
        ? `${stat.type} (${stat.unit}/op)`
        : `${stat.type}/op`

      charts.push({
        title,
        statType: stat.type,
        statUnit: stat.unit,
        workloads,
        series
      })
    })

    return charts
  })

  return {
    chartData
  }
}
