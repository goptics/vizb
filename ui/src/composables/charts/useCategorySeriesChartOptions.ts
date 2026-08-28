import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { getNextColorFor, hasXAxis } from '@/lib/utils'
import {
  createAxisConfig,
  createGridConfig,
  createValueModeGridConfig,
  createLabelConfig,
  createLegendConfig,
  createPinnedAxisTooltip,
  createTooltipConfig,
  createValueAxisConfig,
  createValueModeTooltip,
  getChartStyling,
  isLargeXAxis,
  makeLegendTitle,
  LARGE_DATA_THRESHOLD,
  resolveCartesianDataZoom,
  scatterSeriesLargeOpts,
} from './shared/chartConfig'
import {
  adjustForLogScaleLine,
  useSortedSeriesData,
  resolveLogScale,
  computeSeriesTotals,
} from './shared/common'
import { asLogXPairs, axisIsLog, axisLogBase, numericLogXValues, parseScale } from '@/lib/scale'
import { resolveSeriesSymbol } from './shared/seriesConfig'
import { resolve2DScatterVisualMap } from './shared/visualMap'
import { buildValueAxes2DOptions } from './shared/valueMode'
import { buildMixedAxes2DOptions } from './shared/mixedMode'

export type CategorySeriesKind = 'line' | 'scatter'

const SERIES_STYLE: Record<
  CategorySeriesKind,
  {
    defaultSymbol: { symbol: 'circle'; symbolSize: number }
    largeSymbol: { symbol: 'circle'; symbolSize: number }
    connectNulls?: true
  }
> = {
  line: {
    defaultSymbol: { symbol: 'circle', symbolSize: 7 },
    largeSymbol: { symbol: 'circle', symbolSize: 7 },
    connectNulls: true,
  },
  scatter: {
    defaultSymbol: { symbol: 'circle', symbolSize: 8 },
    largeSymbol: { symbol: 'circle', symbolSize: 5 },
  },
}

type SeriesPoint = number | null | [number, number | null]

const groupedScatterColorValues = (seriesList: { data: SeriesPoint[] }[]): number[] => {
  const vals: number[] = []
  for (const s of seriesList) {
    for (const v of s.data) {
      const n = Array.isArray(v) ? v[1] : v
      if (n != null && isFinite(n)) vals.push(n)
    }
  }
  return vals
}

const logYValue = (val: number | null, yLog: boolean): number | null =>
  adjustForLogScaleLine(val, yLog ? 'log' : 'linear')

export function useCategorySeriesChartOptions(config: BaseChartConfig, kind: CategorySeriesKind) {
  const { chartData, sort, isDark, showLabels, scale, stack, visualMap } = config
  const sortedData = useSortedSeriesData(chartData, sort)
  const style = SERIES_STYLE[kind]

  const options = computed<EChartsOption>(() => {
    if (chartData.value.mixedTuples?.length) {
      return buildMixedAxes2DOptions(config, kind)
    }
    if (chartData.value.valueTuples?.length) {
      return buildValueAxes2DOptions(config, kind)
    }

    const { series, xAxisData, hasYAxis } = sortedData.value
    const baseOptions = getBaseOptions(config)
    const styling = getChartStyling(isDark.value)
    const parsed = parseScale(scale?.value)
    const xWant = axisIsLog(parsed, 'x', ['y'])
    const yWant = axisIsLog(parsed, 'y', ['y'])
    const yScale = resolveLogScale(
      yWant ? 'log' : 'linear',
      series.flatMap((s) => s.values)
    )
    const yLog = yScale === 'log'
    const xNums = numericLogXValues(xAxisData, xWant)
    const coerceX = xNums !== null
    const largeX = isLargeXAxis(xAxisData)
    const xLabel = chartData.value.axisLabels?.x
    const yLogBase = axisLogBase(parsed, 'y')
    const groupedAxes = coerceX
      ? createValueAxisConfig(styling, xLabel, undefined, yScale, true, {
          xScale: 'log',
          xLogBase: axisLogBase(parsed, 'x'),
          yLogBase,
        })
      : createAxisConfig(styling, xAxisData, yScale, xLabel, largeX, true, yLogBase)
    const zoom = resolveCartesianDataZoom(kind, {
      numericX: coerceX,
      largeX,
      styling,
    })
    const seriesExtras = resolveSeriesSymbol(
      largeX ? style.largeSymbol : style.defaultSymbol,
      config.symbol?.value,
      config.symbolSize?.value
    )
    const useVisualMap = kind === 'scatter' && visualMap?.value === true
    const smoothLines = kind === 'line' && config.smooth?.value === true
    const useStack = kind === 'line' && stack?.value === true && yScale !== 'log'

    if (!hasYAxis) {
      const singleSeries = {
        name: chartData.value.title,
        type: kind,
        data: xNums
          ? asLogXPairs(
              xNums,
              series.map((s) => s.values[0] ?? null),
              yLog
            )
          : series.map((s) => logYValue(s.values[0] ?? null, yLog)),
        label: createLabelConfig(showLabels.value, styling),
        ...(kind === 'scatter'
          ? scatterSeriesLargeOpts(useVisualMap)
          : { large: true as const, largeThreshold: LARGE_DATA_THRESHOLD }),
        ...(style.connectNulls ? { connectNulls: true } : {}),
        ...(kind === 'line' ? { smooth: smoothLines, showAllSymbol: true as const } : {}),
        ...(useVisualMap ? {} : { itemStyle: { color: getNextColorFor(chartData.value.title) } }),
        ...seriesExtras,
      }
      return {
        ...baseOptions,
        grid: createValueModeGridConfig(zoom.hasXSlider),
        tooltip: coerceX
          ? createValueModeTooltip(isDark.value, xLabel, chartData.value.axisLabels?.y, true)
          : createPinnedAxisTooltip(isDark.value),
        ...groupedAxes,
        ...(zoom.dataZoom ? { dataZoom: zoom.dataZoom } : {}),
        legend: { show: false },
        visualMap: resolve2DScatterVisualMap(
          useVisualMap,
          groupedScatterColorValues([singleSeries]),
          styling,
          1
        ),
        series: [singleSeries],
      } as EChartsOption
    }

    const yAxisLabels = chartData.value.yAxis
    const transposedSeries = yAxisLabels.map((yAxisLabel, yIndex) => ({
      name: yAxisLabel,
      type: kind,
      data: xNums
        ? asLogXPairs(
            xNums,
            series.map((s) => s.values[yIndex] ?? null),
            yLog
          )
        : series.map((s) => logYValue(s.values[yIndex] ?? null, yLog)),
      label: createLabelConfig(showLabels.value, styling),
      ...(kind === 'scatter'
        ? scatterSeriesLargeOpts(useVisualMap)
        : { large: true as const, largeThreshold: LARGE_DATA_THRESHOLD }),
      ...(style.connectNulls ? { connectNulls: true } : {}),
      ...(kind === 'line'
        ? {
            smooth: smoothLines,
            stack: useStack ? 'total' : null,
            areaStyle: useStack ? {} : null,
            showAllSymbol: true as const,
          }
        : {}),
      ...(useVisualMap ? {} : { itemStyle: { color: getNextColorFor(yAxisLabel) } }),
      ...seriesExtras,
    }))

    const seriesTotals = computeSeriesTotals(transposedSeries)
    const yLabel = chartData.value.axisLabels?.y
    const showXBreakdown = kind === 'line' || hasXAxis(chartData)

    return {
      ...baseOptions,
      ...(yLabel ? { title: makeLegendTitle(yLabel, styling) } : {}),
      grid: createGridConfig(transposedSeries.length, zoom.hasXSlider, !!yLabel),
      visualMap: resolve2DScatterVisualMap(
        useVisualMap,
        groupedScatterColorValues(transposedSeries),
        styling,
        1
      ),
      tooltip: coerceX
        ? transposedSeries.length <= 1
          ? createValueModeTooltip(isDark.value, xLabel, yLabel, true)
          : createTooltipConfig(showXBreakdown, isDark.value, seriesTotals, 'line')
        : createTooltipConfig(showXBreakdown, isDark.value, seriesTotals),
      ...groupedAxes,
      ...(zoom.dataZoom ? { dataZoom: zoom.dataZoom } : {}),
      legend: createLegendConfig(
        transposedSeries.map((s) => ({ xAxis: s.name })),
        styling,
        true,
        yLabel ? { top: 24 } : undefined
      ),
      series: transposedSeries,
    } as EChartsOption
  })

  return { options }
}
