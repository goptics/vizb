import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { getNextColorFor } from '../../lib/utils'
import {
  createAxisConfig,
  createGridConfig,
  createLegendConfig,
  createTooltipConfig,
  getChartStyling,
  makeLegendTitle,
} from './shared'
import {
  adjustForLogScaleLine,
  useSortedSeriesData,
  getEffectiveScale,
  computeSeriesTotals,
} from './shared/common'
import { makeDataItem } from './shared/seriesConfig'

const symbolSize = 7
const symbol = 'circle'

export function useLineChartOptions(config: BaseChartConfig) {
  const { chartData, sort, isDark, showLabels, scale } = config

  const sortedData = useSortedSeriesData(chartData, sort)

  const options = computed<EChartsOption>(() => {
    const { series, xAxisData, hasYAxis } = sortedData.value
    const baseOptions = getBaseOptions(config)
    const styling = getChartStyling(isDark.value)
    const { minValue, effectiveScale } = getEffectiveScale(series, scale.value)

    if (!hasYAxis) {
      return {
        ...baseOptions,
        grid: createGridConfig(1),
        tooltip: createTooltipConfig(false, isDark.value),
        ...createAxisConfig(styling, xAxisData, effectiveScale, minValue, chartData.value.axisLabels?.x),
        legend: { show: false },
        series: [
          {
            name: chartData.value.title,
            type: 'line' as const,
            smooth: true,
            data: series.map((s) =>
              makeDataItem(
                adjustForLogScaleLine(s.values[0] ?? 0, effectiveScale),
                showLabels.value,
                styling
              )
            ),
            connectNulls: true,
            itemStyle: { color: getNextColorFor(chartData.value.title) },
            symbol,
            symbolSize,
          },
        ],
      } as EChartsOption
    }

    // Dual categories: transpose — each y-axis value becomes a line series
    const yAxisLabels = chartData.value.yAxis
    const transposedSeries = yAxisLabels.map((yAxisLabel, yIndex) => ({
      name: yAxisLabel,
      type: 'line' as const,
      smooth: true,
      data: series.map((s) =>
        makeDataItem(
          adjustForLogScaleLine(s.values[yIndex] ?? 0, effectiveScale),
          showLabels.value,
          styling
        )
      ),
      connectNulls: true,
      itemStyle: { color: getNextColorFor(yAxisLabel) },
      symbol,
      symbolSize,
    }))

    const seriesTotals = computeSeriesTotals(transposedSeries)

    // The legend encodes the y group; title it above the legend when known.
    const yLabel = chartData.value.axisLabels?.y

    return {
      ...baseOptions,
      ...(yLabel ? { title: makeLegendTitle(yLabel, styling) } : {}),
      grid: createGridConfig(transposedSeries.length),
      tooltip: createTooltipConfig(true, isDark.value, seriesTotals),
      ...createAxisConfig(styling, xAxisData, effectiveScale, minValue, chartData.value.axisLabels?.x),
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
