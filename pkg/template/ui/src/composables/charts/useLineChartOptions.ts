import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { getNextColorFor, hasYAxis } from '../../lib/utils'
import {
  createAxisConfig,
  createGridConfig,
  createLabelConfig,
  createLegendConfig,
  createTooltipConfig,
  getChartStyling,
} from './shared'
import { sortByTotal } from './shared/common'

export function useLineChartOptions(config: BaseChartConfig) {
  const { chartData, sort, showLabels, isDark } = config

  const sortedData = computed(() => {
    if (!sort.value.enabled) {
      return {
        series: chartData.value.series,
        xAxisData: chartData.value.series.map((s) => s.xAxis), // Always use framework names on x-axis
        hasYAxis: hasYAxis(chartData),
      }
    }

    // Sort series by total values
    const seriesWithTotals = chartData.value.series.map((series) => ({
      ...series,
      total: series.values.reduce((sum, val) => sum + val, 0),
    }))

    if (sort.value.enabled) {
      seriesWithTotals.sort(sortByTotal(sort.value.order))
    }

    return {
      series: seriesWithTotals,
      xAxisData: seriesWithTotals.map((s) => s.xAxis), // Always use framework names on x-axis
      hasYAxis: hasYAxis(chartData),
    }
  })
  const symbolSize = 7
  const symbol = 'circle'
  
  const options = computed<EChartsOption>(() => {
    const { series, xAxisData, hasYAxis } = sortedData.value
    const baseOptions = getBaseOptions(config)
    const styling = getChartStyling(isDark.value)

    // Single category case: one series with multiple x-axis points
    if (!hasYAxis) {
      return {
        ...baseOptions,
        grid: createGridConfig(1),
        tooltip: createTooltipConfig(false, 1),
        ...createAxisConfig(styling, xAxisData),
        legend: { show: false },
        series: [
          {
            name: chartData.value.title,
            type: 'line' as const,
            data: series.map((seriesData) => ({
              value: seriesData.values[0],
              label: createLabelConfig(showLabels.value, styling),
            })),
            itemStyle: { color: getNextColorFor(chartData.value.title) },
            symbol,
            symbolSize,
          },
        ],
      } as EChartsOption
    }

    // Dual categories case: transpose data to show y-axis values as series
    const yAxisLabels = chartData.value.yAxis
    const transposedSeries = yAxisLabels.map((yAxisLabel, yIndex) => ({
      name: yAxisLabel,
      type: 'line' as const,
      data: series.map((seriesData) => ({
        value: seriesData.values[yIndex] || 0,
        label: createLabelConfig(showLabels.value, styling),
      })),
      itemStyle: { color: getNextColorFor(yAxisLabel)},
      symbol,
      symbolSize,
    }))

    return {
      ...baseOptions,
      grid: createGridConfig(transposedSeries.length),
      tooltip: createTooltipConfig(true, transposedSeries.length),
      ...createAxisConfig(styling, xAxisData),
      legend: createLegendConfig(
        transposedSeries.map((s) => ({ xAxis: s.name })),
        styling,
        true
      ),
      series: transposedSeries,
    } as EChartsOption
  })

  return { options }
}
