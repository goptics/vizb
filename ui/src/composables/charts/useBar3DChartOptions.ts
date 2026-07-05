import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { COLOR_PALETTE, getNextColorFor } from '@/lib/utils'
import { getChartStyling, getTooltipTheme } from './shared'
import {
  EMPTY_RENDER,
  makeAxis3DCommon,
  axis3DName,
  create3DTooltipFormatter,
  createZLegendConfig,
  create3DGridConfig,
  create3DCellLabel,
  barSizeFor3DGrid,
  resolve3DVisualMap,
  createValue3DTooltipFormatter,
  buildContinuous3DOptions,
  makeContinuous3DParams,
  valuePoints3DToSeries,
  type Continuous3DContext,
} from './shared'
import { buildMixedAxes3DOptions } from './shared/mixedMode'
import type { Series3DData } from '@/types'

export function useBar3DChartOptions(config: BaseChartConfig) {
  const { chartData, isDark, threeDRotate, visibleZ, showLabels, scale, threeDVisualMap } = config

  const options = computed<EChartsOption>(() => {
    const styling = getChartStyling(isDark.value)
    const base = getBaseOptions(config)
    const render = chartData.value.render3D ?? EMPTY_RENDER
    const { xValues, yValues, zValues } = render
    const useVisualMap = threeDVisualMap?.value === true
    const defaultColor = COLOR_PALETTE[0]!
    const axisCommon = makeAxis3DCommon(styling)
    const zAxis3DBase = {
      ...(scale?.value === 'log'
        ? { type: 'log' as const, logBase: 10 }
        : { type: 'value' as const }),
      ...axisCommon,
    }
    if (render.mode === 'mixed') {
      return buildMixedAxes3DOptions(config, 'bar3D')
    }

    const isValueMode = render.mode === 'value'
    const grid3D = create3DGridConfig({
      styling,
      autoRotate: threeDRotate?.value ?? false,
      xCount: xValues.length,
      yCount: yValues.length,
      mode: isValueMode ? 'value' : 'grouped',
    })
    const barSize = isValueMode
      ? barSizeFor3DGrid(xValues.length, yValues.length, grid3D.boxWidth, grid3D.boxDepth)
      : undefined

    if (isValueMode) {
      const seriesData = render.barSeries
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
        series: seriesData.map((s: Series3DData) => ({
          name: s.name,
          type: 'bar3D' as const,
          bevelSize: 0.3,
          bevelSmoothness: 3,
          ...(barSize ? { barSize } : {}),
          data: s.data,
          ...(useVisualMap ? {} : { itemStyle: { color: defaultColor } }),
          shading: 'lambert',
          label: create3DCellLabel(showLabels.value, cellTotals, styling.textColor),
          emphasis: { label: { show: false } },
        })),
      } as unknown as EChartsOption
    }

    const continuousCtx: Continuous3DContext = {
      base,
      styling,
      isDark: isDark.value,
      showLabels: showLabels.value,
      useVisualMap,
      defaultColor,
      threeDRotate: threeDRotate?.value ?? false,
      scale: scale?.value ?? 'linear',
      axisLabels: chartData.value.axisLabels,
    }

    const valuePoints3D = chartData.value.valuePoints3D
    if (!render.barSeries.length && valuePoints3D?.length) {
      return buildContinuous3DOptions(
        makeContinuous3DParams(
          continuousCtx,
          valuePoints3DToSeries(valuePoints3D, chartData.value.title)
        ),
        'bar3D'
      )
    }

    if (render.mode === 'continuous') {
      return buildContinuous3DOptions(
        makeContinuous3DParams(continuousCtx, render.barSeries),
        'bar3D'
      )
    }

    const points = chartData.value.points ?? []
    const seriesData = render.barSeries
    const sel = visibleZ?.value ?? {}
    const aggPoints = points.filter((p) => sel[p.zAxis] !== false)
    const cellTotals = render.cellTotals ?? {}
    const lastVisibleZName =
      [...zValues].reverse().find((z) => sel[z] !== false) ?? zValues[zValues.length - 1]

    const series = seriesData.map((s: Series3DData) => {
      const isTop = showLabels.value && s.name === lastVisibleZName
      return {
        name: s.name,
        type: 'bar3D' as const,
        stack: 'z',
        bevelSize: 0.3,
        bevelSmoothness: 3,
        data: s.data,
        itemStyle: { color: getNextColorFor(s.name) },
        shading: 'lambert',
        label: create3DCellLabel(isTop, cellTotals, styling.textColor),
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
      zAxis3D: {
        ...zAxis3DBase,
      },
      grid3D,
      series,
    } as unknown as EChartsOption
  })

  return { options }
}
