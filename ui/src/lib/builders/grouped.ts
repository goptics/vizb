import type { ChartData, DataPoint, SeriesData, Point3D } from '@/types'
import type { ChartBuilder, BuildContext } from './types'
import { finalizeChart } from './finalize'
import { toStatSignature } from '../transform'

// Grouped (non-preserveRows) chart shape: one series per x-axis value, values
// averaged per (yAxis, xAxis) cell. This is the default grouped-bar/line path.
export class GroupedBuilder implements ChartBuilder {
  build(data: DataPoint[], ctx: BuildContext): ChartData {
    const { signature, statTemplate, labels } = ctx
    const xAxisSet = new Set<string>()
    const yAxisSet = new Set<string>()
    const zAxisSet = new Set<string>()
    const points: Point3D[] = []

    const dataMap = new Map<string, Map<string, number>>()
    const countMap = new Map<string, Map<string, number>>()

    for (const benchmarkData of data) {
      const { xAxis = '', yAxis = '', zAxis = '' } = benchmarkData
      const matchingStat = benchmarkData.stats?.find((s) => toStatSignature(s) === signature)
      const value = matchingStat?.value
      if (value === undefined) continue

      yAxisSet.add(yAxis)
      xAxisSet.add(xAxis)
      zAxisSet.add(zAxis)
      points.push({ xAxis, yAxis, zAxis, value })

      if (!dataMap.has(yAxis)) {
        dataMap.set(yAxis, new Map())
        countMap.set(yAxis, new Map())
      }
      const yMap = dataMap.get(yAxis)!
      const cMap = countMap.get(yAxis)!
      yMap.set(xAxis, (yMap.get(xAxis) ?? 0) + value)
      cMap.set(xAxis, (cMap.get(xAxis) ?? 0) + 1)
    }

    for (const [yAxis, xMap] of dataMap) {
      const cMap = countMap.get(yAxis)!
      for (const [xAxis, sum] of xMap) xMap.set(xAxis, sum / cMap.get(xAxis)!)
    }

    const yAxisValues = Array.from(yAxisSet)
    const xAxisValuesAgg = Array.from(xAxisSet)
    const builtSeries: SeriesData[] = xAxisValuesAgg.map((xAxis) => ({
      xAxis,
      values: yAxisValues.map((yAxis) => dataMap.get(yAxis)?.get(xAxis) ?? null),
      benchmarkId: data[0]?.name || '',
    }))

    return finalizeChart(
      {
        statType: statTemplate.type,
        statUnit: statTemplate.unit,
        title: statTemplate.type,
        yAxisValues,
        zAxisValues: Array.from(zAxisSet),
        series: builtSeries,
        points,
        axisLabels: labels,
        xSet: xAxisSet,
      },
      ctx
    )
  }

  plottable(chart: ChartData): boolean {
    return chart.series.length > 0 || chart.points.length > 0
  }

  badgeCount(chart: ChartData, axis: 'x' | 'y' | 'z'): number {
    if (axis === 'x') return chart.series.length
    if (axis === 'y') return chart.yAxis.length
    return chart.zAxis.length
  }

  grandTotal(chart: ChartData, visibleZ?: Record<string, boolean>): number {
    if (chart.points.length > 0) {
      const filterZ = chart.zAxis.length > 0 && chart.zAxis[0] !== ''
      let total = 0
      for (const pt of chart.points) {
        if (filterZ && pt.zAxis && visibleZ?.[pt.zAxis] === false) continue
        total += pt.value
      }
      return total
    }
    let total = 0
    for (const s of chart.series) {
      for (const v of s.values) {
        if (v != null) total += v
      }
    }
    return total
  }

  is3D(chart: ChartData, cfg?: { threeD?: boolean }): boolean {
    const hasX = chart.series.some((s) => s.xAxis && s.xAxis.trim() !== '')
    const hasY = chart.yAxis.length > 0 && chart.yAxis[0] !== ''
    const hasZ = chart.zAxis.length > 0 && chart.zAxis[0] !== ''
    return (hasX && hasY && hasZ) || (hasX && hasY && !hasZ && cfg?.threeD === true)
  }
}
