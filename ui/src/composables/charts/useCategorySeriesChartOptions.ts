import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { getNextColorFor, hasXAxis } from '@/lib/utils'
import {
  createAxisConfig,
  createDataZoomConfig,
  createGridConfig,
  createLabelConfig,
  createLegendConfig,
  createPinnedAxisTooltip,
  createTooltipConfig,
  getChartStyling,
  isLargeXAxis,
  makeLegendTitle,
  LARGE_DATA_THRESHOLD,
} from './shared'
import {
  adjustForLogScaleLine,
  useSortedSeriesData,
  getEffectiveScale,
  computeSeriesTotals,
} from './shared/common'

export type CategorySeriesKind = 'line' | 'scatter'

const SERIES_STYLE: Record<
  CategorySeriesKind,
  {
    defaultSymbol: { symbol: 'circle'; symbolSize: number }
    largeSymbol: { symbol: 'circle' | 'none'; symbolSize?: number; sampling?: 'lttb' }
    connectNulls?: true
  }
> = {
  line: {
    defaultSymbol: { symbol: 'circle', symbolSize: 7 },
    largeSymbol: { symbol: 'none', sampling: 'lttb' },
    connectNulls: true,
  },
  scatter: {
    defaultSymbol: { symbol: 'circle', symbolSize: 8 },
    largeSymbol: { symbol: 'circle', symbolSize: 5 },
  },
}

export function useCategorySeriesChartOptions(config: BaseChartConfig, kind: CategorySeriesKind) {
  const { chartData, sort, isDark, showLabels, scale } = config
  const sortedData = useSortedSeriesData(chartData, sort)
  const style = SERIES_STYLE[kind]

  const options = computed<EChartsOption>(() => {
    const { series, xAxisData, hasYAxis } = sortedData.value
    const baseOptions = getBaseOptions(config)
    const styling = getChartStyling(isDark.value)
    const { minValue, effectiveScale } = getEffectiveScale(series, scale?.value ?? 'linear')
    const largeX = isLargeXAxis(xAxisData)
    const xLabel = chartData.value.axisLabels?.x
    const seriesExtras = largeX ? style.largeSymbol : style.defaultSymbol

    if (!hasYAxis) {
      return {
        ...baseOptions,
        grid: createGridConfig(1, largeX),
        tooltip: createPinnedAxisTooltip(isDark.value),
        ...createAxisConfig(styling, xAxisData, effectiveScale, minValue, xLabel, largeX, true),
        ...(largeX ? { dataZoom: createDataZoomConfig(xAxisData, styling) } : {}),
        legend: { show: false },
        series: [
          {
            name: chartData.value.title,
            type: kind,
            data: series.map((s) => adjustForLogScaleLine(s.values[0] ?? null, effectiveScale)),
            label: createLabelConfig(showLabels.value, styling),
            large: true,
            largeThreshold: LARGE_DATA_THRESHOLD,
            ...(style.connectNulls ? { connectNulls: true } : {}),
            itemStyle: { color: getNextColorFor(chartData.value.title) },
            ...seriesExtras,
          },
        ],
      } as EChartsOption
    }

    const yAxisLabels = chartData.value.yAxis
    const transposedSeries = yAxisLabels.map((yAxisLabel, yIndex) => ({
      name: yAxisLabel,
      type: kind,
      data: series.map((s) => adjustForLogScaleLine(s.values[yIndex] ?? null, effectiveScale)),
      label: createLabelConfig(showLabels.value, styling),
      large: true,
      largeThreshold: LARGE_DATA_THRESHOLD,
      ...(style.connectNulls ? { connectNulls: true } : {}),
      itemStyle: { color: getNextColorFor(yAxisLabel) },
      ...seriesExtras,
    }))

    const seriesTotals = computeSeriesTotals(transposedSeries)
    const yLabel = chartData.value.axisLabels?.y
    const showXBreakdown = kind === 'line' || hasXAxis(chartData)

    return {
      ...baseOptions,
      ...(yLabel ? { title: makeLegendTitle(yLabel, styling) } : {}),
      grid: createGridConfig(transposedSeries.length, largeX),
      tooltip: createTooltipConfig(showXBreakdown, isDark.value, seriesTotals),
      ...createAxisConfig(styling, xAxisData, effectiveScale, minValue, xLabel, largeX, true),
      ...(largeX ? { dataZoom: createDataZoomConfig(xAxisData, styling) } : {}),
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
