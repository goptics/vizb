export type SortOrder = 'asc' | 'desc'

export type ChartType = 'bar' | 'line' | 'pie'

export type Stat = {
  type: string
  value?: number
  unit?: string
  per?: string
}

// Represents a single benchmark result for one subject
export type BenchmarkData = {
  name?: string
  yAxis?: string
  xAxis?: string
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
}

export type Benchmark = {
  name: string
  description?: string
  pkg?: string
  cpu?: {
    name?: string
    cores?: number
  }
  settings: Settings
  data: BenchmarkData[]
}

// Chart data structure
export type ChartData = {
  title: string
  statType: string
  statUnit?: string
  yAxis: string[]
  series: SeriesData[]
}

export type SeriesData = {
  xAxis: string
  values: number[]
  benchmarkId: string
}
