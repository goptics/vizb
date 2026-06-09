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
  // Precomputed 3D render data (built in the transform worker for charts that
  // have x, y and z). Absent for 2D charts. Holds the sorted axis category
  // arrays plus the per-z series data for both bar3D (filled grid) and line3D
  // (sparse) so a chart-type switch needs no recompute.
  render3D?: Render3D
}

export type Series3DData = {
  name: string
  data: { value: number[] }[]
}

export type Render3D = {
  xValues: string[]
  yValues: string[]
  zValues: string[]
  barSeries: Series3DData[]
  lineSeries: Series3DData[]
  // Precomputed sum of all z-group values per (xi,yi) cell. Key: "${xi},${yi}".
  // Computed in the transform worker so the Vue computed only does O(1) lookups.
  cellTotals: Record<string, number>
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
