import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'
import type { ChartType, Meta, ChartData, DataPoint, Axis } from '../types'
import type { Ref } from 'vue'
import { arrangementHasChartZ } from './swap'
import { builderForChart, pickBuilder } from './builders'
import { activePalette, palettePrimary, THEMES } from './themes'

/**
 * Utility function to merge Tailwind CSS classes
 * Used for conditional styling with shadcn components
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export const isValidIndex = (id: number, length: number): boolean => id >= 0 && id < length

export const COLOR_PALETTE = THEMES.default
export const getThemePalette = () => activePalette.value
export const getDefaultThemeColor = () => palettePrimary(activePalette.value)

const colorMap = new Map<string, number>()
let i = 0

export function getNextColorFor(key: string) {
  if (colorMap.has(key)) {
    const palette = activePalette.value
    return palette[colorMap.get(key)! % palette.length]
  }

  const palette = activePalette.value
  const rawIndex = i++
  const color = palette[rawIndex % palette.length]
  colorMap.set(key, rawIndex)

  return color
}

export const resetColor = () => {
  i = 0
  colorMap.clear()
}

export const chartHasXAxis = (chart: ChartData) =>
  chart.series.some((series) => series.xAxis && series.xAxis.trim() !== '')

export const chartHasYAxis = (chart: ChartData) =>
  chart.yAxis && chart.yAxis.length > 0 && chart.yAxis[0] !== ''

export const chartHasZAxis = (chart: ChartData) =>
  chart.zAxis && chart.zAxis.length > 0 && chart.zAxis[0] !== ''

export const hasXAxis = (chartData: Ref<ChartData, ChartData>) => chartHasXAxis(chartData.value)

export const hasYAxis = (chartData: Ref<ChartData, ChartData>) => chartHasYAxis(chartData.value)

export const hasZAxis = (chartData: Ref<ChartData, ChartData>) => chartHasZAxis(chartData.value)

// Grouped 3D: x, y, and z dimensions are all present in the chart data.
export const isGrouped3D = (chart: ChartData) =>
  chartHasXAxis(chart) && chartHasYAxis(chart) && chartHasZAxis(chart)

// Value 3D: x+y only — y categories become depth, metric becomes height.
export const isValue3DEligible = (chart: ChartData) =>
  chartHasXAxis(chart) && chartHasYAxis(chart) && !chartHasZAxis(chart)

export const valueModeHasZAxis = (axes: Axis[] | undefined): boolean =>
  !!axes?.some((a) => a.key === 'z')

export const mixedModeHasZAxis = (axes: Axis[] | undefined): boolean =>
  !!axes?.some((a) => a.key === 'z' && a.type === 'value')

export const VALUE_CHART_TYPES = new Set<ChartType>(['scatter', 'bar', 'line'])

export const isValueChartType = (chartType?: ChartType): boolean =>
  !!chartType && VALUE_CHART_TYPES.has(chartType)

/** Value-mode continuous 3D (swap-driven, not category --3d). */
export const isValueModeContinuous3D = (
  chart: ChartData,
  axes: Axis[] | undefined,
  targetString?: string,
  chartType?: ChartType
): boolean =>
  isValueChartType(chartType) &&
  isValueMode(axes) &&
  valueModeHasZAxis(axes) &&
  !!targetString &&
  arrangementHasChartZ(targetString) &&
  chart.render3D?.mode === 'continuous'

export const is3D = (
  chartData: Ref<ChartData, ChartData>,
  threeD?: boolean,
  targetString?: string,
  axes?: Axis[],
  chartType?: ChartType
) => {
  const chart = chartData.value
  if (chart.statType === 'value') {
    return isValueModeContinuous3D(chart, axes, targetString, chartType)
  }
  if (chart.statType === 'mixed') {
    return (
      chart.render3D?.mode === 'mixed' ||
      (isValueChartType(chartType) && isMixedMode(axes) && mixedModeHasZAxis(axes))
    )
  }
  return builderForChart(chart).is3D(chart, { threeD })
}

// Data-shape dimensionality tag, derived from the raw `DataPoint[]` rows. The
// panel uses this to decide which fields render (e.g. `threeDRotate` is 3D-only).
// Independent of the chart's post-group `ChartData` — purely the source shape.
// `undefined` (empty/unknown data) means "no constraint" so the panel still
// shows every field by default until data arrives.
export type Dimension = '1D' | '2D' | '3D'

export const datasetDimension = (data: DataPoint[] | undefined): Dimension | undefined => {
  if (!data?.length) return undefined
  if (data.some((p) => !!p.zAxis)) return '3D'
  if (data.some((p) => !!p.xAxis && !!p.yAxis)) return '2D'
  return '1D'
}

export const datasetHasBothXY = (data: DataPoint[] | undefined): boolean =>
  !!data?.some((p) => !!p.xAxis && !!p.yAxis)

/** 3D chunk is baked into the HTML bundle (z in raw data, or --3d flag was set). */
export const bundleHas3DChunk = (
  data: DataPoint[] | undefined,
  cfg?: { threeD?: boolean }
): boolean => datasetDimension(data) === '3D' || cfg?.threeD !== undefined

/** Category value-mode 3D toggle (--3d on x+y grouped data). Hidden for --axes value mode. */
export const canOfferValue3D = (
  chartType: ChartType,
  data: DataPoint[] | undefined,
  hasZOnChart: boolean,
  cfg?: { threeD?: boolean },
  axes?: Axis[]
): boolean => {
  const builder = pickBuilder({
    valueMode: isValueMode(axes),
    mixedMode: isMixedMode(axes),
  })
  return builder.canOfferValue3D(chartType, data, hasZOnChart, cfg)
}

/** Round to 2 decimals — matches tooltip number formatting. */
export const formatChartTotal = (value: number) => String(Math.round(value * 100) / 100)

export const isValueModeChart = (chart: ChartData): boolean => chart.statType === 'value'

export const chartHasPlottableData = (chart: ChartData): boolean =>
  chart.series.length > 0 ||
  chart.points.length > 0 ||
  (chart.valueTuples?.length ?? 0) > 0 ||
  (chart.valuePoints3D?.length ?? 0) > 0 ||
  (chart.mixedTuples?.length ?? 0) > 0 ||
  (chart.render3D?.mode === 'mixed' && (chart.render3D.lineSeries[0]?.data.length ?? 0) > 0)

/** Category labels for the x/series dimension across chart shapes. */
export const chartSeriesLabels = (chart: ChartData): string[] => {
  if (chart.series.length) return chart.series.map((s) => s.xAxis)
  if (chart.xCategories?.length) return chart.xCategories
  return [...new Set(chart.points.map((p) => p.xAxis))]
}

/** Cardinality shown on ChartCard axis badges (category count or unique value-mode coords). */
export const chartAxisBadgeCount = (chart: ChartData, axis: 'x' | 'y' | 'z'): number =>
  builderForChart(chart).badgeCount(chart, axis)

/**
 * Sum every plotted metric value in the chart. Works for 1D (x only), 2D (x×y),
 * grouped 3D (x×y×z), and value-mode 3D. When a z legend is active, hidden z
 * series are excluded — same contract as grouped 3D tooltips.
 */
export const computeChartGrandTotal = (
  chart: ChartData,
  visibleZ?: Record<string, boolean>
): number => builderForChart(chart).grandTotal(chart, visibleZ)

export const CPUtoString = (cpu: Meta['cpu']) => {
  if (!cpu) {
    return ''
  }

  if (cpu.name && cpu.cores) {
    return `${cpu.name} (${cpu.cores} cores)`
  }

  if (cpu.name) {
    return cpu.name
  }

  if (cpu.cores) {
    return `${cpu.cores} cores`
  }

  return ''
}

/** All axes are continuous numeric (--axes x,y[,z] value mode). */
export const isValueMode = (axes: Axis[] | undefined): boolean =>
  !!axes?.length && axes.every((a) => a.type === 'value')

/** Category x + value y[,z] (solo --select mixed mode). */
export const isMixedMode = (axes: Axis[] | undefined): boolean =>
  !!axes?.length && axes.some((a) => a.type === 'value') && axes.some((a) => a.type !== 'value')

export const isMixedModeChart = (chart: ChartData): boolean => chart.statType === 'mixed'

/** Scatter datasets routed through value or mixed transform paths. */
export const isScatterTransformMode = (axes: Axis[] | undefined): boolean =>
  isValueMode(axes) || isMixedMode(axes)
