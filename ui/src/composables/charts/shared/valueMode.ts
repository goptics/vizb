import type { EChartsOption } from 'echarts'
import type { ScaleType, ChartType } from '@/types'
import { getNextColorFor, VALUE_CHART_TYPES } from '@/lib/utils'
import { type BaseChartConfig, getBaseOptions } from '../baseChartOptions'
import {
  createGridConfig,
  createLabelConfig,
  createValueAxisConfig,
  createValueModeTooltip,
  getChartStyling,
  isLargeXAxis,
  LARGE_DATA_THRESHOLD,
} from './chartConfig'
import { adjustForLogScaleLine, getEffectiveScale } from './common'
import { resolveSeriesSymbol } from './seriesConfig'

const defaultScatterSymbol = { symbol: 'circle' as const, symbolSize: 8 }
const largeScatterSymbol = { symbol: 'circle' as const, symbolSize: 5 }
const defaultLineSymbol = { symbol: 'circle' as const, symbolSize: 7 }
const largeLineSymbol = { symbol: 'none', sampling: 'lttb' as const }

export function sortValueTuples(
  tuples: [number, number][],
  enabled: boolean,
  order: 'asc' | 'desc'
): [number, number][] {
  if (!enabled) return tuples
  const sorted = [...tuples].sort((a, b) => a[1] - b[1])
  return order === 'asc' ? sorted : sorted.reverse()
}

export function scaleValueTuples(
  tuples: [number, number][],
  scale: ScaleType
): [number, number | null][] {
  if (scale !== 'log') return tuples
  return tuples.map(([x, y]) => [x, adjustForLogScaleLine(y, scale)])
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

  const label = {
    ...createLabelConfig(showLabels.value, styling),
    formatter: (p: { data: [number, number | null] }) => {
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
    itemStyle: { color: getNextColorFor(chartData.value.title) },
    ...seriesSymbol(chartType, largeX, config.symbol?.value, config.symbolSize?.value),
  }

  return {
    ...baseOptions,
    legend: { show: false },
    grid: createGridConfig(1, false),
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
