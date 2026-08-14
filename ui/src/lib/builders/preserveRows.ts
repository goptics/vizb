import type { ChartData, DataPoint, SeriesData, Point3D } from '@/types'
import type { GroupingBuilder, BuildContext } from './types'
import { finalizeChart } from './finalize'
import { groupedQueries } from './grouped'
import { statsForSignature } from '../transform'

// PreserveRows chart shape: one row per data point (no averaging across
// duplicate (x,y)). When all y values are empty, falls back to a category
// scatter (mixedTuples against xCategories); otherwise emits one series per
// row with null-padded values aligned to the first-seen y order.
export class PreserveRowsBuilder implements GroupingBuilder {
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

  badgeCount = groupedQueries.badgeCount
  grandTotal = groupedQueries.grandTotal
  is3D = groupedQueries.is3D

  canOfferValue3D(): boolean {
    return false
  }
}
