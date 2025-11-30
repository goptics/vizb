import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { getNextColorFor, hasXAxis, hasYAxis } from '../../lib/utils'
import {
  createAxisConfig,
  createGridConfig,
  createLabelConfig,
  createLegendConfig,
  createTooltipConfig,
  getChartStyling,
} from './shared'
import { sortByTotal } from './shared/common'

export function useBarChartOptions(config: BaseChartConfig) {
  const { chartData, sort, showLabels, isDark } = config

  const sortedData = computed(() => {
    // Check if we have y-axis data (dual categories)
    if (!sort.value.enabled) {
      return {
        series: chartData.value.series,
        xAxisData: chartData.value.series.map((s) => s.xAxis), // Always use framework names on x-axis
        hasYAxis: hasYAxis(chartData),
      }
    }

    // Sort series by their total values
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

  const options = computed<EChartsOption>(() => {
    const { series, xAxisData, hasYAxis } = sortedData.value
    const baseOptions = getBaseOptions(config)
    const styling = getChartStyling(isDark.value)

    // For single category: create one series with each x-axis value as a data point
    // For dual categories: create multiple series (one per x-axis value)
    if (!hasYAxis) {
      // Single category case: one series with multiple x-axis points
      return {
        ...baseOptions,
        grid: createGridConfig(1),
        tooltip: createTooltipConfig(false),
        legend: { show: false },
        ...createAxisConfig(styling, xAxisData),
        series: [
          {
            name: chartData.value.title,
            type: 'bar' as const,
            data: series.map((seriesData) => ({
              value: seriesData.values[0], // Take the first (and only) value
              label: createLabelConfig(showLabels.value, styling),
            })),
            itemStyle: { color: getNextColorFor(chartData.value.title) },
          },
        ],
      } as EChartsOption
    }

    // Dual categories case: transpose data to show y-axis values as series
    // Each y-axis value becomes a bar group, with x-axis values (frameworks) as bars
    const yAxisLabels = chartData.value.yAxis
    const transposedSeries = yAxisLabels.map((yAxisLabel, yIndex) => ({
      name: yAxisLabel,
      type: 'bar' as const,
      data: series.map((seriesData) => ({
        value: seriesData.values[yIndex] || 0,
        label: createLabelConfig(showLabels.value, styling),
      })),
      itemStyle: { color: getNextColorFor(yAxisLabel) },
    }))

    // Sort y-axis groups if there's only one x-axis group
    if (sort.value.enabled && xAxisData.length === 1) {
      transposedSeries.sort((a, b) => {
        const valA = a.data[0]?.value || 0
        const valB = b.data[0]?.value || 0
        if (sort.value.order === 'asc') {
          return valA - valB
        }

        return valB - valA
      })
    }

    const hasMultipleSeries = transposedSeries.length > 1

    return {
      ...baseOptions,
      grid: createGridConfig(transposedSeries.length),
      tooltip: createTooltipConfig(hasXAxis(chartData), transposedSeries.length),
      legend: createLegendConfig(
        transposedSeries.map((s) => ({ xAxis: s.name })),
        styling,
        hasMultipleSeries
      ),
      ...createAxisConfig(styling, xAxisData),
      series: transposedSeries,
    } as EChartsOption
  })

  return { options }
}
