import type { ChartData, DataPoint, SeriesData, Point3D } from '@/types'
import type { ChartBuilder, BuildContext } from './types'
import { finalizeChart } from './finalize'
import { statsForSignature } from '../transform'

// PreserveRows chart shape: one row per data point (no averaging across
// duplicate (x,y)). When all y values are empty, falls back to a category
// scatter (mixedTuples against xCategories); otherwise emits one series per
// row with null-padded values aligned to the first-seen y order.
export class PreserveRowsBuilder implements ChartBuilder {
  build(data: DataPoint[], ctx: BuildContext): ChartData {
    const { signature, statTemplate, labels } = ctx
    const xAxisSet = new Set<string>()
    const yAxisSet = new Set<string>()
    const zAxisSet = new Set<string>()
    const points: Point3D[] = []

    const yOrder: string[] = []
    const ySeen = new Set<string>()

    for (const benchmarkData of data) {
      const { xAxis = '', yAxis = '', zAxis = '' } = benchmarkData
      for (const matchingStat of statsForSignature(benchmarkData.stats, signature)) {
        const value = matchingStat.value
        if (value === undefined) continue

        yAxisSet.add(yAxis)
        xAxisSet.add(xAxis)
        zAxisSet.add(zAxis)
        points.push({ xAxis, yAxis, zAxis, value })

        if (!ySeen.has(yAxis)) {
          ySeen.add(yAxis)
          yOrder.push(yAxis)
        }
      }
    }

    const yAxisValues = yOrder.length ? yOrder : Array.from(yAxisSet)
    const useCategoryScatter =
      yAxisValues.length === 0 || (yAxisValues.length === 1 && yAxisValues[0] === '')

    let series: SeriesData[] = []
    let mixedTuples: [number, number][] | undefined
    let xCategories: string[] | undefined

    if (useCategoryScatter) {
      const xIndex = new Map<string, number>()
      const cats: string[] = []
      const tuples: [number, number][] = []

      for (const benchmarkData of data) {
        const { xAxis = '' } = benchmarkData
        for (const matchingStat of statsForSignature(benchmarkData.stats, signature)) {
          const value = matchingStat.value
          if (value === undefined) continue

          if (!xIndex.has(xAxis)) {
            xIndex.set(xAxis, cats.length)
            cats.push(xAxis)
          }
          tuples.push([xIndex.get(xAxis)!, value])
        }
      }

      xCategories = cats
      mixedTuples = tuples
    } else {
      for (const benchmarkData of data) {
        const { xAxis = '', yAxis = '' } = benchmarkData
        for (const matchingStat of statsForSignature(benchmarkData.stats, signature)) {
          const value = matchingStat.value
          if (value === undefined) continue

          series.push({
            xAxis,
            values: yAxisValues.map((y) => (y === yAxis ? value : null)),
            benchmarkId: benchmarkData.name || '',
          })
        }
      }
    }

    return finalizeChart(
      {
        statType: statTemplate.type,
        statUnit: statTemplate.unit,
        title: statTemplate.type,
        yAxisValues,
        zAxisValues: Array.from(zAxisSet),
        series,
        points,
        axisLabels: labels,
        xSet: xAxisSet,
        mixedTuples,
        xCategories,
      },
      ctx
    )
  }

  plottable(chart: ChartData): boolean {
    return (
      chart.series.length > 0 || chart.points.length > 0 || (chart.mixedTuples?.length ?? 0) > 0
    )
  }

  badgeCount(chart: ChartData, axis: 'x' | 'y' | 'z'): number {
    if (chart.mixedTuples?.length && chart.xCategories?.length) {
      if (axis === 'x') return chart.xCategories.length
      if (axis === 'z') return 0
      return chart.mixedTuples.length
    }
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
    if (chart.mixedTuples?.length) {
      return chart.mixedTuples.reduce((sum, [, y]) => sum + y, 0)
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

  canOfferValue3D(): boolean {
    return false
  }
}
