import type { EChartsOption } from 'echarts'
import type { ScaleType } from '@/types'
import { getNextColorFor } from '@/lib/utils'
import { type BaseChartConfig, getBaseOptions } from '../baseChartOptions'
import {
  createAxisConfig,
  createLabelConfig,
  createValueModeGridConfig,
  getChartStyling,
  getTooltipTheme,
  isLargeXAxis,
  scatterSeriesLargeOpts,
} from './chartConfig'
import { adjustForLogScaleLine, getEffectiveScale } from './common'
import { resolveSeriesSymbol } from './seriesConfig'
import { resolve2DScatterVisualMap } from './visualMap'

const defaultScatterSymbol = { symbol: 'circle' as const, symbolSize: 8 }
const largeScatterSymbol = { symbol: 'circle' as const, symbolSize: 5 }

export function scaleMixedTuples(
  tuples: [number, number][],
  scale: ScaleType
): [number, number | null][] {
  if (scale !== 'log') return tuples
  return tuples.map(([x, y]) => [x, adjustForLogScaleLine(y, scale)])
}

export function createMixedModeTooltip(
  isDark: boolean,
  xCategories: string[],
  xLabel?: string,
  yLabel?: string
): EChartsOption['tooltip'] {
  const theme = getTooltipTheme(isDark)
  const xName = xLabel ?? 'x'
  const yName = yLabel ?? 'y'

  return {
    trigger: 'item',
    ...theme,
    axisPointer: { type: 'shadow' },
    formatter: (params: unknown) => {
      const [xi, y] = (params as { data: [number, number | null] }).data
      const category = xCategories[xi] ?? String(xi)
      const yText = y === null || y === undefined ? '' : String(Math.round(y * 100) / 100)
      return `<strong>${xName}: ${category}</strong><br/>${yName}: <b>${yText}</b>`
    },
  }
}

/** Mixed-axis 2D scatter: category x + value y, one point per row. */
export function buildMixedAxes2DOptions(
  config: BaseChartConfig,
  chartType: 'scatter' = 'scatter'
): EChartsOption {
  const { chartData, showLabels, isDark, scale, visualMap } = config
  const tuples = chartData.value.mixedTuples ?? []
  const xCategories = chartData.value.xCategories ?? []
  const xLabel = chartData.value.axisLabels?.x
  const yLabel = chartData.value.axisLabels?.y
  const baseOptions = getBaseOptions(config)
  const styling = getChartStyling(isDark.value)
  const effectiveScale = scale?.value ?? 'linear'
  const { minValue, effectiveScale: yScale } = getEffectiveScale(
    [{ xAxis: chartData.value.title, values: tuples.map(([, y]) => y), benchmarkId: '' }],
    effectiveScale
  )

  const data = scaleMixedTuples(tuples, yScale)
  const largeX = isLargeXAxis(xCategories)
  const useVisualMap = chartType === 'scatter' && visualMap?.value === true
  const colorValues = tuples.map(([, y]) => y).filter((v) => isFinite(v))

  const label = {
    ...createLabelConfig(showLabels.value, styling),
    formatter: (p: { data: [number, number | null] }) => {
      const y = p.data[1]
      return y === null || y === undefined ? '' : String(Math.round(y * 100) / 100)
    },
  }

  const series = {
    name: chartData.value.title,
    type: 'scatter' as const,
    data,
    label,
    ...scatterSeriesLargeOpts(useVisualMap),
    ...(useVisualMap ? {} : { itemStyle: { color: getNextColorFor(chartData.value.title) } }),
    ...resolveSeriesSymbol(
      largeX ? largeScatterSymbol : defaultScatterSymbol,
      config.symbol?.value,
      config.symbolSize?.value
    ),
  }

  const axisConfig = createAxisConfig(
    styling,
    xCategories,
    yScale,
    minValue,
    xLabel,
    largeX,
    chartType === 'scatter'
  )

  return {
    ...baseOptions,
    legend: { show: false },
    grid: createValueModeGridConfig(largeX),
    visualMap: resolve2DScatterVisualMap(useVisualMap, colorValues, styling, 1),
    tooltip: createMixedModeTooltip(isDark.value, xCategories, xLabel, yLabel),
    ...axisConfig,
    dataZoom: largeX
      ? [
          { type: 'inside', xAxisIndex: 0 },
          { type: 'slider', xAxisIndex: 0, bottom: 34, height: 28 },
        ]
      : [
          { type: 'inside', xAxisIndex: 0 },
          { type: 'inside', yAxisIndex: 0 },
        ],
    series: [series],
  } as EChartsOption
}
