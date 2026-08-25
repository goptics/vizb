import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import type { BarBackground } from '@/types'
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
  createValueAxisConfig,
  createValueModeTooltip,
  getChartStyling,
  horizontalLegendBottom,
  isLargeXAxis,
  makeLegendTitle,
  LARGE_DATA_THRESHOLD,
} from './shared/chartConfig'
import { useSortedSeriesData, resolveLogScale, computeSeriesTotals } from './shared/common'
import {
  axisIsLog,
  axisLogBase,
  DEFAULT_LOG_AXES,
  numericLogXValues,
  parseScale,
} from '@/lib/scale'
import { buildValueAxes2DOptions } from './shared/valueMode'
import { buildMixedAxes2DOptions } from './shared/mixedMode'

const barNullable = (val: number | null, yLog: boolean): number | null =>
  val === null ? null : yLog && val <= 0 ? null : val

const asLogXPairs = (
  xNums: number[],
  values: (number | null)[],
  yLog: boolean
): [number, number | null][] => xNums.map((x, i) => [x, barNullable(values[i] ?? null, yLog)])

// ECharts itemStyle.borderRadius: [TL, TR, BR, BL] (len 1–4; [8] = all corners).
type BarItemStyle = { color?: string; borderRadius?: number[] }

// Style payload for ECharts series.backgroundStyle — the wire background minus
// its `active` on-switch.
type BarBackgroundStyle = Omit<BarBackground, 'active'>

// Loose bar-series shape mutated by the background post-pass. `itemStyle` is
// part of the shape only so the series literals (which always carry it) stay
// assignable to this otherwise all-optional weak type.
type BarSeriesLike = {
  itemStyle?: BarItemStyle
  showBackground?: boolean
  backgroundStyle?: BarBackgroundStyle
}

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

/** Style keys for ECharts `backgroundStyle`: every defined key of the wire
 * background, with the `active` on-switch stripped. Unset keys stay absent so
 * ECharts applies its own defaults (no invented defaults in the UI). */
function backgroundStyleOf(background: BarBackground): BarBackgroundStyle {
  const style = { ...background }
  delete style.active
  return Object.fromEntries(
    Object.entries(style).filter(([, value]) => value !== undefined)
  ) as BarBackgroundStyle
}

/** Stamp the wire background onto every bar series. Only `active: true` writes
 * ECharts keys; `active: false` / missing `background` leave `showBackground`
 * and `backgroundStyle` unset so this renderer does not invent defaults.
 * ChartCard assigns a new option reference on each update, so vue-echarts
 * uses `notMerge: true` and omitted keys cannot stick from a previous chart. */
function applyBackgroundToSeries(
  seriesList: BarSeriesLike[],
  backgroundStyle: BarBackgroundStyle | undefined
): void {
  if (!backgroundStyle) return
  for (const series of seriesList) {
    series.showBackground = true
    series.backgroundStyle = backgroundStyle
  }
}

export function useBarChartOptions(config: BaseChartConfig) {
  const {
    chartData,
    sort,
    showLabels,
    isDark,
    scale,
    stack,
    horizontal,
    borderRadius,
    background,
  } = config

  const sortedData = useSortedSeriesData(chartData, sort)

  const options = computed<EChartsOption>(() => {
    const isHorizontal = horizontal?.value ?? false
    const radius = borderRadius?.value
    const activeRadius = isActiveRadius(radius)
    const bg = background?.value
    const activeBackground = bg?.active === true
    const backgroundStyle = activeBackground ? backgroundStyleOf(bg!) : undefined

    if (chartData.value.mixedTuples?.length) {
      const result = buildMixedAxes2DOptions(config, 'bar')
      const withRadius = activeRadius ? applyBorderRadiusToSeries(result, radius) : result
      applyBackgroundToSeries(withRadius.series as BarSeriesLike[], backgroundStyle)
      return withRadius
    }
    if (chartData.value.valueTuples?.length) {
      const result = buildValueAxes2DOptions(config, 'bar')
      const withRadius = activeRadius ? applyBorderRadiusToSeries(result, radius) : result
      applyBackgroundToSeries(withRadius.series as BarSeriesLike[], backgroundStyle)
      return withRadius
    }

    const { series, xAxisData, hasYAxis } = sortedData.value
    const baseOptions = getBaseOptions(config)
    const styling = getChartStyling(isDark.value)
    // `scale` is optional on BaseChartConfig (relaxed in Task 7) — pie/heatmap/
    // radar pass a config without it. The bar composable is the only consumer,
    // so we default at the call site.
    const parsed = parseScale(scale?.value)
    const defaultAxes = isHorizontal ? DEFAULT_LOG_AXES.horizontalBar : DEFAULT_LOG_AXES.grouped2d
    const xWant = axisIsLog(parsed, 'x', defaultAxes)
    const yWant = axisIsLog(parsed, 'y', defaultAxes)
    const valueLogWant = isHorizontal ? xWant : yWant
    const yScale = resolveLogScale(
      valueLogWant ? 'log' : 'linear',
      series.flatMap((s) => s.values)
    )
    const valueLog = yScale === 'log'
    const xNums = !isHorizontal ? numericLogXValues(xAxisData, xWant) : null
    const largeX = isLargeXAxis(xAxisData)
    const xLabel = chartData.value.axisLabels?.x
    const yLogBase = axisLogBase(parsed, isHorizontal ? 'x' : 'y')
    const groupedAxes = xNums
      ? createValueAxisConfig(styling, xLabel, undefined, yScale, false, {
          xScale: 'log',
          xLogBase: axisLogBase(parsed, 'x'),
          yLogBase,
        })
      : isHorizontal
        ? createHorizontalAxisConfig(styling, xAxisData, yScale, xLabel, largeX, yLogBase)
        : createAxisConfig(styling, xAxisData, yScale, xLabel, largeX, false, yLogBase)
    const groupedDataZoom = xNums
      ? [
          { type: 'inside' as const, xAxisIndex: 0 },
          { type: 'inside' as const, yAxisIndex: 0 },
        ]
      : largeX
        ? isHorizontal
          ? createHorizontalDataZoomConfig(styling)
          : createDataZoomConfig(xAxisData, styling)
        : undefined
    const useStack = stack?.value === true && yScale !== 'log'

    if (!hasYAxis && isHorizontal) {
      const seriesItem = {
        name: chartData.value.title,
        type: 'bar' as const,
        data: xNums
          ? asLogXPairs(
              xNums,
              series.map((s) => s.values[0] ?? null),
              valueLog
            )
          : series.map((s) => barNullable(s.values[0] ?? null, valueLog)),
        label: createLabelConfig(showLabels.value, styling, 'horizontal'),
        large: true,
        largeThreshold: LARGE_DATA_THRESHOLD,
        itemStyle: { color: getNextColorFor(chartData.value.title) } as BarItemStyle,
      }
      if (activeRadius) {
        seriesItem.itemStyle = { ...seriesItem.itemStyle, borderRadius: radius }
      }
      applyBackgroundToSeries([seriesItem], backgroundStyle)
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
        ...groupedAxes,
        ...(groupedDataZoom ? { dataZoom: groupedDataZoom } : {}),
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
        data: xNums
          ? asLogXPairs(
              xNums,
              series.map((s) => s.values[0] ?? null),
              valueLog
            )
          : series.map((s) => barNullable(s.values[0] ?? null, valueLog)),
        label: createLabelConfig(showLabels.value, styling),
        large: true,
        largeThreshold: LARGE_DATA_THRESHOLD,
        itemStyle: { color: getNextColorFor(chartData.value.title) } as BarItemStyle,
      }
      if (activeRadius) {
        seriesItem.itemStyle = { ...seriesItem.itemStyle, borderRadius: radius }
      }
      applyBackgroundToSeries([seriesItem], backgroundStyle)
      return {
        ...baseOptions,
        grid: createGridConfig(1, largeX),
        tooltip: xNums
          ? createValueModeTooltip(isDark.value, xLabel, chartData.value.axisLabels?.y, true)
          : createTooltipConfig(false, isDark.value),
        legend: { show: false },
        ...groupedAxes,
        ...(groupedDataZoom ? { dataZoom: groupedDataZoom } : {}),
        series: [seriesItem],
      } as EChartsOption
    }

    // Dual categories: transpose — each y-axis value becomes a bar group
    const yAxisLabels = chartData.value.yAxis
    const transposedSeries = yAxisLabels.map((yAxisLabel, yIndex) => ({
      name: yAxisLabel,
      type: 'bar' as const,
      data: xNums
        ? asLogXPairs(
            xNums,
            series.map((s) => s.values[yIndex] ?? null),
            valueLog
          )
        : series.map((s) => barNullable(s.values[yIndex] ?? null, valueLog)),
      label: createLabelConfig(
        showLabels.value,
        styling,
        isHorizontal ? 'horizontal' : 'vertical',
        useStack
      ),
      large: true,
      largeThreshold: LARGE_DATA_THRESHOLD,
      stack: useStack ? 'total' : null,
      itemStyle: { color: getNextColorFor(yAxisLabel) } as BarItemStyle,
    }))

    // Secondary sort when there is only one x-group (sort within the group).
    // data items are now plain numbers (or null). Apply after sort so stacked
    // top-only radius tracks the outermost segment.
    if (sort.value.enabled && xAxisData.length === 1) {
      const metricAt = (d: number | null | [number, number | null] | undefined): number => {
        if (Array.isArray(d)) return d[1] ?? 0
        return d ?? 0
      }
      transposedSeries.sort((a, b) => {
        const valA = metricAt(a.data[0])
        const valB = metricAt(b.data[0])
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

    // Unlike the stack-cap radius, the background applies uniformly to every
    // grouped/stacked segment (it spans the whole category column).
    applyBackgroundToSeries(transposedSeries, backgroundStyle)

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
        ...groupedAxes,
        ...(groupedDataZoom ? { dataZoom: groupedDataZoom } : {}),
        series: transposedSeries,
      } as EChartsOption
    }

    return {
      ...baseOptions,
      ...(showLegendTitle ? { title: makeLegendTitle(yLabel!, styling) } : {}),
      grid: createGridConfig(transposedSeries.length, largeX),
      tooltip: xNums
        ? hasMultipleSeries
          ? createTooltipConfig(hasXAxis(chartData), isDark.value, seriesTotals, 'line')
          : createValueModeTooltip(isDark.value, xLabel, yLabel, true)
        : createTooltipConfig(hasXAxis(chartData), isDark.value, seriesTotals),
      legend: createLegendConfig(
        transposedSeries.map((s) => ({ xAxis: s.name })),
        styling,
        hasMultipleSeries,
        showLegendTitle ? { top: 24 } : undefined
      ),
      ...groupedAxes,
      ...(groupedDataZoom ? { dataZoom: groupedDataZoom } : {}),
      series: transposedSeries,
    } as EChartsOption
  })

  return { options }
}
