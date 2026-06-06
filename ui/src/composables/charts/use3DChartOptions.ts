import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { getNextColorFor } from '../../lib/utils'
import { getChartStyling, getTooltipTheme, tooltipDivider, type ChartStyling } from './shared'
import { sortByAxisTotal } from './shared/common'
import type { Point3D } from '../../types'

type Series3DType = 'bar3D' | 'line3D'

const round2 = (v: number) => Math.round(v * 100) / 100

function mapPoints(
  points: Point3D[],
  z: string,
  xIndex: Map<string, number>,
  yIndex: Map<string, number>
) {
  return points
    .filter((p) => p.zAxis === z)
    .map((p) => ({ value: [xIndex.get(p.xAxis), yIndex.get(p.yAxis), p.value] }))
}

function makeLabel(
  show: boolean,
  textColor: string,
  formatter: (p: { value: number[] }) => string
) {
  return { show, fontSize: 12, textStyle: { color: textColor }, formatter }
}

function makeAxisCommon(styling: ChartStyling) {
  return {
    axisLabel: { color: styling.textColor },
    axisLine: { lineStyle: { color: styling.axisColor } },
  }
}

export function use3DChartOptions(config: BaseChartConfig, seriesType: Series3DType) {
  const { chartData, sort, showLabels, isDark, autoRotate, visibleZ } = config

  const options = computed<EChartsOption>(() => {
    const styling = getChartStyling(isDark.value)
    const base = getBaseOptions(config)
    const points: Point3D[] = chartData.value.points ?? []

    // Legend can toggle z series off. Aggregates (tooltip sums + bar-top labels)
    // must reflect only the currently-visible z, so sum over the filtered set.
    // echarts treats a missing legend key as selected → default everything on.
    const sel = visibleZ?.value ?? {}
    const aggPoints = points.filter((p) => sel[p.zAxis] !== false)

    let xValues = Array.from(new Set(points.map((p) => p.xAxis)))
    let yValues = Array.from(new Set(points.map((p) => p.yAxis)))
    let zValues = chartData.value.zAxis.filter((z) => z !== '')

    // Marginal totals per category (sum over the other two visible axes).
    const sumByKey = (key: 'xAxis' | 'yAxis') => {
      const t = new Map<string, number>()
      for (const p of aggPoints) t.set(p[key], (t.get(p[key]) ?? 0) + p.value)
      return t
    }
    const xTotals = sumByKey('xAxis')
    const yTotals = sumByKey('yAxis')

    if (sort.value.enabled) {
      xValues = sortByAxisTotal(xValues, 'xAxis', points, sort.value.order)
      yValues = sortByAxisTotal(yValues, 'yAxis', points, sort.value.order)
      zValues = sortByAxisTotal(zValues, 'zAxis', points, sort.value.order)
    }

    const xIndex = new Map(xValues.map((v, i) => [v, i]))
    const yIndex = new Map(yValues.map((v, i) => [v, i]))

    // Precompute per-cell aggregates so tooltip hover is O(1), not O(points):
    // cellTotals = stacked height (Σ over z); cellZ = z→value for the breakdown.
    const cellTotals = new Map<string, number>()
    const cellZ = new Map<string, Map<string, number>>()
    for (const p of aggPoints) {
      const key = `${xIndex.get(p.xAxis)},${yIndex.get(p.yAxis)}`
      cellTotals.set(key, (cellTotals.get(key) ?? 0) + p.value)
      let zmap = cellZ.get(key)
      if (!zmap) {
        zmap = new Map()
        cellZ.set(key, zmap)
      }
      zmap.set(p.zAxis, (zmap.get(p.zAxis) ?? 0) + p.value)
    }

    const series = zValues.map((z, zi) => ({
      name: z,
      type: seriesType,
      ...(seriesType === 'bar3D'
        ? { stack: 'z', bevelSize: 0.4, bevelSmoothness: 4 }
        : { lineStyle: { width: 2 } }),
      data: mapPoints(points, z, xIndex, yIndex),
      itemStyle: { color: getNextColorFor(z) },
      shading: 'lambert',
      label: makeLabel(
        showLabels.value && seriesType === 'bar3D' && zi === zValues.length - 1,
        styling.textColor,
        ({ value: [xi = 0, yi = 0] }) => {
          const total = cellTotals.get(`${xi},${yi}`)
          return total === undefined ? '' : String(round2(total))
        }
      ),
      emphasis: { disabled: true },
    }))

    // line3D can't render vertex markers or labels itself — scatter3D overlay handles both.
    const labelSeries =
      seriesType === 'line3D'
        ? zValues.map((z) => ({
            name: z,
            type: 'scatter3D',
            data: mapPoints(points, z, xIndex, yIndex),
            symbolSize: 10,
            itemStyle: { color: getNextColorFor(z) },
            label: makeLabel(showLabels.value, styling.textColor, ({ value: [, , v] }) =>
              v === undefined ? '' : String(round2(v))
            ),
            emphasis: { disabled: true },
          }))
        : []

    const axisCommon = makeAxisCommon(styling)

    const tooltipFormatter = (params: { value: number[] }) => {
      const [xi = 0, yi = 0] = params.value
      const xName = xValues[xi]
      const yName = yValues[yi]
      const key = `${xi},${yi}`
      const zmap = cellZ.get(key)
      const rows = zValues
        .map((z) => {
          const v = zmap?.get(z)
          if (v === undefined) return ''
          const dot = `<span style="display:inline-block;width:10px;height:10px;border-radius:50%;background:${getNextColorFor(z)};margin-right:6px"></span>`
          return `${dot}${z}: <b>${round2(v)}</b>`
        })
        .filter(Boolean)

      // Σ over z = stacked bar height at this (x,y). First line under the
      // divider, above the x/y marginals, when there's more than one z to sum.
      const zSumLine =
        zmap && zmap.size > 1 ? `Σ z: <b>${round2(cellTotals.get(key) ?? 0)}</b><br/>` : ''

      // Marginal totals: sum over the other two axes for this x / this y.
      const margins =
        tooltipDivider(isDark.value) +
        zSumLine +
        `Σ ${xName}: <b>${round2(xTotals.get(xName ?? '') ?? 0)}</b><br/>` +
        `Σ ${yName}: <b>${round2(yTotals.get(yName ?? '') ?? 0)}</b>`

      return `<b>${xName} / ${yName}</b><br/>${rows.join('<br/>')}${margins}`
    }

    return {
      ...base,
      legend: {
        ...base.legend,
        show: zValues.length > 1,
        orient: 'vertical',
        left: 'left',
        top: 'middle',
        textStyle: { color: styling.textColor },
        // Controlled selection: persist toggles across recomputes (without this,
        // re-applying the option would reset every z back to visible).
        selected: sel,
      },
      tooltip: {
        ...base.tooltip,
        ...getTooltipTheme(isDark.value),
        formatter: tooltipFormatter,
      },
      xAxis3D: { type: 'category', data: xValues, ...axisCommon },
      yAxis3D: { type: 'category', data: yValues, ...axisCommon },
      zAxis3D: { type: 'value', ...axisCommon },
      grid3D: {
        boxWidth: 100,
        boxDepth: 100,
        axisLine: { lineStyle: { color: styling.axisColor } },
        splitLine: { lineStyle: { color: styling.axisColor, opacity: styling.opacity } },
        viewControl: {
          distance: 200,
          autoRotate: autoRotate.value,
          ...(seriesType === 'line3D' ? { projection: 'orthographic' } : {}),
        },
        light: {
          main: { intensity: 0.3, shadow: false },
          ambient: { intensity: 0.9 },
        },
      },
      series: [...series, ...labelSeries],
    } as unknown as EChartsOption
  })

  return { options }
}
