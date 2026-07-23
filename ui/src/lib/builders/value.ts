import type { ChartData, DataPoint } from '@/types'
import type { ChartBuilder, BuildContext } from './types'
import { arrangementHasChartZ, swapAxisLabels } from '../swap'
import {
  axisLabelsFromAxes,
  identityStringFromAxes,
  projectValueCoords,
  isSourceFieldOnChart,
  buildValueMode3DRender,
  type ValuePoint3D,
} from '../transform'

// Value-mode chart shape: continuous numeric axes (--axes x,y[,z]). Coordinates
// are projected onto chart axes per the active swap; swap with z on a chart axis
// yields continuous 3D, otherwise 2D valueTuples for chart x vs y.
export class ValueBuilder implements ChartBuilder {
  build(data: DataPoint[], ctx: BuildContext): ChartData {
    const { axes, identityString, targetString } = ctx
    const identity = identityString ?? (axes ? identityStringFromAxes(axes) : '')
    const target = targetString ?? identity
    const scale = ctx.scale
    const baseLabels = axisLabelsFromAxes(axes ?? [])
    const labels = { ...(swapAxisLabels(identity, target, baseLabels) ?? baseLabels) }
    const use3D = ctx.threeD && arrangementHasChartZ(target)

    const valueTuples: [number, number, number?][] = []
    const valuePoints3D: ValuePoint3D[] = []

    for (const row of data) {
      const coords = projectValueCoords(row, identity, target)
      if (!coords) continue

      if (use3D) {
        const cx = coords.xAxis
        const cy = coords.yAxis
        const cz = coords.zAxis
        if (cx === undefined || cy === undefined || cz === undefined) continue
        if (scale === 'log' && (cx <= 0 || cy <= 0 || cz <= 0)) continue
        const metricRaw = row.metric
        const metricNum =
          metricRaw !== undefined && metricRaw !== '' ? Number(metricRaw) : undefined
        if (metricNum !== undefined && isFinite(metricNum)) {
          valuePoints3D.push([cx, cy, cz, metricNum])
        } else {
          valuePoints3D.push([cx, cy, cz])
        }
      } else {
        const cx = coords.xAxis
        const cy = coords.yAxis
        if (cx === undefined || cy === undefined) continue
        if (scale === 'log' && cy <= 0) continue

        let colorDim: number | undefined
        const metricRaw = row.metric
        const metricNum =
          metricRaw !== undefined && metricRaw !== '' ? Number(metricRaw) : undefined
        if (metricNum !== undefined && isFinite(metricNum)) {
          colorDim = metricNum
        } else if (!isSourceFieldOnChart(identity, target, 'zAxis')) {
          const zNum = Number(row.zAxis)
          if (row.zAxis !== undefined && row.zAxis !== '' && isFinite(zNum)) {
            colorDim = zNum
          }
        }

        valueTuples.push(colorDim !== undefined ? [cx, cy, colorDim] : [cx, cy])
      }
    }

    const xLabel = labels.x ?? 'x'
    const yLabel = labels.y ?? 'y'
    const zLabel = labels.z ?? 'z'
    const title = use3D ? `${xLabel} · ${yLabel} · ${zLabel}` : `${xLabel} vs ${yLabel}`

    const chart: ChartData = {
      title,
      statType: 'value',
      yAxis: [],
      zAxis: [],
      series: [],
      points: [],
      axisLabels: labels,
      ...(!use3D ? { valueTuples } : {}),
      ...(valuePoints3D.length ? { valuePoints3D } : {}),
    }

    if (use3D && valuePoints3D.length) {
      chart.render3D = buildValueMode3DRender(valuePoints3D, title, ctx.showLabels, scale)
    }

    return chart
  }

  plottable(chart: ChartData): boolean {
    return (chart.valueTuples?.length ?? 0) > 0 || (chart.valuePoints3D?.length ?? 0) > 0
  }

  badgeCount(chart: ChartData, axis: 'x' | 'y' | 'z'): number {
    if (chart.valuePoints3D?.length) {
      const idx = axis === 'x' ? 0 : axis === 'y' ? 1 : 2
      return new Set(chart.valuePoints3D.map((p) => p[idx])).size
    }
    if (chart.valueTuples?.length) {
      if (axis === 'z') return 0
      const idx = axis === 'x' ? 0 : 1
      return new Set(chart.valueTuples.map((p) => p[idx])).size
    }
    return 0
  }

  grandTotal(chart: ChartData): number {
    if (chart.valuePoints3D?.length) {
      return chart.valuePoints3D.reduce((sum, [, , z, metric]) => {
        const value = metric ?? z
        return sum + (Number.isFinite(value) ? value : 0)
      }, 0)
    }
    if (chart.valueTuples?.length) {
      return chart.valueTuples.reduce((sum, [, y]) => sum + (Number.isFinite(y) ? y : 0), 0)
    }
    return 0
  }

  is3D(chart: ChartData): boolean {
    return chart.render3D?.mode === 'continuous'
  }

  canOfferValue3D(): boolean {
    return false
  }
}
