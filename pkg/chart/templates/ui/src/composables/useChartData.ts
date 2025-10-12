import { computed, type Ref } from 'vue'
import type { BenchmarkResult, ChartData, SeriesData } from '../types/benchmark'

export function useChartData(results: Ref<BenchmarkResult[]> | BenchmarkResult[]) {
  const chartData = computed<ChartData[]>(() => {
    const resultsList = Array.isArray(results) ? results : results.value
    if (!resultsList?.length) return []

    const firstResult = resultsList[0]
    if (!firstResult?.stats) return []

    return firstResult.stats.map((stat, statIndex) => {
      const dataMap = new Map<string, Map<string, number>>()
      const workloadsSet = new Set<string>()
      const subjectsSet = new Set<string>()

      resultsList.forEach(result => {
        const workload = result.workload || ''
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
      let subjects = Array.from(subjectsSet)

      // When we have workloads, create series per workload (workload becomes legend)
      // Otherwise, create series per subject (subject becomes legend)
      const hasMultipleWorkloads = workloads.length > 1

      // When we have multiple workloads, we need to keep subjects in original order
      // because sorting will be handled in useEChartOptions based on sort settings
      if (hasMultipleWorkloads) {
        // Calculate total value for each subject across all workloads
        const subjectTotals = subjects.map(subject => {
          const total = workloads.reduce((sum, workload) => {
            return sum + (dataMap.get(workload)?.get(subject) || 0)
          }, 0)
          return { subject, total }
        })

        // Store totals for sorting but don't sort here - let chart options handle it
        const series: SeriesData[] = workloads.map(workload => ({
          subject: workload,
          values: subjects.map(subject => dataMap.get(workload)?.get(subject) || 0),
          subjectTotals // Pass totals for sorting in chart options
        }))

        const title = stat.unit ? `${stat.type} (${stat.unit}/op)` : `${stat.type}/op`

        return {
          title,
          statType: stat.type,
          statUnit: stat.unit,
          workloads: subjects,
          series,
          subjectTotals: subjectTotals.reduce((acc, { subject, total }) => {
            acc[subject] = total
            return acc
          }, {} as Record<string, number>)
        }
      }

      // Single workload case - sort by subject values
      const series: SeriesData[] = subjects.map(subject => ({
        subject,
        values: workloads.map(workload => dataMap.get(workload)?.get(subject) || 0)
      }))

      const title = stat.unit ? `${stat.type} (${stat.unit}/op)` : `${stat.type}/op`

      return {
        title,
        statType: stat.type,
        statUnit: stat.unit,
        workloads,
        series
      }
    })
  })

  return { chartData }
}
