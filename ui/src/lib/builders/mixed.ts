import type { ChartData, DataPoint } from '@/types'
import type { ChartBuilder, BuildContext } from './types'
import { axisLabelsFromAxes, mixedModeHasZ } from '../transform'

// Mixed-axis scatter chart shape: category x + value y[,z] (solo --select mixed
// mode). Each row is one point; x is categorical, y/z continuous. Emits 2D
// mixedTuples against xCategories, or a 3D line render3D when z is present.
export class MixedBuilder implements ChartBuilder {
  build(data: DataPoint[], ctx: BuildContext): ChartData {
    const { axes } = ctx
    const scale = ctx.scale
    const labels = axisLabelsFromAxes(axes ?? [])
    const xLabel = labels.x ?? 'x'
    const yLabel = labels.y ?? 'y'
    const zLabel = labels.z ?? 'z'
    const use3D = mixedModeHasZ(axes ?? [])
    const title = use3D ? `${xLabel} · ${yLabel} · ${zLabel}` : `${xLabel} vs ${yLabel}`

    const xCategories: string[] = []
    const xIndex = new Map<string, number>()
    const mixedTuples: [number, number][] = []
    const points3D: { value: number[] }[] = []

    for (const row of data) {
      const x = row.xAxis ?? ''
      if (!x) continue

      const y = Number(row.yAxis)
      if (!isFinite(y)) continue

      if (!xIndex.has(x)) {
        xIndex.set(x, xCategories.length)
        xCategories.push(x)
      }
      const xi = xIndex.get(x)!

      if (use3D) {
        const z = Number(row.zAxis)
        if (!isFinite(z)) continue
        if (scale === 'log' && (y <= 0 || z <= 0)) continue
        points3D.push({ value: [xi, y, z] })
      } else {
        if (scale === 'log' && y <= 0) continue
        mixedTuples.push([xi, y])
      }
    }

    const chart: ChartData = {
      title,
      statType: 'mixed',
      yAxis: [],
      zAxis: [],
      series: [],
      points: [],
      axisLabels: labels,
      xCategories,
      ...(!use3D ? { mixedTuples } : {}),
    }

    if (use3D && points3D.length) {
      chart.render3D = {
        mode: 'mixed',
        xValues: xCategories,
        yValues: [],
        zValues: [],
        barSeries: [],
        lineSeries: [{ name: title, data: points3D }],
        cellTotals: {},
      }
    }

    return chart
  }

  plottable(chart: ChartData): boolean {
    if ((chart.mixedTuples?.length ?? 0) > 0) return true
    return chart.render3D?.mode === 'mixed' && (chart.render3D.lineSeries[0]?.data.length ?? 0) > 0
  }

  badgeCount(chart: ChartData, axis: 'x' | 'y' | 'z'): number {
    if (chart.mixedTuples?.length && chart.xCategories?.length) {
      if (axis === 'x') return chart.xCategories.length
      if (axis === 'z') return 0
      return chart.mixedTuples.length
    }
    if (axis === 'x') return chart.xCategories?.length ?? 0
    if (axis === 'z') return 0
    const pts = chart.render3D?.lineSeries[0]?.data ?? []
    return new Set(pts.map((p) => p.value[1])).size
  }

  grandTotal(chart: ChartData): number {
    if (chart.mixedTuples?.length) {
      return chart.mixedTuples.reduce((sum, [, y]) => sum + y, 0)
    }
    if (chart.render3D?.mode === 'mixed') {
      return (chart.render3D.lineSeries[0]?.data ?? []).reduce(
        (sum, p) => sum + (p.value[1] ?? 0),
        0
      )
    }
    return 0
  }

  is3D(chart: ChartData): boolean {
    return chart.render3D?.mode === 'mixed'
  }
}
