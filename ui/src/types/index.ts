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

export type BenchmarkData = {
  name?: string
  yAxis?: string
  xAxis?: string
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

export type HistoryEntry = {
  tag: string
  timestamp: string
  cpu?: {
    name?: string
    cores?: number
  }
  os?: string
}

export type Benchmark = {
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
  data: BenchmarkData[]
}

export type ChartData = {
  title: string
  statType: string
  statUnit?: string
  yAxis: string[]
  zAxis: string[]
  series: SeriesData[]
  points: Point3D[]
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
