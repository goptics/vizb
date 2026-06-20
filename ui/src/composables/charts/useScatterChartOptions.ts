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
  buildScatterAxes2DOptions,
} from './shared'
import {
  adjustForLogScaleLine,
  useSortedSeriesData,
  getEffectiveScale,
  computeSeriesTotals,
} from './shared/common'

const defaultSymbol = { symbol: 'circle' as const, symbolSize: 8 }
const largeSymbol = { symbol: 'circle' as const, symbolSize: 5 }

export function useScatterChartOptions(config: BaseChartConfig) {
  const { chartData, sort, isDark, showLabels, scale } = config

  const sortedData = useSortedSeriesData(chartData, sort)

  const options = computed<EChartsOption>(() => {
    if (chartData.value.valueTuples?.length) {
      return buildScatterAxes2DOptions(config)
    }

    const { series, xAxisData, hasYAxis } = sortedData.value
    const baseOptions = getBaseOptions(config)
    const styling = getChartStyling(isDark.value)
    const { minValue, effectiveScale } = getEffectiveScale(series, scale?.value ?? 'linear')
    const largeX = isLargeXAxis(xAxisData)
    const xLabel = chartData.value.axisLabels?.x

    if (!hasYAxis) {
      return {
        ...baseOptions,
        grid: createGridConfig(1, largeX),
        tooltip: createPinnedAxisTooltip(isDark.value),
        ...createAxisConfig(styling, xAxisData, effectiveScale, minValue, xLabel, largeX),
        ...(largeX ? { dataZoom: createDataZoomConfig(xAxisData, styling) } : {}),
        legend: { show: false },
        series: [
          {
            name: chartData.value.title,
            type: 'scatter' as const,
            data: series.map((s) => adjustForLogScaleLine(s.values[0] ?? null, effectiveScale)),
            label: createLabelConfig(showLabels.value, styling),
            large: true,
            largeThreshold: LARGE_DATA_THRESHOLD,
            itemStyle: { color: getNextColorFor(chartData.value.title) },
            ...(largeX ? largeSymbol : defaultSymbol),
          },
        ],
      } as EChartsOption
    }

    const yAxisLabels = chartData.value.yAxis
    const transposedSeries = yAxisLabels.map((yAxisLabel, yIndex) => ({
      name: yAxisLabel,
      type: 'scatter' as const,
      data: series.map((s) => adjustForLogScaleLine(s.values[yIndex] ?? null, effectiveScale)),
      label: createLabelConfig(showLabels.value, styling),
      large: true,
      largeThreshold: LARGE_DATA_THRESHOLD,
      itemStyle: { color: getNextColorFor(yAxisLabel) },
      ...(largeX ? largeSymbol : defaultSymbol),
    }))

    const seriesTotals = computeSeriesTotals(transposedSeries)
    const yLabel = chartData.value.axisLabels?.y

    return {
      ...baseOptions,
      ...(yLabel ? { title: makeLegendTitle(yLabel, styling) } : {}),
      grid: createGridConfig(transposedSeries.length, largeX),
      tooltip: createTooltipConfig(hasXAxis(chartData), isDark.value, seriesTotals),
      ...createAxisConfig(styling, xAxisData, effectiveScale, minValue, xLabel, largeX),
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
