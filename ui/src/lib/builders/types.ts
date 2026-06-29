import type { ChartData, DataPoint, AxisLabels, Sort, ScaleType, Stat, Axis } from '@/types'
import type { CanonicalAxisOrders } from '../transform'

export interface BuildContext {
  signature: string
  statTemplate: Stat
  labels?: AxisLabels
  sort: Sort
  showLabels: boolean
  scale: ScaleType
  canonical?: CanonicalAxisOrders
  threeD: boolean
  preserveRows: boolean
  // Value/mixed-mode extras (ignored by grouped/preserveRows builders).
  axes?: Axis[]
  identityString?: string
  targetString?: string
}

export interface ChartBuilder {
  /** Build the ChartData for this chart shape from the raw data points. */
  build(data: DataPoint[], ctx: BuildContext): ChartData
  /** Whether the chart has any plottable data. */
  plottable(chart: ChartData): boolean
  /** Cardinality for an axis badge. */
  badgeCount(chart: ChartData, axis: 'x' | 'y' | 'z'): number
  /** Sum of every plotted metric value. */
  grandTotal(chart: ChartData, visibleZ?: Record<string, boolean>): number
  /** Whether this chart should render as 3D. */
  is3D(chart: ChartData, cfg?: { threeD?: boolean }, axes?: Axis[]): boolean
}

export const builderStatType = (chart: ChartData): string => chart.statType ?? 'grouped'
