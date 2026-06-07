import { computed, unref, type MaybeRef, type Ref } from 'vue'
import type { AxisLabels, DataPoint, ChartData, Point3D, SeriesData, Stat } from '../types'

type StatSignature = `${Stat['type']}-${Stat['unit']}-${Stat['per']}`

const toStatSignature = (stat: Stat): StatSignature => {
  return `${stat.type}-${stat.unit}-${stat.per}`
}

export function useChartData(
  results: Ref<DataPoint[]> | DataPoint[],
  axisLabels?: MaybeRef<AxisLabels | undefined>
) {
  const chartData = computed<ChartData[]>(() => {
    const data = Array.isArray(results) ? results : results.value
    if (!data?.length) return []

    const labels = unref(axisLabels)

    // Collect all unique stat signatures
    const uniqueStats = data.reduce((acc, benchmark) => {
      for (const stat of benchmark.stats || []) {
        const signature = toStatSignature(stat)
        if (!acc.has(signature)) {
          acc.set(signature, stat)
        }
      }

      return acc
    }, new Map<StatSignature, Stat>())

    return Array.from(uniqueStats.entries()).map(([signature, statTemplate]) => {
      const dataMap = new Map<string, Map<string, number>>()
      const xAxisSet = new Set<string>()
      const yAxisSet = new Set<string>()
      const zAxisSet = new Set<string>()
      const points: Point3D[] = []

      for (const benchmarkData of data) {
        const { xAxis = '', yAxis = '', zAxis = '' } = benchmarkData

        // Find the matching stat for this benchmark
        const matchingStat = benchmarkData.stats?.find((s) => toStatSignature(s) === signature)

        const value = matchingStat?.value

        // Skip benchmarks that don't have the matching stat
        if (value === undefined) continue

        yAxisSet.add(yAxis)
        xAxisSet.add(xAxis)
        zAxisSet.add(zAxis)

        points.push({ xAxis, yAxis, zAxis, value })

        if (!dataMap.has(yAxis)) {
          dataMap.set(yAxis, new Map())
        }

        dataMap.get(yAxis)!.set(xAxis, value)
      }

      const xAxisValues = Array.from(xAxisSet)
      const yAxisValues = Array.from(yAxisSet)
      const zAxisValues = Array.from(zAxisSet)

      // Single workload case - sort by subject values
      const series: SeriesData[] = xAxisValues.map((xAxis) => ({
        xAxis,
        values: yAxisValues.map((yAxis) => dataMap.get(yAxis)?.get(xAxis) || 0),
        benchmarkId: data[0]?.name || '', // Add benchmark identifier
      }))

      return {
        title: statTemplate.type,
        statType: statTemplate.type,
        statUnit: statTemplate.unit,
        yAxis: yAxisValues,
        zAxis: zAxisValues,
        series,
        points,
        axisLabels: labels,
      }
    })
  })

  return { chartData }
}
