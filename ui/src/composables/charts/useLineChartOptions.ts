import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { getNextColorFor } from '../../lib/utils'
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

const defaultSymbol = { symbol: 'circle', symbolSize: 7 }
const largeSymbol = { symbol: 'none', sampling: 'lttb' }

export function useLineChartOptions(config: BaseChartConfig) {
  const { chartData, sort, isDark, showLabels, scale } = config

  const sortedData = useSortedSeriesData(chartData, sort)

  const options = computed<EChartsOption>(() => {
    const { series, xAxisData, hasYAxis } = sortedData.value
    const baseOptions = getBaseOptions(config)
    const styling = getChartStyling(isDark.value)
    // `scale` is optional on BaseChartConfig (relaxed in Task 7) — pie/heatmap/
    // radar pass a config without it. The line composable is the only consumer,
    // so we default at the call site.
    const { minValue, effectiveScale } = getEffectiveScale(series, scale?.value ?? 'linear')
    const largeX = isLargeXAxis(xAxisData)

    if (!hasYAxis) {
      return {
        ...baseOptions,
        grid: createGridConfig(1, largeX),
        tooltip: createPinnedAxisTooltip(isDark.value),
        ...createAxisConfig(
          styling,
          xAxisData,
          effectiveScale,
          minValue,
          chartData.value.axisLabels?.x
        ),
        ...(largeX ? { dataZoom: createDataZoomConfig(xAxisData, styling) } : {}),
        legend: { show: false },
        series: [
          {
            name: chartData.value.title,
            type: 'line' as const,
            data: series.map((s) => adjustForLogScaleLine(s.values[0] ?? 0, effectiveScale)),
            label: createLabelConfig(showLabels.value, styling),
            large: true,
            largeThreshold: LARGE_DATA_THRESHOLD,
            connectNulls: true,
            itemStyle: { color: getNextColorFor(chartData.value.title) },
            ...(largeX ? largeSymbol : defaultSymbol),
          },
        ],
      } as EChartsOption
    }

    // Dual categories: transpose — each y-axis value becomes a line series
    const yAxisLabels = chartData.value.yAxis
    const transposedSeries = yAxisLabels.map((yAxisLabel, yIndex) => ({
      name: yAxisLabel,
      type: 'line' as const,
      data: series.map((s) => adjustForLogScaleLine(s.values[yIndex] ?? 0, effectiveScale)),
      label: createLabelConfig(showLabels.value, styling),
      large: true,
      largeThreshold: LARGE_DATA_THRESHOLD,
      connectNulls: true,
      itemStyle: { color: getNextColorFor(yAxisLabel) },
      ...(largeX ? largeSymbol : defaultSymbol),
    }))

    const seriesTotals = computeSeriesTotals(transposedSeries)

    // The legend encodes the y group; title it above the legend when known.
    const yLabel = chartData.value.axisLabels?.y

    return {
      ...baseOptions,
      ...(yLabel ? { title: makeLegendTitle(yLabel, styling) } : {}),
      grid: createGridConfig(transposedSeries.length, largeX),
      tooltip: createTooltipConfig(true, isDark.value, seriesTotals),
      ...createAxisConfig(
        styling,
        xAxisData,
        effectiveScale,
        minValue,
        chartData.value.axisLabels?.x
      ),
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
