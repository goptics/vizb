import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { getDefaultThemeColor, getNextColorFor } from '@/lib/utils'
import { getChartStyling, getTooltipTheme } from './shared'
import {
  EMPTY_RENDER,
  makeAxis3DCommon,
  axis3DName,
  axis3DNameInvisible,
  create3DTooltipFormatter,
  createZLegendConfig,
  create3DGridConfig,
  create3DCellLabel,
  symbolSizeFor3DGrid,
  resolve3DVisualMap,
  createValue3DTooltipFormatter,
  buildContinuous3DOptions,
  makeContinuous3DParams,
  valuePoints3DToSeries,
  visible3DCellTotals,
  type Continuous3DContext,
} from './shared'
import { resolve3DSymbolProps } from './shared/seriesConfig'
import { buildMixedAxes3DOptions } from './shared/mixedMode'
import type { Series3DData } from '@/types'

export function useLine3DChartOptions(config: BaseChartConfig) {
  const {
    chartData,
    isDark,
    threeDRotate,
    visibleZ,
    showLabels,
    labelMode,
    chartTotal,
    scale,
    threeDVisualMap,
    symbol,
    symbolSize,
  } = config

  const options = computed<EChartsOption>(() => {
    const styling = getChartStyling(isDark.value)
    const base = getBaseOptions(config)
    const render = chartData.value.render3D ?? EMPTY_RENDER
    const { xValues, yValues, zValues } = render
    const useVisualMap = threeDVisualMap?.value === true
    const defaultColor = getDefaultThemeColor()
    const axisCommon = makeAxis3DCommon(styling)
    const zAxis3DBase = {
      ...(scale?.value === 'log'
        ? { type: 'log' as const, logBase: 10 }
        : { type: 'value' as const }),
      ...axisCommon,
    }
    if (render.mode === 'mixed') {
      return buildMixedAxes3DOptions(config, 'line3D')
    }

    const isValueMode = render.mode === 'value'
    const grid3D = create3DGridConfig({
      styling,
      autoRotate: threeDRotate?.value ?? false,
      orthographic: true,
      xCount: xValues.length,
      yCount: yValues.length,
      mode: isValueMode ? 'value' : 'grouped',
    })
    const valueSymbolSize = isValueMode
      ? symbolSizeFor3DGrid(xValues.length, yValues.length, grid3D.boxWidth, grid3D.boxDepth)
      : undefined
    const groupedSymbolSize = 10
    const symbolOverride = symbol?.value
    const symbolSizeOverride = symbolSize?.value

    if (isValueMode) {
      const seriesData = render.lineSeries
      const valueLabel = chartData.value.statUnit
        ? `${chartData.value.title} (${chartData.value.statUnit})`
        : chartData.value.title
      const cellTotals = render.cellTotals ?? {}

      const lineSeries = seriesData.map((s: Series3DData) => ({
        name: s.name,
        type: 'line3D' as const,
        lineStyle: { width: 3, color: defaultColor },
        data: s.data,
        itemStyle: { color: defaultColor },
        shading: 'lambert',
        label: { show: false },
        emphasis: { label: { show: false } },
      }))

      const labelSeries = seriesData.map((s: Series3DData) => ({
        name: s.name,
        type: 'scatter3D',
        data: s.data,
        ...resolve3DSymbolProps(valueSymbolSize, symbolOverride, symbolSizeOverride),
        itemStyle: { color: defaultColor },
        label: create3DCellLabel(
          showLabels.value,
          cellTotals,
          styling.textColor,
          labelMode?.value,
          chartTotal?.value
        ),
        emphasis: { label: { show: false } },
      }))

      return {
        ...base,
        legend: { show: false },
        visualMap: resolve3DVisualMap(useVisualMap, seriesData, styling),
        tooltip: {
          ...base.tooltip,
          ...getTooltipTheme(isDark.value),
          formatter: createValue3DTooltipFormatter({
            xValues,
            yValues,
            seriesData: seriesData[0]?.data ?? [],
            isDark: isDark.value,
            xAxisLabel: chartData.value.axisLabels?.x,
            yAxisLabel: chartData.value.axisLabels?.y,
            valueLabel,
            seriesColor: defaultColor,
          }),
        },
        xAxis3D: {
          type: 'category',
          data: xValues,
          ...axisCommon,
          ...axis3DName(chartData.value.axisLabels?.x, styling),
        },
        yAxis3D: {
          type: 'category',
          data: yValues,
          ...axisCommon,
          ...axis3DName(chartData.value.axisLabels?.y, styling),
        },
        zAxis3D: {
          ...zAxis3DBase,
          ...axis3DName(valueLabel, styling),
        },
        grid3D,
        series: [...lineSeries, ...labelSeries],
      } as unknown as EChartsOption
    }

    const continuousCtx: Continuous3DContext = {
      base,
      styling,
      isDark: isDark.value,
      showLabels: showLabels.value,
      labelMode: labelMode?.value,
      chartTotal: chartTotal?.value,
      useVisualMap,
      defaultColor,
      threeDRotate: threeDRotate?.value ?? false,
      scale: scale?.value ?? 'linear',
      axisLabels: chartData.value.axisLabels,
      symbol: symbolOverride,
      symbolSize: symbolSizeOverride,
    }

    const valuePoints3D = chartData.value.valuePoints3D
    if (!render.lineSeries.length && valuePoints3D?.length) {
      return buildContinuous3DOptions(
        makeContinuous3DParams(
          continuousCtx,
          valuePoints3DToSeries(valuePoints3D, chartData.value.title)
        ),
        'line3D'
      )
    }

    if (render.mode === 'continuous') {
      return buildContinuous3DOptions(
        makeContinuous3DParams(continuousCtx, render.lineSeries),
        'line3D'
      )
    }

    const points = chartData.value.points ?? []
    const seriesData = render.lineSeries
    const sel = visibleZ?.value ?? {}
    const aggPoints = points.filter((p) => sel[p.zAxis] !== false)
    const cellTotals = visible3DCellTotals(render.lineSeries, sel)
    const lastVisibleZName =
      [...zValues].reverse().find((z) => sel[z] !== false) ?? zValues[zValues.length - 1]

    const series = seriesData.map((s: Series3DData) => {
      const color = getNextColorFor(s.name)
      return {
        name: s.name,
        type: 'line3D' as const,
        lineStyle: { width: 3, color },
        data: s.data,
        itemStyle: { color },
        shading: 'lambert',
        label: { show: false },
        emphasis: { label: { show: false } },
      }
    })

    const labelSeries = render.lineSeries.map((s: Series3DData) => {
      const color = getNextColorFor(s.name)
      return {
        name: s.name,
        type: 'scatter3D',
        data: s.data,
        ...resolve3DSymbolProps(groupedSymbolSize, symbolOverride, symbolSizeOverride),
        itemStyle: { color },
        label: create3DCellLabel(
          showLabels.value && s.name === lastVisibleZName,
          cellTotals,
          styling.textColor,
          labelMode?.value,
          chartTotal?.value
        ),
        emphasis: { label: { show: false } },
      }
    })

    const tooltipFormatter = create3DTooltipFormatter({
      xValues,
      yValues,
      zValues,
      aggPoints,
      isDark: isDark.value,
      xAxisLabel: chartData.value.axisLabels?.x,
      yAxisLabel: chartData.value.axisLabels?.y,
      zAxisLabel: chartData.value.axisLabels?.z,
    })

    return {
      ...base,
      legend: {
        ...base.legend,
        ...createZLegendConfig(zValues, styling, sel),
      },
      visualMap: resolve3DVisualMap(useVisualMap, seriesData, styling),
      tooltip: {
        ...base.tooltip,
        ...getTooltipTheme(isDark.value),
        formatter: tooltipFormatter,
      },
      xAxis3D: {
        type: 'category',
        data: xValues,
        ...axisCommon,
        ...axis3DName(chartData.value.axisLabels?.x, styling),
      },
      yAxis3D: {
        type: 'category',
        data: yValues,
        ...axisCommon,
        ...axis3DName(chartData.value.axisLabels?.y, styling),
      },
      // Invisible name keeps nameGap framing; explicit name clears sticky merges.
      zAxis3D: {
        ...zAxis3DBase,
        ...axis3DNameInvisible(chartData.value.axisLabels?.z),
      },
      grid3D,
      series: [...series, ...labelSeries],
    } as unknown as EChartsOption
  })

  return { options }
}
