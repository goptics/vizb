import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { getNextColorFor } from '../../lib/utils'
import { getChartStyling } from './shared'

type Series3DType = 'bar3D' | 'line3D'

/**
 * Builds 3D (bar3D / line3D) chart options via echarts-gl.
 * Dimension mapping: xAxis -> X floor, yAxis -> Y depth, stat value -> Z height,
 * zAxis -> color (one 3D series per z value, listed in a left-side legend).
 * No in-canvas title/axis names; the chart title lives outside the canvas.
 */
export function use3DChartOptions(config: BaseChartConfig, seriesType: Series3DType) {
  const { chartData, sort, showLabels, isDark } = config

  const options = computed<EChartsOption>(() => {
    const styling = getChartStyling(isDark.value)
    const base = getBaseOptions(config)
    const cd = chartData.value
    const points = cd.points ?? []

    let xValues = Array.from(new Set(points.map((p) => p.xAxis)))
    let yValues = Array.from(new Set(points.map((p) => p.yAxis)))
    let zValues = cd.zAxis.filter((z) => z !== '')

    // Sort each axis (X floor, Y depth, Z color/stack) by its total value,
    // mirroring the 2D charts.
    if (sort.value.enabled) {
      const sortByTotal = (values: string[], key: 'xAxis' | 'yAxis' | 'zAxis') => {
        const totals = new Map<string, number>()
        for (const p of points) totals.set(p[key], (totals.get(p[key]) ?? 0) + p.value)
        return values.sort((a, b) => {
          const diff = (totals.get(a) ?? 0) - (totals.get(b) ?? 0)
          return sort.value.order === 'asc' ? diff : -diff
        })
      }
      xValues = sortByTotal(xValues, 'xAxis')
      yValues = sortByTotal(yValues, 'yAxis')
      zValues = sortByTotal(zValues, 'zAxis')
    }

    const round2 = (v: number) => Math.round(v * 100) / 100

    const xIndex = new Map(xValues.map((v, i) => [v, i]))
    const yIndex = new Map(yValues.map((v, i) => [v, i]))

    // Total value per (x, y) cell, summed across z — shown as the stack-top label.
    const cellTotals = new Map<string, number>()
    for (const p of points) {
      const key = `${xIndex.get(p.xAxis)},${yIndex.get(p.yAxis)}`
      cellTotals.set(key, (cellTotals.get(key) ?? 0) + p.value)
    }

    // One series per z value, stacked at each (x, y) cell so z values don't
    // overlap. bar3D reads [xIdx, yIdx, height]; `stack` accumulates heights.
    // Only the topmost series carries a label, showing the cell total.
    const series = zValues.map((z, zi) => ({
      name: z,
      type: seriesType,
      ...(seriesType === 'bar3D' ? { stack: 'z', bevelSize: 0.4, bevelSmoothness: 4 } : {}),
      data: points
        .filter((p) => p.zAxis === z)
        .map((p) => ({ value: [xIndex.get(p.xAxis), yIndex.get(p.yAxis), p.value] })),
      itemStyle: { color: getNextColorFor(z) },
      shading: 'lambert',
      label: {
        show: showLabels.value && zi === zValues.length - 1,
        fontSize: 12,
        textStyle: { color: styling.textColor },
        formatter: (p: { value: number[] }) => {
          const [xi = 0, yi = 0] = p.value
          const total = cellTotals.get(`${xi},${yi}`)
          return total === undefined ? '' : String(round2(total))
        },
      },
      emphasis: { disabled: true },
    }))

    const axisCommon = {
      axisLabel: { color: styling.textColor },
      axisLine: { lineStyle: { color: styling.axisColor } },
    }

    const opt = {
      ...base,
      emphasis: { focus: 'none' },
      legend: {
        ...base.legend,
        show: zValues.length > 1,
        orient: 'vertical',
        left: 'left',
        top: 'middle',
      },
      tooltip: {
        ...base.tooltip,
        formatter: (params: { value: number[] }) => {
          const [xi = 0, yi = 0] = params.value
          const xName = xValues[xi]
          const yName = yValues[yi]
          const cell = points.filter((p) => p.xAxis === xName && p.yAxis === yName)
          const rows = zValues
            .map((z) => {
              const pt = cell.find((p) => p.zAxis === z)
              if (!pt) return ''
              const dot = `<span style="display:inline-block;width:10px;height:10px;border-radius:50%;background:${getNextColorFor(z)};margin-right:6px"></span>`
              return `${dot}${z}: <b>${round2(pt.value)}</b>`
            })
            .filter(Boolean)
          return `<b>${xName} / ${yName}</b><br/>${rows.join('<br/>')}`
        },
      },
      xAxis3D: { type: 'category', data: xValues, ...axisCommon },
      yAxis3D: { type: 'category', data: yValues, ...axisCommon },
      zAxis3D: { type: 'value', ...axisCommon },
      grid3D: {
        boxWidth: 100,
        boxDepth: 100,
        axisLine: { lineStyle: { color: styling.axisColor } },
        splitLine: { lineStyle: { color: styling.axisColor, opacity: styling.opacity } },
        viewControl: { autoRotate: false, distance: 200 },
        light: {
          main: { intensity: 1.2, shadow: false },
          ambient: { intensity: 0.4 },
        },
      },
      series,
    }

    return opt as unknown as EChartsOption
  })

  return { options }
}
