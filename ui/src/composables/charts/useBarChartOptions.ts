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
  createTooltipConfig,
  getChartStyling,
  isLargeXAxis,
  makeLegendTitle,
  LARGE_DATA_THRESHOLD,
} from './shared'
import { useSortedSeriesData, getEffectiveScale, computeSeriesTotals } from './shared/common'

const barNullable = (val: number | null, scale: string): number | null =>
  val === null ? null : scale === 'log' && val <= 0 ? null : val

export function useBarChartOptions(config: BaseChartConfig) {
  const { chartData, sort, showLabels, isDark, scale } = config

  const sortedData = useSortedSeriesData(chartData, sort)

  const options = computed<EChartsOption>(() => {
    const { series, xAxisData, hasYAxis } = sortedData.value
    const baseOptions = getBaseOptions(config)
    const styling = getChartStyling(isDark.value)
    // `scale` is optional on BaseChartConfig (relaxed in Task 7) — pie/heatmap/
    // radar pass a config without it. The bar composable is the only consumer,
    // so we default at the call site.
    const { minValue, effectiveScale } = getEffectiveScale(series, scale?.value ?? 'linear')
    const largeX = isLargeXAxis(xAxisData)

    if (!hasYAxis) {
      return {
        ...baseOptions,
        grid: createGridConfig(1, largeX),
        tooltip: createTooltipConfig(false, isDark.value),
        legend: { show: false },
        ...createAxisConfig(
          styling,
          xAxisData,
          effectiveScale,
          minValue,
          chartData.value.axisLabels?.x
        ),
        ...(largeX ? { dataZoom: createDataZoomConfig(xAxisData, styling) } : {}),
        series: [
          {
            name: chartData.value.title,
            type: 'bar' as const,
            // Plain values + one series-level label, not a per-point {value,label}
            // object — a 100k-bar chart would otherwise allocate 100k label configs
            // on every recompute. `large` keeps the draw on one frame past the
            // threshold.
            data: series.map((s) => barNullable(s.values[0] ?? null, effectiveScale)),
            label: createLabelConfig(showLabels.value, styling),
            large: true,
            largeThreshold: LARGE_DATA_THRESHOLD,
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
      data: series.map((s) => barNullable(s.values[yIndex] ?? null, effectiveScale)),
      label: createLabelConfig(showLabels.value, styling),
      large: true,
      largeThreshold: LARGE_DATA_THRESHOLD,
      itemStyle: { color: getNextColorFor(yAxisLabel) },
    }))

    // Secondary sort when there is only one x-group (sort within the group).
    // data items are now plain numbers (or null).
    if (sort.value.enabled && xAxisData.length === 1) {
      transposedSeries.sort((a, b) => {
        const valA = a.data[0] ?? 0
        const valB = b.data[0] ?? 0
        return sort.value.order === 'asc' ? valA - valB : valB - valA
      })
    }

    const hasMultipleSeries = transposedSeries.length > 1
    const seriesTotals = computeSeriesTotals(transposedSeries)

    // The legend encodes the y group; title it above the legend when known.
    const yLabel = chartData.value.axisLabels?.y
    const showLegendTitle = hasMultipleSeries && !!yLabel

    return {
      ...baseOptions,
      ...(showLegendTitle ? { title: makeLegendTitle(yLabel!, styling) } : {}),
      grid: createGridConfig(transposedSeries.length, largeX),
      tooltip: createTooltipConfig(hasXAxis(chartData), isDark.value, seriesTotals),
      legend: createLegendConfig(
        transposedSeries.map((s) => ({ xAxis: s.name })),
        styling,
        hasMultipleSeries,
        showLegendTitle ? { top: 24 } : undefined
      ),
      ...createAxisConfig(
        styling,
        xAxisData,
        effectiveScale,
        minValue,
        chartData.value.axisLabels?.x
      ),
      ...(largeX ? { dataZoom: createDataZoomConfig(xAxisData, styling) } : {}),
      series: transposedSeries,
    } as EChartsOption
  })

  return { options }
}
