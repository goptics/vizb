import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { COLOR_PALETTE, getNextColorFor } from '@/lib/utils'
import {
  EMPTY_RENDER,
  makeAxis3DCommon,
  axis3DName,
  create3DTooltipFormatter,
  createZLegendConfig,
  create3DGridConfig,
  create3DCellLabel,
  symbolSizeFor3DGrid,
  resolve3DVisualMap,
  createValue3DTooltipFormatter,
  buildContinuous3DOptions,
  getChartStyling,
  getTooltipTheme,
  makeContinuous3DParams,
  valuePoints3DToSeries,
  type Continuous3DContext,
} from './shared'
import { resolve3DSymbolProps } from './shared/seriesConfig'
import { buildMixedAxes3DOptions } from './shared/mixedMode'
import type { Series3DData } from '@/types'

export function useScatter3DChartOptions(config: BaseChartConfig) {
  const {
    chartData,
    isDark,
    threeDRotate,
    visibleZ,
    showLabels,
    scale,
    threeDVisualMap,
    visualMap,
    symbol,
    symbolSize,
  } = config

  const options = computed<EChartsOption>(() => {
    const styling = getChartStyling(isDark.value)
    const base = getBaseOptions(config)
    const render = chartData.value.render3D ?? EMPTY_RENDER
    const { xValues, yValues, zValues } = render
    const useVisualMap = threeDVisualMap?.value === true || visualMap?.value === true
    const defaultColor = COLOR_PALETTE[0]!
    const axisCommon = makeAxis3DCommon(styling)
    const zAxis3DBase = {
      ...(scale?.value === 'log'
        ? { type: 'log' as const, logBase: 10 }
        : { type: 'value' as const }),
      ...axisCommon,
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
    const axisLabels = chartData.value.axisLabels
    const symbolOverride = symbol?.value
    const symbolSizeOverride = symbolSize?.value
    const continuousCtx: Continuous3DContext = {
      base,
      styling,
      isDark: isDark.value,
      showLabels: showLabels.value,
      useVisualMap,
      defaultColor,
      threeDRotate: threeDRotate?.value ?? false,
      scale: scale?.value ?? 'linear',
      axisLabels,
      symbol: symbolOverride,
      symbolSize: symbolSizeOverride,
    }

    const valuePoints3D = chartData.value.valuePoints3D
    if (!render.lineSeries.length && valuePoints3D?.length) {
      return buildContinuous3DOptions(
        makeContinuous3DParams(
          continuousCtx,
          valuePoints3DToSeries(valuePoints3D, chartData.value.title)
        )
      )
    }

    if (render.mode === 'continuous') {
      return buildContinuous3DOptions(makeContinuous3DParams(continuousCtx, render.lineSeries))
    }

    if (render.mode === 'mixed') {
      return buildMixedAxes3DOptions(config, 'scatter3D')
    }

    if (isValueMode) {
      const seriesData = render.lineSeries
      const valueLabel = chartData.value.statUnit
        ? `${chartData.value.title} (${chartData.value.statUnit})`
        : chartData.value.title
      const cellTotals = render.cellTotals ?? {}

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
            xAxisLabel: axisLabels?.x,
            yAxisLabel: axisLabels?.y,
            valueLabel,
            seriesColor: defaultColor,
          }),
        },
        xAxis3D: {
          type: 'category',
          data: xValues,
          ...axisCommon,
          ...axis3DName(axisLabels?.x, styling),
        },
        yAxis3D: {
          type: 'category',
          data: yValues,
          ...axisCommon,
          ...axis3DName(axisLabels?.y, styling),
        },
        zAxis3D: {
          ...zAxis3DBase,
          ...axis3DName(valueLabel, styling),
        },
        grid3D,
        series: seriesData.map((s: Series3DData) => ({
          name: s.name,
          type: 'scatter3D' as const,
          data: s.data,
          ...resolve3DSymbolProps(valueSymbolSize, symbolOverride, symbolSizeOverride),
          ...(useVisualMap ? {} : { itemStyle: { color: defaultColor } }),
          label: create3DCellLabel(showLabels.value, cellTotals, styling.textColor),
          emphasis: { label: { show: false } },
        })),
      } as unknown as EChartsOption
    }

    const points = chartData.value.points ?? []
    const seriesData = render.lineSeries
    const sel = visibleZ?.value ?? {}
    const aggPoints = points.filter((p) => sel[p.zAxis] !== false)
    const cellTotals = render.cellTotals ?? {}
    const lastVisibleZName =
      [...zValues].reverse().find((z) => sel[z] !== false) ?? zValues[zValues.length - 1]

    const tooltipFormatter = create3DTooltipFormatter({
      xValues,
      yValues,
      zValues,
      aggPoints,
      isDark: isDark.value,
      xAxisLabel: axisLabels?.x,
      yAxisLabel: axisLabels?.y,
      zAxisLabel: axisLabels?.z,
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
        ...axis3DName(axisLabels?.x, styling),
      },
      yAxis3D: {
        type: 'category',
        data: yValues,
        ...axisCommon,
        ...axis3DName(axisLabels?.y, styling),
      },
      zAxis3D: {
        ...zAxis3DBase,
      },
      grid3D,
      series: seriesData.map((s: Series3DData) => {
        const color = getNextColorFor(s.name)
        return {
          name: s.name,
          type: 'scatter3D' as const,
          data: s.data,
          ...resolve3DSymbolProps(groupedSymbolSize, symbolOverride, symbolSizeOverride),
          itemStyle: { color },
          label: create3DCellLabel(
            showLabels.value && s.name === lastVisibleZName,
            cellTotals,
            styling.textColor
          ),
          emphasis: { label: { show: false } },
        }
      }),
    } as unknown as EChartsOption
  })

  return { options }
}
