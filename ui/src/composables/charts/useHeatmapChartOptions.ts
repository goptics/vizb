import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { getNextColorFor, isGrouped3D, COLOR_PALETTE } from '@/lib/utils'
import {
  getChartStyling,
  getTooltipTheme,
  tooltipDivider,
  tooltipSpreadRows,
  renderDonutSvg,
  isLargeXAxis,
  createHeatmapDataZoomConfig,
} from './shared'

const round2 = (v: number) => Math.round(v * 100) / 100

function formatCellNumber(v: number): string {
  if (Math.abs(v) >= 1e6) return (v / 1e6).toFixed(1) + 'M'
  if (Math.abs(v) >= 1e3) return (v / 1e3).toFixed(1) + 'K'
  return String(round2(v))
}

function heatmapGrid(
  seriesLength: number,
  largeX: boolean,
  largeY: boolean,
  hasLegend = true
): any {
  const legendSpace = hasLegend ? Math.min(15 + Math.floor((seriesLength - 1) / 15) * 2, 35) : 5
  return {
    left: largeX ? 100 : '3%',
    right: largeY ? 50 : '3%',
    bottom: largeX ? 110 : '13%',
    top: `${legendSpace}%`,
    containLabel: !largeX,
  }
}

function heatmapVisualMap(
  min: number,
  max: number,
  colors: string[],
  styling: any,
  largeX: boolean
): any {
  return {
    min,
    max,
    calculable: true,
    orient: 'horizontal',
    left: 'center',
    bottom: largeX ? 5 : '5%',
    inRange: { color: colors },
    textStyle: { color: styling.textColor },
  }
}

function build2DHeatmap(config: BaseChartConfig): EChartsOption {
  const { chartData, isDark, showLabels } = config
  const base = getBaseOptions(config)
  const styling = getChartStyling(isDark.value)
  const data = chartData.value

  const xCategories = data.series.map((s) => s.xAxis)
  const yCategories = data.yAxis
  const largeX = isLargeXAxis(xCategories)
  const largeY = isLargeXAxis(yCategories)
  const hasZoom = largeX || largeY

  const heatmapData: number[][] = []
  let minVal = Infinity
  let maxVal = -Infinity

  const xTotals = new Map<number, number>()
  const yTotals = new Map<number, number>()

  for (let xi = 0; xi < data.series.length; xi++) {
    const series = data.series[xi]
    if (!series) continue
    for (let yi = 0; yi < series.values.length; yi++) {
      const v = series.values[yi] ?? 0
      heatmapData.push([xi, yi, v])
      if (v < minVal) minVal = v
      if (v > maxVal) maxVal = v
      xTotals.set(xi, (xTotals.get(xi) ?? 0) + v)
      yTotals.set(yi, (yTotals.get(yi) ?? 0) + v)
    }
  }

  if (minVal === maxVal) {
    minVal = minVal - 1
    maxVal = maxVal + 1
  }

  return {
    ...base,
    grid: heatmapGrid(1, largeX, largeY, false),
    legend: { show: false },
    ...(hasZoom
      ? {
          dataZoom: createHeatmapDataZoomConfig(
            largeX,
            largeY,
            xCategories.length,
            yCategories.length,
            styling
          ),
        }
      : {}),
    tooltip: {
      ...getTooltipTheme(isDark.value),
      position: 'top',
      formatter: (params: any) => {
        const [xi, yi, val] = params.data
        const xName = xCategories[xi] ?? ''
        const yName = yCategories[yi] ?? ''
        const xTotal = xTotals.get(xi) ?? 0
        const yTotal = yTotals.get(yi) ?? 0
        const xPct = xTotal > 0 ? ((val / xTotal) * 100).toFixed(1) : '0.0'
        const yPct = yTotal > 0 ? ((val / yTotal) * 100).toFixed(1) : '0.0'

        let html = `<b>${xName} / ${yName}</b><br/>`
        html += `Value: <b>${round2(val)}</b><br/>`
        html += tooltipDivider(isDark.value)
        html += `Σ ${xName}: <b>${round2(xTotal)}</b> (${xPct}%)<br/>`
        html += `Σ ${yName}: <b>${round2(yTotal)}</b> (${yPct}%)`

        return html
      },
    },
    xAxis: {
      type: 'category',
      data: xCategories,
      axisLabel: {
        color: styling.textColor,
        fontSize: 12,
        interval: largeX ? 'auto' : 0,
        rotate: largeX ? 0 : xCategories.reduce((a, c) => a + c.length, 0) > 100 ? 30 : 0,
      },
      axisLine: { lineStyle: { color: styling.axisColor } },
      ...(data.axisLabels?.x
        ? {
            name: data.axisLabels.x,
            nameLocation: 'middle',
            nameGap: largeX ? 41 : 30,
            nameTextStyle: { color: styling.textColor, fontSize: 14, fontWeight: 'bold' },
          }
        : {}),
    },
    yAxis: {
      type: 'category',
      data: yCategories,
      axisLabel: { color: styling.textColor, fontSize: 12, interval: largeY ? 'auto' : 0 },
      axisLine: { lineStyle: { color: styling.axisColor } },
      ...(data.axisLabels?.y
        ? {
            name: data.axisLabels.y,
            nameLocation: 'middle',
            nameGap: 60,
            nameTextStyle: { color: styling.textColor, fontSize: 14, fontWeight: 'bold' },
          }
        : {}),
    },
    visualMap: heatmapVisualMap(
      minVal,
      maxVal,
      [COLOR_PALETTE[0]!, COLOR_PALETTE[4]!],
      styling,
      largeX
    ),
    series: [
      {
        type: 'heatmap',
        data: heatmapData,
        label: {
          show: showLabels.value,
          formatter: (params: any) => formatCellNumber(params.data[2]),
          color: styling.textColor,
          fontSize: 11,
        },
        emphasis: {
          itemStyle: { shadowBlur: 10, shadowColor: 'rgba(0,0,0,0.5)' },
        },
      },
    ],
  } as EChartsOption
}

function build3DHeatmap(config: BaseChartConfig): EChartsOption {
  const { chartData, isDark, showLabels, visibleZ } = config
  const base = getBaseOptions(config)
  const styling = getChartStyling(isDark.value)
  const data = chartData.value
  const points = data.points ?? []

  const render = data.render3D ?? {
    xValues: [],
    yValues: [],
    zValues: [],
    barSeries: [],
    lineSeries: [],
    cellTotals: {},
  }
  const { xValues, yValues, zValues } = render
  const largeX = isLargeXAxis(xValues)
  const largeY = isLargeXAxis(yValues)
  const hasZoom = largeX || largeY

  const sel = visibleZ?.value ?? {}
  const visiblePoints = points.filter((p) => sel[p.zAxis] !== false)

  const cellLookup = new Map<string, Map<string, number>>()
  const xMarginals = new Map<string, number>()
  const yMarginals = new Map<string, number>()
  for (const p of visiblePoints) {
    const key = `${p.xAxis},${p.yAxis}`
    let zmap = cellLookup.get(key)
    if (!zmap) {
      zmap = new Map()
      cellLookup.set(key, zmap)
    }
    zmap.set(p.zAxis, (zmap.get(p.zAxis) ?? 0) + p.value)
    xMarginals.set(p.xAxis, (xMarginals.get(p.xAxis) ?? 0) + p.value)
    yMarginals.set(p.yAxis, (yMarginals.get(p.yAxis) ?? 0) + p.value)
  }

  const heatmapData: number[][] = []
  const cellDataMap = new Map<string, { zBreakdown: Map<string, number>; total: number }>()
  let minTotal = Infinity
  let maxTotal = -Infinity

  for (let xi = 0; xi < xValues.length; xi++) {
    for (let yi = 0; yi < yValues.length; yi++) {
      const xName = xValues[xi]
      const yName = yValues[yi]
      const zmap = cellLookup.get(`${xName},${yName}`) ?? new Map()

      let cellTotal = 0
      for (const v of zmap.values()) cellTotal += v

      cellDataMap.set(`${xi},${yi}`, { zBreakdown: zmap, total: cellTotal })

      if (cellTotal > 0) {
        if (cellTotal < minTotal) minTotal = cellTotal
        if (cellTotal > maxTotal) maxTotal = cellTotal
      }

      heatmapData.push([xi, yi, cellTotal])
    }
  }

  if (minTotal === maxTotal) {
    minTotal = minTotal - 1
    maxTotal = maxTotal + 1
  }

  const xLabel = data.axisLabels?.x ?? 'x'
  const yLabel = data.axisLabels?.y ?? 'y'
  const zLabel = data.axisLabels?.z ?? 'z'

  const tooltipFormatter = (params: any) => {
    const val = Array.isArray(params.value) ? params.value : params.data
    const [xi, yi] = val
    const xName = xValues[xi] ?? ''
    const yName = yValues[yi] ?? ''
    const cell = cellDataMap.get(`${xi},${yi}`)
    if (!cell) return ''

    const { zBreakdown, total } = cell

    const rows = zValues
      .map((z) => {
        const v = zBreakdown.get(z)
        if (v === undefined) return ''
        const color = getNextColorFor(z) ?? COLOR_PALETTE[0]!
        const dot = `<span style="display:inline-block;width:10px;height:10px;border-radius:50%;background:${color};margin-right:6px"></span>`
        return `${dot}${z}: <b>${round2(v)}</b>`
      })
      .filter(Boolean)

    const zSumLine = zBreakdown.size > 1 ? `Σ ${zLabel}: <b>${round2(total)}</b><br/>` : ''

    const xMarginal = xMarginals.get(xName) ?? 0
    const yMarginal = yMarginals.get(yName) ?? 0

    const margins =
      tooltipDivider(isDark.value) +
      zSumLine +
      `Σ ${xLabel}(${xName}): <b>${round2(xMarginal)}</b><br/>` +
      `Σ ${yLabel}(${yName}): <b>${round2(yMarginal)}</b>`

    const spread = tooltipSpreadRows(Array.from(zBreakdown.values()), isDark.value)

    const donut =
      zBreakdown.size > 1
        ? renderDonutSvg(
            zValues
              .filter((z) => zBreakdown.has(z))
              .map((z) => ({
                value: zBreakdown.get(z)!,
                color: getNextColorFor(z) ?? '',
                name: z,
              }))
          )
        : ''

    return `<b>${xLabel}: ${xName} / ${yLabel}: ${yName}</b><br/>${rows.join('<br/>')}${margins}${spread}${donut ? tooltipDivider(isDark.value) + donut : ''}`
  }

  return {
    ...base,
    grid: heatmapGrid(zValues.length, largeX, largeY, zValues.length > 1),
    ...(hasZoom
      ? {
          dataZoom: createHeatmapDataZoomConfig(
            largeX,
            largeY,
            xValues.length,
            yValues.length,
            styling
          ),
        }
      : {}),
    legend: {
      show: zValues.length > 1,
      left: 'center',
      itemWidth: 10,
      itemHeight: 10,
      textStyle: { color: styling.textColor },
      selected: sel,
      data: zValues.map((z) => ({
        name: z,
        itemStyle: { color: getNextColorFor(z) ?? COLOR_PALETTE[0]! },
      })),
    },
    tooltip: {
      ...base.tooltip,
      ...getTooltipTheme(isDark.value),
      trigger: 'item',
      formatter: tooltipFormatter,
    },
    xAxis: {
      type: 'category',
      data: xValues,
      axisLabel: {
        color: styling.textColor,
        fontSize: 12,
        interval: largeX ? 'auto' : 0,
        rotate: largeX ? 0 : xValues.reduce((a, c) => a + c.length, 0) > 100 ? 30 : 0,
      },
      axisLine: { lineStyle: { color: styling.axisColor } },
      ...(data.axisLabels?.x
        ? {
            name: data.axisLabels.x,
            nameLocation: 'middle',
            nameGap: largeX ? 41 : 30,
            nameTextStyle: { color: styling.textColor, fontSize: 14, fontWeight: 'bold' },
          }
        : {}),
    },
    yAxis: {
      type: 'category',
      data: yValues,
      axisLabel: { color: styling.textColor, fontSize: 12, interval: largeY ? 'auto' : 0 },
      axisLine: { lineStyle: { color: styling.axisColor } },
      ...(data.axisLabels?.y
        ? {
            name: data.axisLabels.y,
            nameLocation: 'middle',
            nameGap: 60,
            nameTextStyle: { color: styling.textColor, fontSize: 14, fontWeight: 'bold' },
          }
        : {}),
    },
    visualMap: heatmapVisualMap(
      minTotal,
      maxTotal,
      [COLOR_PALETTE[0]!, COLOR_PALETTE[4]!],
      styling,
      largeX
    ),
    series: [
      {
        type: 'heatmap',
        data: heatmapData,
        label: {
          show: showLabels.value,
          formatter: (params: any) => {
            const v = Array.isArray(params.value) ? params.value[2] : params.data?.[2]
            return v === undefined || v === 0 ? '' : formatCellNumber(v)
          },
          color: '#fff',
          fontSize: 11,
          textBorderColor: 'rgba(0,0,0,0.5)',
          textBorderWidth: 2,
        },
        emphasis: {
          itemStyle: { shadowBlur: 10, shadowColor: 'rgba(0,0,0,0.5)' },
        },
      },
      ...zValues.map((z) => ({
        name: z,
        type: 'scatter' as const,
        data: [] as number[][],
        itemStyle: { color: getNextColorFor(z) ?? COLOR_PALETTE[0]! },
        symbolSize: 0,
        silent: true,
      })),
    ],
  } as EChartsOption
}

export function useHeatmapChartOptions(config: BaseChartConfig) {
  const { chartData } = config

  const options = computed<EChartsOption>(() => {
    const grouped3D = isGrouped3D(chartData.value)
    return grouped3D ? build3DHeatmap(config) : build2DHeatmap(config)
  })

  return { options }
}
