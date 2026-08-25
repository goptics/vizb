import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { getDefaultThemeColor, getNextColorFor } from '@/lib/utils'
import {
  EMPTY_RENDER,
  makeAxis3DCommon,
  axis3DName,
  axis3DNameInvisible,
  create3DTooltipFormatter,
  createZLegendConfig,
  create3DGridConfig,
  create3DCellLabel,
  barSizeFor3DGrid,
  symbolSizeFor3DGrid,
  resolve3DVisualMap,
  createValue3DTooltipFormatter,
  buildContinuous3DOptions,
  makeContinuous3DParams,
  valuePoints3DToSeries,
  resolve3DZAxisType,
  zAxisLogBase,
  type Continuous3DContext,
} from './shared/3d'
import { getChartStyling, getTooltipTheme } from './shared/chartConfig'
import { resolve3DSymbolProps } from './shared/seriesConfig'
import { buildMixedAxes3DOptions } from './shared/mixedMode'
import type { Series3DData } from '@/types'

export type Chart3DKind = 'bar3D' | 'line3D' | 'scatter3D'

export function use3DChartOptions(config: BaseChartConfig, kind: Chart3DKind) {
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
    const useVisualMap =
      kind === 'scatter3D'
        ? threeDVisualMap?.value === true || visualMap?.value === true
        : threeDVisualMap?.value === true
    const defaultColor = getDefaultThemeColor()
    const axisCommon = makeAxis3DCommon(styling)
    const seriesData = kind === 'bar3D' ? render.barSeries : render.lineSeries
    const zType = resolve3DZAxisType(scale?.value ?? 'linear', seriesData)
    const zAxis3DBase = {
      type: zType,
      ...(zType === 'log' ? { logBase: zAxisLogBase(scale?.value ?? 'linear') } : {}),
      ...axisCommon,
    }
    if (render.mode === 'mixed') {
      return buildMixedAxes3DOptions(config, kind)
    }

    const isValueMode = render.mode === 'value'
    const grid3D = create3DGridConfig({
      styling,
      autoRotate: threeDRotate?.value ?? false,
      ...(kind !== 'bar3D' ? { orthographic: true } : {}),
      xCount: xValues.length,
      yCount: yValues.length,
      mode: isValueMode ? 'value' : 'grouped',
    })
    const barSize =
      isValueMode && kind === 'bar3D'
        ? barSizeFor3DGrid(xValues.length, yValues.length, grid3D.boxWidth, grid3D.boxDepth)
        : undefined
    const valueSymbolSize =
      isValueMode && kind !== 'bar3D'
        ? symbolSizeFor3DGrid(xValues.length, yValues.length, grid3D.boxWidth, grid3D.boxDepth)
        : undefined
    const groupedSymbolSize = 10
    const axisLabels = chartData.value.axisLabels
    const symbolOverride = symbol?.value
    const symbolSizeOverride = symbolSize?.value

    if (isValueMode) {
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
        series: buildValue3DSeries({
          kind,
          seriesData,
          defaultColor,
          useVisualMap,
          showLabels: showLabels.value,
          cellTotals,
          textColor: styling.textColor,
          barSize,
          valueSymbolSize,
          symbolOverride,
          symbolSizeOverride,
        }),
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
      axisLabels,
      ...(kind !== 'bar3D' ? { symbol: symbolOverride, symbolSize: symbolSizeOverride } : {}),
    }

    const valuePoints3D = chartData.value.valuePoints3D
    if (!seriesData.length && valuePoints3D?.length) {
      return buildContinuous3DOptions(
        makeContinuous3DParams(
          continuousCtx,
          valuePoints3DToSeries(valuePoints3D, chartData.value.title)
        ),
        kind
      )
    }

    if (render.mode === 'continuous') {
      return buildContinuous3DOptions(makeContinuous3DParams(continuousCtx, seriesData), kind)
    }

    const points = chartData.value.points ?? []
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
      // Invisible name keeps nameGap framing; explicit name clears sticky merges.
      zAxis3D: {
        ...zAxis3DBase,
        ...axis3DNameInvisible(axisLabels?.z),
      },
      grid3D,
      series: buildGrouped3DSeries({
        kind,
        seriesData,
        showLabels: showLabels.value,
        lastVisibleZName,
        cellTotals,
        textColor: styling.textColor,
        groupedSymbolSize,
        symbolOverride,
        symbolSizeOverride,
      }),
    } as unknown as EChartsOption
  })

  return { options }
}

function buildValue3DSeries(params: {
  kind: Chart3DKind
  seriesData: Series3DData[]
  defaultColor: string
  useVisualMap: boolean
  showLabels: boolean
  cellTotals: Record<string, number>
  textColor: string
  barSize?: [number, number]
  valueSymbolSize?: number
  symbolOverride?: string
  symbolSizeOverride?: number
}) {
  const {
    kind,
    seriesData,
    defaultColor,
    useVisualMap,
    showLabels,
    cellTotals,
    textColor,
    barSize,
    valueSymbolSize,
    symbolOverride,
    symbolSizeOverride,
  } = params

  if (kind === 'bar3D') {
    return seriesData.map((s: Series3DData) => ({
      name: s.name,
      type: 'bar3D' as const,
      bevelSize: 0.3,
      bevelSmoothness: 3,
      barSize,
      data: s.data,
      ...(useVisualMap ? {} : { itemStyle: { color: defaultColor } }),
      shading: 'lambert',
      label: create3DCellLabel(showLabels, cellTotals, textColor),
      emphasis: { label: { show: false } },
    }))
  }

  if (kind === 'line3D') {
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
      label: create3DCellLabel(showLabels, cellTotals, textColor),
      emphasis: { label: { show: false } },
    }))
    return [...lineSeries, ...labelSeries]
  }

  return seriesData.map((s: Series3DData) => ({
    name: s.name,
    type: 'scatter3D' as const,
    data: s.data,
    ...resolve3DSymbolProps(valueSymbolSize, symbolOverride, symbolSizeOverride),
    ...(useVisualMap ? {} : { itemStyle: { color: defaultColor } }),
    label: create3DCellLabel(showLabels, cellTotals, textColor),
    emphasis: { label: { show: false } },
  }))
}

function buildGrouped3DSeries(params: {
  kind: Chart3DKind
  seriesData: Series3DData[]
  showLabels: boolean
  lastVisibleZName: string | undefined
  cellTotals: Record<string, number>
  textColor: string
  groupedSymbolSize: number
  symbolOverride?: string
  symbolSizeOverride?: number
}) {
  const {
    kind,
    seriesData,
    showLabels,
    lastVisibleZName,
    cellTotals,
    textColor,
    groupedSymbolSize,
    symbolOverride,
    symbolSizeOverride,
  } = params

  if (kind === 'bar3D') {
    return seriesData.map((s: Series3DData) => {
      const isTop = showLabels && s.name === lastVisibleZName
      return {
        name: s.name,
        type: 'bar3D' as const,
        stack: 'z',
        bevelSize: 0.3,
        bevelSmoothness: 3,
        data: s.data,
        itemStyle: { color: getNextColorFor(s.name) },
        shading: 'lambert',
        label: create3DCellLabel(isTop, cellTotals, textColor),
        emphasis: { label: { show: false } },
      }
    })
  }

  if (kind === 'line3D') {
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
    const labelSeries = seriesData.map((s: Series3DData) => {
      const color = getNextColorFor(s.name)
      return {
        name: s.name,
        type: 'scatter3D',
        data: s.data,
        ...resolve3DSymbolProps(groupedSymbolSize, symbolOverride, symbolSizeOverride),
        itemStyle: { color },
        label: create3DCellLabel(showLabels && s.name === lastVisibleZName, cellTotals, textColor),
        emphasis: { label: { show: false } },
      }
    })
    return [...series, ...labelSeries]
  }

  return seriesData.map((s: Series3DData) => {
    const color = getNextColorFor(s.name)
    return {
      name: s.name,
      type: 'scatter3D' as const,
      data: s.data,
      ...resolve3DSymbolProps(groupedSymbolSize, symbolOverride, symbolSizeOverride),
      itemStyle: { color },
      label: create3DCellLabel(showLabels && s.name === lastVisibleZName, cellTotals, textColor),
      emphasis: { label: { show: false } },
    }
  })
}
