import type {
  ChartData,
  DataPoint,
  AxisLabels,
  Sort,
  ScaleType,
  Stat,
  Axis,
  ChartType,
} from '@/types'
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
  /** Cardinality for an axis badge. */
  badgeCount(chart: ChartData, axis: 'x' | 'y' | 'z'): number
  /** Sum of every plotted metric value. */
  grandTotal(chart: ChartData, visibleZ?: Record<string, boolean>): number
  /** Whether this chart should render as 3D. */
  is3D(chart: ChartData, cfg?: { threeD?: boolean }, axes?: Axis[]): boolean
  /** Whether the category --3d toggle can be offered for this shape. */
  canOfferValue3D(
    chartType: ChartType,
    data: DataPoint[] | undefined,
    hasZOnChart: boolean,
    cfg?: { threeD?: boolean }
  ): boolean
}

/** Grouped/preserveRows builders also materialise ChartData from raw points. */
export interface GroupingBuilder extends ChartBuilder {
  build(data: DataPoint[], ctx: BuildContext): ChartData
}

export const builderStatType = (chart: ChartData): string => chart.statType ?? 'grouped'
