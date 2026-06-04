import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { getNextColorFor, hasXAxis } from '../../lib/utils'
import {
  createAxisConfig,
  createGridConfig,
  createLegendConfig,
  createTooltipConfig,
  getChartStyling,
} from './shared'
import {
  useSortedSeriesData,
  getEffectiveScale,
  computeSeriesTotals,
} from './shared/common'
import { makeDataItem } from './shared/seriesConfig'

const barNullable = (val: number, scale: string): number | null =>
  scale === 'log' && val === 0 ? null : val

export function useBarChartOptions(config: BaseChartConfig) {
  const { chartData, sort, showLabels, isDark, scale } = config

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
        tooltip: createTooltipConfig(false, 1, isDark.value),
        legend: { show: false },
        ...createAxisConfig(styling, xAxisData, effectiveScale, minValue),
        series: [
          {
            name: chartData.value.title,
            type: 'bar' as const,
            data: series.map((s) =>
              makeDataItem(barNullable(s.values[0] ?? 0, effectiveScale), showLabels.value, styling)
            ),
            itemStyle: { color: getNextColorFor(chartData.value.title) },
          },
        ],
      } as EChartsOption
    }

    // Dual categories: transpose — each y-axis value becomes a bar group
    const yAxisLabels = chartData.value.yAxis
    const transposedSeries = yAxisLabels.map((yAxisLabel, yIndex) => ({
      name: yAxisLabel,
      type: 'bar' as const,
      data: series.map((s) =>
        makeDataItem(barNullable(s.values[yIndex] || 0, effectiveScale), showLabels.value, styling)
      ),
      itemStyle: { color: getNextColorFor(yAxisLabel) },
    }))

    // Secondary sort when there is only one x-group (sort within the group)
    if (sort.value.enabled && xAxisData.length === 1) {
      transposedSeries.sort((a, b) => {
        const valA = a.data[0]?.value || 0
        const valB = b.data[0]?.value || 0
        return sort.value.order === 'asc' ? valA - valB : valB - valA
      })
    }

    const hasMultipleSeries = transposedSeries.length > 1
    const seriesTotals = computeSeriesTotals(transposedSeries)

    return {
      ...baseOptions,
      grid: createGridConfig(transposedSeries.length),
      tooltip: createTooltipConfig(
        hasXAxis(chartData),
        transposedSeries.length,
        isDark.value,
        seriesTotals
      ),
      legend: createLegendConfig(
        transposedSeries.map((s) => ({ xAxis: s.name })),
        styling,
        hasMultipleSeries
      ),
      ...createAxisConfig(styling, xAxisData, effectiveScale, minValue),
      series: transposedSeries,
    } as EChartsOption
  })

  return { options }
}
