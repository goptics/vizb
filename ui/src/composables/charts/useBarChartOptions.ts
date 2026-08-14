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
} from './shared/chartConfig'
import { useSortedSeriesData, resolveLogScale, computeSeriesTotals } from './shared/common'
import { buildValueAxes2DOptions } from './shared/valueMode'
import { buildMixedAxes2DOptions } from './shared/mixedMode'

const barNullable = (val: number | null, scale: string): number | null =>
  val === null ? null : scale === 'log' && val <= 0 ? null : val

// ECharts itemStyle.borderRadius: [TL, TR, BR, BL] (len 1–4; [8] = all corners).
type BarItemStyle = { color?: string; borderRadius?: number[] }

function isActiveRadius(r: number[] | undefined): r is number[] {
  return !!r && r.length > 0 && r.some((n) => n > 0)
}

/** Stack cap: first two radii on free outer end; other corners stay square.
 * Caller only passes an active radius (length ≥ 1, some n > 0). */
function stackCapRadius(radius: number[], horizontal: boolean): number[] {
  const r0 = radius[0]!
  const r1 = radius.length >= 2 ? radius[1]! : r0
  // Horizontal free outer end (value axis → right): TR, BR.
  // Vertical free outer end (value axis → top): TL, TR.
  return horizontal ? [0, r0, r1, 0] : [r0, r1, 0, 0]
}

function applyBorderRadiusToSeries(result: EChartsOption, radius: number[]): EChartsOption {
  for (const s of result.series as { itemStyle?: BarItemStyle }[]) {
    s.itemStyle = { ...s.itemStyle, borderRadius: radius }
  }
  return result
}

export function useBarChartOptions(config: BaseChartConfig) {
  const { chartData, sort, showLabels, isDark, scale, stack, horizontal, borderRadius } = config

  const sortedData = useSortedSeriesData(chartData, sort)

  const options = computed<EChartsOption>(() => {
    const isHorizontal = horizontal?.value ?? false
    const radius = borderRadius?.value
    const activeRadius = isActiveRadius(radius)

    if (chartData.value.mixedTuples?.length) {
      const result = buildMixedAxes2DOptions(config, 'bar')
      return activeRadius ? applyBorderRadiusToSeries(result, radius) : result
    }
    if (chartData.value.valueTuples?.length) {
      const result = buildValueAxes2DOptions(config, 'bar')
      return activeRadius ? applyBorderRadiusToSeries(result, radius) : result
    }

    const { series, xAxisData, hasYAxis } = sortedData.value
    const baseOptions = getBaseOptions(config)
    const styling = getChartStyling(isDark.value)
    // `scale` is optional on BaseChartConfig (relaxed in Task 7) — pie/heatmap/
    // radar pass a config without it. The bar composable is the only consumer,
    // so we default at the call site.
    const yScale = resolveLogScale(
      scale?.value ?? 'linear',
      series.flatMap((s) => s.values)
    )
    const largeX = isLargeXAxis(xAxisData)
    const xLabel = chartData.value.axisLabels?.x
    const useStack = stack?.value === true && yScale !== 'log'

    if (!hasYAxis && isHorizontal) {
      const seriesItem = {
        name: chartData.value.title,
        type: 'bar' as const,
        data: series.map((s) => barNullable(s.values[0] ?? null, yScale)),
        label: createLabelConfig(showLabels.value, styling, 'horizontal'),
        large: true,
        largeThreshold: LARGE_DATA_THRESHOLD,
        itemStyle: { color: getNextColorFor(chartData.value.title) } as BarItemStyle,
      }
      if (activeRadius) {
        seriesItem.itemStyle = { ...seriesItem.itemStyle, borderRadius: radius }
      }
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
        ...createHorizontalAxisConfig(styling, xAxisData, yScale, xLabel, largeX),
        ...(largeX ? { dataZoom: createHorizontalDataZoomConfig(styling) } : {}),
        series: [seriesItem],
      } as EChartsOption
    }

    if (!hasYAxis) {
      const seriesItem = {
        name: chartData.value.title,
        type: 'bar' as const,
        // Plain values + one series-level label, not a per-point {value,label}
        // object — a 100k-bar chart would otherwise allocate 100k label configs
        // on every recompute. `large` keeps the draw on one frame past the
        // threshold.
        data: series.map((s) => barNullable(s.values[0] ?? null, yScale)),
        label: createLabelConfig(showLabels.value, styling),
        large: true,
        largeThreshold: LARGE_DATA_THRESHOLD,
        itemStyle: { color: getNextColorFor(chartData.value.title) } as BarItemStyle,
      }
      if (activeRadius) {
        seriesItem.itemStyle = { ...seriesItem.itemStyle, borderRadius: radius }
      }
      return {
        ...baseOptions,
        grid: createGridConfig(1, largeX),
        tooltip: createTooltipConfig(false, isDark.value),
        legend: { show: false },
        ...createAxisConfig(styling, xAxisData, yScale, xLabel, largeX),
        ...(largeX ? { dataZoom: createDataZoomConfig(xAxisData, styling) } : {}),
        series: [seriesItem],
      } as EChartsOption
    }

    // Dual categories: transpose — each y-axis value becomes a bar group
    const yAxisLabels = chartData.value.yAxis
    const transposedSeries = yAxisLabels.map((yAxisLabel, yIndex) => ({
      name: yAxisLabel,
      type: 'bar' as const,
      data: series.map((s) => barNullable(s.values[yIndex] ?? null, yScale)),
      label: createLabelConfig(
        showLabels.value,
        styling,
        isHorizontal ? 'horizontal' : 'vertical',
        useStack
      ),
      large: true,
      largeThreshold: LARGE_DATA_THRESHOLD,
      ...(useStack ? { stack: 'total' } : {}),
      itemStyle: { color: getNextColorFor(yAxisLabel) } as BarItemStyle,
    }))

    // Secondary sort when there is only one x-group (sort within the group).
    // data items are now plain numbers (or null). Apply after sort so stacked
    // top-only radius tracks the outermost segment.
    if (sort.value.enabled && xAxisData.length === 1) {
      transposedSeries.sort((a, b) => {
        const valA = a.data[0] ?? 0
        const valB = b.data[0] ?? 0
        return sort.value.order === 'asc' ? valA - valB : valB - valA
      })
    }

    if (activeRadius) {
      // Always set borderRadius on every series so vue-echarts/ECharts merge
      // cannot keep a previous full radius after toggling stack on.
      const cap = useStack ? stackCapRadius(radius, isHorizontal) : radius
      transposedSeries.forEach((seriesItem, index) => {
        const isOuter = !useStack || index === transposedSeries.length - 1
        seriesItem.itemStyle = {
          ...seriesItem.itemStyle,
          borderRadius: isOuter ? cap : [0, 0, 0, 0],
        }
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
        ...createHorizontalAxisConfig(styling, xAxisData, yScale, xLabel, largeX),
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
      ...createAxisConfig(styling, xAxisData, yScale, xLabel, largeX),
      ...(largeX ? { dataZoom: createDataZoomConfig(xAxisData, styling) } : {}),
      series: transposedSeries,
    } as EChartsOption
  })

  return { options }
}
