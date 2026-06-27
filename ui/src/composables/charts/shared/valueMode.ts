import type { EChartsOption } from 'echarts'
import type { ScaleType, ChartType } from '@/types'
import { getNextColorFor, VALUE_CHART_TYPES } from '@/lib/utils'
import { type BaseChartConfig, getBaseOptions } from '../baseChartOptions'
import {
  createValueModeGridConfig,
  createLabelConfig,
  createValueAxisConfig,
  createValueModeTooltip,
  getChartStyling,
  isLargeXAxis,
  LARGE_DATA_THRESHOLD,
} from './chartConfig'
import { adjustForLogScaleLine, getEffectiveScale } from './common'
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
  scale: ScaleType
): [number, number | null, number?][] {
  if (scale !== 'log') return tuples
  return tuples.map(([x, y, c]) => [x, adjustForLogScaleLine(y, scale), c])
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
  const effectiveScale = scale?.value ?? 'linear'
  const { minValue, effectiveScale: yScale } = getEffectiveScale(
    [{ xAxis: chartData.value.title, values: tuples.map((t) => t[1]), benchmarkId: '' }],
    effectiveScale
  )

  const sorted = sortValueTuples(tuples, sort.value.enabled, sort.value.order)
  const data = scaleValueTuples(sorted, yScale)
  const largeX = isLargeXAxis(data.map((_, i) => String(i)))

  const useVisualMap = chartType === 'scatter' && config.visualMap?.value === true
  const hasColorDim = tuples.some((t) => t[2] !== undefined)
  const colorDimension = (hasColorDim ? 2 : 1) as 1 | 2
  const colorValues = sorted
    .map((t) => (hasColorDim ? t[2] : t[1]))
    .filter((v): v is number => v !== undefined && v !== null && isFinite(v))

  const label = {
    ...createLabelConfig(showLabels.value, styling),
    formatter: (p: { data: [number, number | null, number?] }) => {
      const y = p.data[1]
      return y === null || y === undefined ? '' : String(Math.round(y * 100) / 100)
    },
  }

  const series = {
    name: chartData.value.title,
    type: chartTypeForECharts(chartType) as 'scatter' | 'bar' | 'line',
    data,
    label,
    large: true,
    largeThreshold: LARGE_DATA_THRESHOLD,
    ...(useVisualMap ? {} : { itemStyle: { color: getNextColorFor(chartData.value.title) } }),
    ...seriesSymbol(chartType, largeX, config.symbol?.value, config.symbolSize?.value),
  }

  return {
    ...baseOptions,
    legend: { show: false },
    grid: createValueModeGridConfig(false),
    visualMap: resolve2DScatterVisualMap(useVisualMap, colorValues, styling, colorDimension),
    tooltip: createValueModeTooltip(isDark.value, xLabel, yLabel),
    ...createValueAxisConfig(
      styling,
      xLabel,
      yLabel,
      yScale,
      minValue,
      chartType === 'line' || chartType === 'scatter'
    ),
    dataZoom: [
      { type: 'inside', xAxisIndex: 0 },
      { type: 'inside', yAxisIndex: 0 },
    ],
    series: [series],
  } as EChartsOption
}
