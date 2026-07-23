import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { getNextColorFor, hasXAxis } from '@/lib/utils'
import {
  createAxisConfig,
  createDataZoomConfig,
  createGridConfig,
  createHorizontalAxisConfig,
  createHorizontalDataZoomConfig,
  createLabelConfig,
  createLegendConfig,
  createTooltipConfig,
  getChartStyling,
  horizontalLegendBottom,
  isLargeXAxis,
  makeLegendTitle,
  LARGE_DATA_THRESHOLD,
} from './shared'
import { useSortedSeriesData, getEffectiveScale, computeSeriesTotals } from './shared/common'
import { buildValueAxes2DOptions } from './shared/valueMode'
import { buildMixedAxes2DOptions } from './shared/mixedMode'
import { percentageFormatter } from './shared/labels'

const barNullable = (val: number | null, scale: string): number | null =>
  val === null ? null : scale === 'log' && val <= 0 ? null : val

export function useBarChartOptions(config: BaseChartConfig) {
  const { chartData, sort, showLabels, labelMode, chartTotal, isDark, scale, stack, horizontal } =
    config

  const sortedData = useSortedSeriesData(chartData, sort)

  const options = computed<EChartsOption>(() => {
    const isHorizontal = horizontal?.value ?? false

    if (chartData.value.mixedTuples?.length) {
      return buildMixedAxes2DOptions(config, 'bar')
    }
    if (chartData.value.valueTuples?.length) {
      return buildValueAxes2DOptions(config, 'bar')
    }

    const { series, xAxisData, hasYAxis } = sortedData.value
    const baseOptions = getBaseOptions(config)
    const styling = getChartStyling(isDark.value)
    // `scale` is optional on BaseChartConfig (relaxed in Task 7) — pie/heatmap/
    // radar pass a config without it. The bar composable is the only consumer,
    // so we default at the call site.
    const { minValue, effectiveScale } = getEffectiveScale(series, scale?.value ?? 'linear')
    const largeX = isLargeXAxis(xAxisData)
    const xLabel = chartData.value.axisLabels?.x
    const useStack = stack?.value === true && effectiveScale !== 'log'
    const formatter = percentageFormatter(
      labelMode?.value ?? 'none',
      chartTotal?.value ?? 0,
      (p: any) =>
        typeof p.value === 'number' ? p.value : Array.isArray(p.value) ? p.value.at(-1) : undefined
    )

    if (!hasYAxis && isHorizontal) {
      return {
        ...baseOptions,
        grid: {
          left: xLabel ? 70 : '3%',
          right: largeX ? 44 : 24,
          bottom: '3%',
          top: 8,
          containLabel: true,
        },
        tooltip: createTooltipConfig(false, isDark.value),
        legend: { show: false },
        ...createHorizontalAxisConfig(styling, xAxisData, effectiveScale, minValue, xLabel, largeX),
        ...(largeX ? { dataZoom: createHorizontalDataZoomConfig(styling) } : {}),
        series: [
          {
            name: chartData.value.title,
            type: 'bar' as const,
            data: series.map((s) => barNullable(s.values[0] ?? null, effectiveScale)),
            label: createLabelConfig(showLabels.value, styling, 'horizontal', false, formatter),
            large: true,
            largeThreshold: LARGE_DATA_THRESHOLD,
            itemStyle: { color: getNextColorFor(chartData.value.title) },
          },
        ],
      } as EChartsOption
    }

    if (!hasYAxis) {
      return {
        ...baseOptions,
        grid: createGridConfig(1, largeX),
        tooltip: createTooltipConfig(false, isDark.value),
        legend: { show: false },
        ...createAxisConfig(styling, xAxisData, effectiveScale, minValue, xLabel, largeX),
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
            label: createLabelConfig(showLabels.value, styling, undefined, false, formatter),
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
      label: createLabelConfig(
        showLabels.value,
        styling,
        isHorizontal ? 'horizontal' : 'vertical',
        useStack,
        formatter
      ),
      large: true,
      largeThreshold: LARGE_DATA_THRESHOLD,
      ...(useStack ? { stack: 'total' } : {}),
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

    const yLabel = chartData.value.axisLabels?.y
    // Vertical: legend encodes the y group with yLabel as a top title.
    const showLegendTitle = hasMultipleSeries && !!yLabel

    if (isHorizontal) {
      return {
        ...baseOptions,
        grid: {
          left: xLabel ? 70 : '3%',
          right: largeX ? 44 : 24,
          bottom: hasMultipleSeries ? horizontalLegendBottom(transposedSeries.length) : '3%',
          top: 8,
          containLabel: true,
        },
        tooltip: createTooltipConfig(hasXAxis(chartData), isDark.value, seriesTotals),
        legend: createLegendConfig(
          transposedSeries.map((s) => ({ xAxis: s.name })),
          styling,
          hasMultipleSeries,
          { bottom: 0 }
        ),
        ...createHorizontalAxisConfig(styling, xAxisData, effectiveScale, minValue, xLabel, largeX),
        ...(largeX ? { dataZoom: createHorizontalDataZoomConfig(styling) } : {}),
        series: transposedSeries,
      } as EChartsOption
    }

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
      ...createAxisConfig(styling, xAxisData, effectiveScale, minValue, xLabel, largeX),
      ...(largeX ? { dataZoom: createDataZoomConfig(xAxisData, styling) } : {}),
      series: transposedSeries,
    } as EChartsOption
  })

  return { options }
}
