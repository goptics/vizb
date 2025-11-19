/**
 * Benchmark Data Type Definitions
 * Defines the structure for benchmark visualization data
 */


export interface Stat {
  type: string
  value: number
  unit: string
  per: string
}

// Represents a single benchmark result for one subject
export interface BenchmarkData {
  name: string
  yAxis: string
  xAxis: string
  stats: Stat[]
}

// Processed chart data structure
export interface ChartData {
  title: string
  statType: string
  statUnit: string
  yAxis: string[]
  series: SeriesData[]
}

export interface SeriesData {
  xAxis: string
  values: number[]
  benchmarkId: string
}

export type Sort = {
  enabled: boolean
  order: SortOrder
}

export interface Settings {
  sort: Sort
  showLabels: boolean
  charts: ChartType[]
}

export const DEFAULT_SETTINGS: Settings = {
  sort: {
    enabled: false,
    order: 'asc',
  },
  showLabels: false,
  charts: ['line', 'pie', 'bar'],
}

export interface Benchmark {
  name: string
  description: string
  cpu: {
    name: string
    cores: number
  }
  settings: Settings
  data: BenchmarkData[]
}

export type SortOrder = 'asc' | 'desc'

export type ChartType = 'bar' | 'line' | 'pie'
