import type { EChartsOption } from 'echarts'
import type { ScaleType } from '@/types'
import { getDefaultThemeColor, getNextColorFor } from '@/lib/utils'
import { type BaseChartConfig, getBaseOptions } from '../baseChartOptions'
import {
  createAxisConfig,
  createDataZoomConfig,
  createLabelConfig,
  createValueModeGridConfig,
  createLineAxisPointer,
  createShadowAxisPointer,
  formatTooltipValue,
  getChartStyling,
  getTooltipTheme,
  isLargeXAxis,
  LARGE_DATA_THRESHOLD,
  scatterSeriesLargeOpts,
} from './chartConfig'
import {
  makeAxis3DCommon,
  axis3DName,
  create3DGridConfig,
  createMixed3DTooltipFormatter,
  continuous3DGridCounts,
  symbolSizeForContinuous3D,
  resolve3DVisualMap,
} from './3d'
import { adjustForLogScaleLine, getEffectiveScale } from './common'
import { resolve3DSymbolProps, resolveSeriesSymbol } from './seriesConfig'
import { resolve2DScatterVisualMap } from './visualMap'
import type { Series3DData } from '@/types'

const defaultScatterSymbol = { symbol: 'circle' as const, symbolSize: 8 }
const largeScatterSymbol = { symbol: 'circle' as const, symbolSize: 5 }
const defaultLineSymbol = { symbol: 'circle' as const, symbolSize: 7 }
const largeLineSymbol = { symbol: 'none' as const, sampling: 'lttb' as const }

export type Mixed2DChartType = 'scatter' | 'bar' | 'line'
export type Mixed3DChartType = 'scatter3D' | 'bar3D' | 'line3D'

export function scaleMixedTuples(
  tuples: [number, number][],
  scale: ScaleType
): [number, number | null][] {
  if (scale !== 'log') return tuples
  return tuples.map(([x, y]) => [x, adjustForLogScaleLine(y, scale)])
}

type AxisTooltipParam = {
  name?: string
  value?: number | [number, number | null]
  data?: [number, number | null]
  marker?: string
}

const mixedAxisYValue = (p: AxisTooltipParam): number | null | undefined => {
  const raw = p.value
  if (Array.isArray(raw)) return raw[1]
  if (typeof raw === 'number') return raw
  return p.data?.[1]
}

export function createMixedModeTooltip(
  isDark: boolean,
  xCategories: string[],
  chartType: Mixed2DChartType,
  xLabel?: string,
  yLabel?: string
): EChartsOption['tooltip'] {
  const theme = getTooltipTheme(isDark)
  const styling = getChartStyling(isDark)
  const xName = xLabel ?? 'x'
  const yName = yLabel ?? 'y'
  const snapToCategory = chartType === 'line' || chartType === 'scatter'

  return {
    trigger: 'axis',
    ...theme,
    axisPointer: snapToCategory
      ? createLineAxisPointer(styling, theme, true)
      : createShadowAxisPointer(styling, theme),
    formatter: (params: unknown) => {
      const items = (Array.isArray(params) ? params : [params]) as AxisTooltipParam[]
      const p = items[0]
      if (!p) return ''

      const category =
        p.name ??
        (() => {
          const xi = p.data?.[0]
          return xi === undefined ? '' : (xCategories[xi] ?? String(xi))
        })()
      const y = mixedAxisYValue(p)
      const yText = y === null || y === undefined ? '' : formatTooltipValue(y)
      const marker = p.marker ?? ''
      return `<strong>${xName}: ${category}</strong><br/>${marker} ${yName}: <b>${yText}</b>`
    },
  }
}

/** Mixed-axis 2D: category x + value y, one point per row. */
export function buildMixedAxes2DOptions(
  config: BaseChartConfig,
  chartType: Mixed2DChartType = 'scatter'
): EChartsOption {
  const { chartData, showLabels, isDark, scale, visualMap } = config
  const tuples = chartData.value.mixedTuples ?? []
  const xCategories = chartData.value.xCategories ?? []
  const xLabel = chartData.value.axisLabels?.x
  const yLabel = chartData.value.axisLabels?.y
  const baseOptions = getBaseOptions(config)
  const styling = getChartStyling(isDark.value)
  const effectiveScale = scale?.value ?? 'linear'
  const { minValue, effectiveScale: yScale } = getEffectiveScale(
    [{ xAxis: chartData.value.title, values: tuples.map(([, y]) => y), benchmarkId: '' }],
    effectiveScale
  )

  const data = scaleMixedTuples(tuples, yScale)
  const largeX = isLargeXAxis(xCategories)
  const useVisualMap = chartType === 'scatter' && visualMap?.value === true
  const smoothLines = chartType === 'line' && config.smooth?.value === true
  const colorValues = tuples.map(([, y]) => y).filter((v) => isFinite(v))
  const color = getNextColorFor(chartData.value.title)

  const label = {
    ...createLabelConfig(showLabels.value, styling),
    formatter: (p: { data: [number, number | null] }) => {
      const y = p.data[1]
      return y === null || y === undefined ? '' : String(Math.round(y * 100) / 100)
    },
  }

  const seriesCommon = {
    name: chartData.value.title,
    data,
    label,
    ...(useVisualMap ? {} : { itemStyle: { color } }),
  }

  const series =
    chartType === 'bar'
      ? {
          ...seriesCommon,
          type: 'bar' as const,
        }
      : chartType === 'line'
        ? {
            ...seriesCommon,
            type: 'line' as const,
            large: true,
            largeThreshold: LARGE_DATA_THRESHOLD,
            ...(smoothLines ? { smooth: true } : {}),
            ...resolveSeriesSymbol(
              largeX ? largeLineSymbol : defaultLineSymbol,
              config.symbol?.value,
              config.symbolSize?.value
            ),
          }
        : {
            ...seriesCommon,
            type: 'scatter' as const,
            ...scatterSeriesLargeOpts(useVisualMap),
            ...resolveSeriesSymbol(
              largeX ? largeScatterSymbol : defaultScatterSymbol,
              config.symbol?.value,
              config.symbolSize?.value
            ),
          }

  const axisConfig = createAxisConfig(
    styling,
    xCategories,
    yScale,
    minValue,
    xLabel,
    largeX,
    chartType === 'scatter'
  )

  return {
    ...baseOptions,
    legend: { show: false },
    grid: createValueModeGridConfig(largeX),
    visualMap: resolve2DScatterVisualMap(useVisualMap, colorValues, styling, 1),
    tooltip: createMixedModeTooltip(isDark.value, xCategories, chartType, xLabel, yLabel),
    ...axisConfig,
    dataZoom: largeX
      ? createDataZoomConfig(xCategories, styling)
      : [
          { type: 'inside', xAxisIndex: 0 },
          { type: 'inside', yAxisIndex: 0 },
        ],
    series: [series],
  } as EChartsOption
}

/** Mixed-axis 3D: category x + value y + value z, one point per row. */
export function buildMixedAxes3DOptions(
  config: BaseChartConfig,
  chartType: Mixed3DChartType = 'scatter3D'
): EChartsOption {
  const { chartData, isDark, threeDRotate, scale, threeDVisualMap, symbol, symbolSize } = config
  const styling = getChartStyling(isDark.value)
  const base = getBaseOptions(config)
  const render = chartData.value.render3D!
  const { xValues } = render
  const axisLabels = chartData.value.axisLabels
  const useVisualMap = threeDVisualMap?.value === true
  const defaultColor = getDefaultThemeColor()
  const axisCommon = makeAxis3DCommon(styling)
  const symbolOverride = symbol?.value
  const symbolSizeOverride = symbolSize?.value

  const seriesData = render.lineSeries
  const pointCount = seriesData[0]?.data.length ?? 0
  const { yCount } = continuous3DGridCounts(pointCount)
  const yScale = scale?.value ?? 'linear'
  const valueType = yScale === 'log' ? ('log' as const) : ('value' as const)
  const logOpts = yScale === 'log' ? { logBase: 10 } : {}
  const grid3D = create3DGridConfig({
    styling,
    autoRotate: threeDRotate?.value ?? false,
    orthographic: true,
    xCount: xValues.length,
    yCount,
    mode: 'mixed',
  })
  const mixedSymbolSize = symbolSizeForContinuous3D(pointCount, grid3D.boxWidth, grid3D.boxDepth)

  const series = seriesData.map((s: Series3DData) => {
    if (chartType === 'bar3D') {
      return {
        name: s.name,
        type: 'bar3D' as const,
        bevelSize: 0.3,
        bevelSmoothness: 3,
        data: s.data,
        ...(useVisualMap ? {} : { itemStyle: { color: defaultColor } }),
        shading: 'lambert',
        label: { show: false },
        emphasis: { label: { show: false } },
      }
    }
    if (chartType === 'line3D') {
      return {
        name: s.name,
        type: 'line3D' as const,
        lineStyle: { width: 3, color: defaultColor },
        data: s.data,
        itemStyle: { color: defaultColor },
        shading: 'lambert',
        label: { show: false },
        emphasis: { label: { show: false } },
      }
    }
    return {
      name: s.name,
      type: 'scatter3D' as const,
      data: s.data,
      ...resolve3DSymbolProps(mixedSymbolSize, symbolOverride, symbolSizeOverride),
      ...(useVisualMap ? {} : { itemStyle: { color: defaultColor } }),
      label: { show: false },
      emphasis: { label: { show: false } },
    }
  })

  return {
    ...base,
    legend: { show: false },
    visualMap: resolve3DVisualMap(useVisualMap, seriesData, styling),
    tooltip: {
      ...base.tooltip,
      ...getTooltipTheme(isDark.value),
      formatter: createMixed3DTooltipFormatter({
        xValues,
        isDark: isDark.value,
        xAxisLabel: axisLabels?.x,
        yAxisLabel: axisLabels?.y,
        zAxisLabel: axisLabels?.z,
      }),
    },
    xAxis3D: {
      type: 'category',
      data: xValues,
      ...axisCommon,
      ...axis3DName(axisLabels?.x, styling),
    },
    yAxis3D: {
      type: valueType,
      ...logOpts,
      ...axisCommon,
      ...axis3DName(axisLabels?.y, styling),
    },
    zAxis3D: {
      type: valueType,
      ...logOpts,
      ...axisCommon,
      ...axis3DName(axisLabels?.z, styling),
    },
    grid3D,
    series,
  } as unknown as EChartsOption
}
