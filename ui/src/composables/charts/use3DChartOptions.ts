import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { getNextColorFor } from '../../lib/utils'
import { getChartStyling, type ChartStyling } from './shared'
import type { Point3D } from '../../types'

type Series3DType = 'bar3D' | 'line3D'

const round2 = (v: number) => Math.round(v * 100) / 100

function mapPoints(
  points: Point3D[],
  z: string,
  xIndex: Map<string, number>,
  yIndex: Map<string, number>,
) {
  return points
    .filter((p) => p.zAxis === z)
    .map((p) => ({ value: [xIndex.get(p.xAxis), yIndex.get(p.yAxis), p.value] }))
}

function makeLabel(
  show: boolean,
  textColor: string,
  formatter: (p: { value: number[] }) => string,
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
  const { chartData, sort, showLabels, isDark, autoRotate } = config

  const options = computed<EChartsOption>(() => {
    const styling = getChartStyling(isDark.value)
    const base = getBaseOptions(config)
    const points: Point3D[] = chartData.value.points ?? []

    let xValues = Array.from(new Set(points.map((p) => p.xAxis)))
    let yValues = Array.from(new Set(points.map((p) => p.yAxis)))
    let zValues = chartData.value.zAxis.filter((z) => z !== '')

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

    const xIndex = new Map(xValues.map((v, i) => [v, i]))
    const yIndex = new Map(yValues.map((v, i) => [v, i]))

    const cellTotals = new Map<string, number>()
    for (const p of points) {
      const key = `${xIndex.get(p.xAxis)},${yIndex.get(p.yAxis)}`
      cellTotals.set(key, (cellTotals.get(key) ?? 0) + p.value)
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
        },
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
              v === undefined ? '' : String(round2(v)),
            ),
            emphasis: { disabled: true },
          }))
        : []

    const axisCommon = makeAxisCommon(styling)

    const tooltipFormatter = (params: { value: number[] }) => {
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
    }

    const grid3DConfig = {
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
        textStyle: { color: styling.textColor },
      },
      tooltip: {
        ...base.tooltip,
        backgroundColor: isDark.value ? '#1f2937' : '#ffffff',
        borderColor: isDark.value ? '#4b5563' : '#e5e7eb',
        textStyle: { color: styling.textColor },
        formatter: tooltipFormatter,
      },
      xAxis3D: { type: 'category', data: xValues, ...axisCommon },
      yAxis3D: { type: 'category', data: yValues, ...axisCommon },
      zAxis3D: { type: 'value', ...axisCommon },
      grid3D: grid3DConfig,
      series: [...series, ...labelSeries],
    }

    return opt as unknown as EChartsOption
  })

  return { options }
}
