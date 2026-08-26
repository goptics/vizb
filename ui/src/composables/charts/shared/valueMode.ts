import type { EChartsOption } from 'echarts'
import type { ScaleType, ChartType } from '@/types'
import { getNextColorFor, VALUE_CHART_TYPES, formatChartNumber } from '@/lib/utils'
import { axisIsLog, axisLogBase, parseScale } from '@/lib/scale'
import { type BaseChartConfig, getBaseOptions } from '../baseChartOptions'
import {
  createValueModeGridConfig,
  createLabelConfig,
  createValueAxisConfig,
  createValueModeTooltip,
  getChartStyling,
  isLargeXAxis,
  INSIDE_XY_ZOOM,
  LARGE_DATA_THRESHOLD,
  scatterSeriesLargeOpts,
} from './chartConfig'
import { adjustForLogScaleLine, resolveLogScale } from './common'
import { resolveSeriesSymbol } from './seriesConfig'
import { resolve2DScatterVisualMap } from './visualMap'

const defaultScatterSymbol = { symbol: 'circle' as const, symbolSize: 8 }
const largeScatterSymbol = { symbol: 'circle' as const, symbolSize: 5 }
const defaultLineSymbol = { symbol: 'circle' as const, symbolSize: 7 }
const largeLineSymbol = { symbol: 'none', sampling: 'lttb' as const }

export function sortValueTuples(
  tuples: [number, number, number?][],
  enabled: boolean,
  order: 'asc' | 'desc'
): [number, number, number?][] {
  if (!enabled) return tuples
  const sorted = [...tuples].sort((a, b) => a[1] - b[1])
  return order === 'asc' ? sorted : sorted.reverse()
}

export function scaleValueTuples(
  tuples: [number, number, number?][],
  yScale: ScaleType,
  xScale: ScaleType = 'linear'
): [number, number | null, number?][] {
  const xLog = xScale === 'log'
  const yLog = yScale === 'log'
  if (!xLog && !yLog) return tuples
  const out: [number, number | null, number?][] = []
  for (const [x, y, c] of tuples) {
    if (xLog && x <= 0) continue
    const yAdj = adjustForLogScaleLine(y, yScale)
    out.push(c !== undefined ? [x, yAdj, c] : [x, yAdj])
  }
  return out
}

const chartTypeForECharts = (chartType: ChartType): string =>
  VALUE_CHART_TYPES.has(chartType) ? chartType : 'scatter'

const seriesSymbol = (
  chartType: ChartType,
  largeX: boolean,
  symbol?: string,
  symbolSize?: number
) => {
  if (chartType === 'scatter') {
    return resolveSeriesSymbol(
      largeX ? largeScatterSymbol : defaultScatterSymbol,
      symbol,
      symbolSize
    )
  }
  if (chartType === 'line') {
    return resolveSeriesSymbol(largeX ? largeLineSymbol : defaultLineSymbol, symbol, symbolSize)
  }
  return {}
}

export function buildValueAxes2DOptions(
  config: BaseChartConfig,
  chartType: ChartType = 'scatter'
): EChartsOption {
  const { chartData, sort, showLabels, isDark, scale } = config
  const tuples = chartData.value.valueTuples ?? []
  const xLabel = chartData.value.axisLabels?.x
  const yLabel = chartData.value.axisLabels?.y
  const baseOptions = getBaseOptions(config)
  const styling = getChartStyling(isDark.value)
  const parsed = parseScale(scale?.value)
  const xWant = axisIsLog(parsed, 'x', ['y'])
  const yWant = axisIsLog(parsed, 'y', ['y'])
  const xScale = resolveLogScale(
    xWant ? 'log' : 'linear',
    tuples.map((t) => t[0])
  )
  const yScale = resolveLogScale(
    yWant ? 'log' : 'linear',
    tuples.map((t) => t[1])
  )

  const sorted = sortValueTuples(tuples, sort.value.enabled, sort.value.order)
  const data = scaleValueTuples(sorted, yScale, xScale)
  const largeX = isLargeXAxis(data.map((_, i) => String(i)))

  const useVisualMap = chartType === 'scatter' && config.visualMap?.value === true
  const smoothLines = chartType === 'line' && config.smooth?.value === true
  const hasColorDim = tuples.some((t) => t[2] !== undefined)
  const colorDimension = (hasColorDim ? 2 : 1) as 1 | 2
  const colorValues = sorted
    .map((t) => (hasColorDim ? t[2] : t[1]))
    .filter((v): v is number => v !== undefined && v !== null && isFinite(v))

  const label = {
    ...createLabelConfig(showLabels.value, styling),
    formatter: (p: { data: [number, number | null, number?] }) => {
      const y = p.data[1]
      return y === null || y === undefined ? '' : formatChartNumber(y)
    },
  }

  const series = {
    name: chartData.value.title,
    type: chartTypeForECharts(chartType) as 'scatter' | 'bar' | 'line',
    data,
    label,
    ...(chartType === 'scatter'
      ? scatterSeriesLargeOpts(useVisualMap)
      : { large: true as const, largeThreshold: LARGE_DATA_THRESHOLD }),
    ...(chartType === 'line' ? { smooth: smoothLines } : {}),
    ...(useVisualMap ? {} : { itemStyle: { color: getNextColorFor(chartData.value.title) } }),
    ...seriesSymbol(chartType, largeX, config.symbol?.value, config.symbolSize?.value),
  }

  return {
    ...baseOptions,
    legend: { show: false },
    grid: createValueModeGridConfig(false),
    visualMap: resolve2DScatterVisualMap(useVisualMap, colorValues, styling, colorDimension),
    tooltip: createValueModeTooltip(
      isDark.value,
      xLabel,
      yLabel,
      chartType === 'scatter' || chartType === 'line'
    ),
    ...createValueAxisConfig(
      styling,
      xLabel,
      yLabel,
      yScale,
      chartType === 'line' || chartType === 'scatter',
      {
        xScale,
        xLogBase: axisLogBase(parsed, 'x'),
        yLogBase: axisLogBase(parsed, 'y'),
      }
    ),
    ...(chartType === 'line' ? {} : { dataZoom: INSIDE_XY_ZOOM }),
    series: [series],
  } as EChartsOption
}
