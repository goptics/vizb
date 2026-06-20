import type { EChartsOption } from 'echarts'
import type { ScaleType } from '@/types'
import { getNextColorFor } from '@/lib/utils'
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

const defaultLineSymbol = { symbol: 'circle' as const, symbolSize: 7 }
const largeLineSymbol = { symbol: 'none' as const, sampling: 'lttb' as const }
const defaultScatterSymbol = { symbol: 'circle' as const, symbolSize: 8 }
const largeScatterSymbol = { symbol: 'circle' as const, symbolSize: 5 }

const barNullableY = (y: number, scale: ScaleType): number | null =>
  scale === 'log' && y <= 0 ? null : y

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
  scale: ScaleType,
  seriesType: 'line' | 'bar' | 'scatter'
): [number, number | null][] {
  if (scale !== 'log') return tuples
  return tuples.map(([x, y]) => {
    const scaledY = seriesType === 'bar' ? barNullableY(y, scale) : adjustForLogScaleLine(y, scale)
    return [x, scaledY]
  })
}

function buildAxes2DValueOptions(
  config: BaseChartConfig,
  seriesType: 'line' | 'bar' | 'scatter'
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
  const data = scaleValueTuples(sorted, yScale, seriesType)
  const largeX = isLargeXAxis(data.map((_, i) => String(i)))

  const label = {
    ...createLabelConfig(showLabels.value, styling),
    formatter: (p: { data: [number, number | null] }) => {
      const y = p.data[1]
      return y === null || y === undefined ? '' : String(Math.round(y * 100) / 100)
    },
  }

  const seriesBase = {
    name: chartData.value.title,
    data,
    label,
    large: true,
    largeThreshold: LARGE_DATA_THRESHOLD,
    itemStyle: { color: getNextColorFor(chartData.value.title) },
  }

  const series =
    seriesType === 'scatter'
      ? {
          ...seriesBase,
          type: 'scatter' as const,
          ...(largeX ? largeScatterSymbol : defaultScatterSymbol),
        }
      : seriesType === 'line'
        ? {
            ...seriesBase,
            type: 'line' as const,
            connectNulls: true,
            showSymbol: data.length <= 200,
            ...(largeX ? largeLineSymbol : defaultLineSymbol),
          }
        : { ...seriesBase, type: 'bar' as const }

  return {
    ...baseOptions,
    legend: { show: false },
    grid: createGridConfig(1, false),
    tooltip: createValueModeTooltip(isDark.value, xLabel, yLabel),
    ...createValueAxisConfig(styling, xLabel, yLabel, yScale, minValue),
    dataZoom: [
      { type: 'inside', xAxisIndex: 0 },
      { type: 'inside', yAxisIndex: 0 },
    ],
    series: [series],
  } as EChartsOption
}

/** 2D scatter for --axes value mode (continuous x/y on value axes). */
export function buildScatterAxes2DOptions(config: BaseChartConfig): EChartsOption {
  return buildAxes2DValueOptions(config, 'scatter')
}

/** Shared 2D options for value-mode bar/line — inherits grouping chart theme + layout. */
export function buildValueMode2DOptions(
  config: BaseChartConfig,
  seriesType: 'line' | 'bar'
): EChartsOption {
  return buildAxes2DValueOptions(config, seriesType)
}
