export type SortOrder = 'asc' | 'desc'
export const SORT_ORDERS: SortOrder[] = ['asc', 'desc']

export type ChartType = 'bar' | 'line' | 'pie'

export type ScaleType = 'linear' | 'log'
export const SCALE_TYPES: ScaleType[] = ['linear', 'log']

export type Stat = {
  type: string
  value?: number
  unit?: string
  per?: string
}

export type DataPoint = {
  name?: string
  xAxis?: string
  yAxis?: string
  zAxis?: string
  stats: Stat[]
}

export type Sort = {
  enabled: boolean
  order: SortOrder
}

export type Settings = {
  sort: Sort
  showLabels: boolean
  charts: ChartType[]
  scale: ScaleType
}

// Human-readable label for each dimension, derived from the --group columns.
// `name` is carried (though not rendered as an axis) so the swap feature can
// rotate it onto x/y/z carrying its label.
export type AxisLabels = {
  name?: string
  x?: string
  y?: string
  z?: string
}

export type HistoryEntry = {
  tag: string
  timestamp: string
  cpu?: {
    name?: string
    cores?: number
  }
  os?: string
}

export type DataSet = {
  name: string
  description?: string
  pkg?: string
  tag?: string
  timestamp?: string
  os?: string
  history?: HistoryEntry[]
  cpu?: {
    name?: string
    cores?: number
  }
  settings: Settings
  axisLabels?: AxisLabels
  data: DataPoint[]
}

export type ChartData = {
  title: string
  statType: string
  statUnit?: string
  yAxis: string[]
  zAxis: string[]
  series: SeriesData[]
  points: Point3D[]
  axisLabels?: AxisLabels
}

export type SeriesData = {
  xAxis: string
  values: number[]
  benchmarkId: string
}

export type Point3D = {
  xAxis: string
  yAxis: string
  zAxis: string
  value: number
}
