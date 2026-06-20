import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'
import type { ChartType, Meta, ChartData, DataPoint } from '../types'
import type { Ref } from 'vue'

/**
 * Utility function to merge Tailwind CSS classes
 * Used for conditional styling with shadcn components
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export const isValidIndex = (id: number, length: number): boolean => id >= 0 && id < length

export const COLOR_PALETTE = [
  '#5470C6', // Blue
  '#3BA272', // Green
  '#FC8452', // Orange
  '#73C0DE', // Light blue
  '#EE6666', // Red
  '#FAC858', // Yellow
  '#9A60B4', // Purple
  '#EA7CCC', // Pink
  '#91CC75', // Lime
  '#FF9F7F', // Coral
]

const colorMap = new Map<string, number>()
let i = 0

export function getNextColorFor(key: string) {
  if (colorMap.has(key)) {
    return COLOR_PALETTE[colorMap.get(key)!]
  }

  const colorIndex = i % COLOR_PALETTE.length
  const color = COLOR_PALETTE[colorIndex]
  colorMap.set(key, colorIndex)

  if (i === COLOR_PALETTE.length) {
    i = 0
  } else {
    i++
  }

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

export const is3D = (chartData: Ref<ChartData, ChartData>, threeD?: boolean) => {
  const chart = chartData.value
  return isGrouped3D(chart) || (isValue3DEligible(chart) && threeD === true)
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

/** 3D engine is present in the HTML bundle (z in raw data, or --3d was baked). */
export const datasetHas3DEngine = (
  data: DataPoint[] | undefined,
  cfg?: { threeD?: boolean }
): boolean => datasetDimension(data) === '3D' || cfg?.threeD !== undefined

/** Value-mode 3D toggle should appear in settings. */
export const canOfferValue3D = (
  chartType: ChartType,
  data: DataPoint[] | undefined,
  hasZOnChart: boolean,
  cfg?: { threeD?: boolean }
): boolean =>
  (chartType === 'bar' || chartType === 'line') &&
  datasetHasBothXY(data) &&
  !hasZOnChart &&
  datasetHas3DEngine(data, cfg)

/** Round to 2 decimals — matches tooltip number formatting. */
export const formatChartTotal = (value: number) => String(Math.round(value * 100) / 100)

/**
 * Sum every plotted metric value in the chart. Works for 1D (x only), 2D (x×y),
 * grouped 3D (x×y×z), and value-mode 3D. When a z legend is active, hidden z
 * series are excluded — same contract as grouped 3D tooltips.
 */
export const computeChartGrandTotal = (
  chart: ChartData,
  visibleZ?: Record<string, boolean>
): number => {
  if (chart.points.length > 0) {
    const filterZ = chartHasZAxis(chart)
    let total = 0
    for (const pt of chart.points) {
      if (filterZ && pt.zAxis && visibleZ?.[pt.zAxis] === false) continue
      total += pt.value
    }
    return total
  }

  let total = 0
  for (const s of chart.series) {
    for (const v of s.values) {
      if (v != null) total += v
    }
  }
  return total
}

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
