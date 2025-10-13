/**
 * Benchmark Data Type Definitions
 * Defines the structure for benchmark visualization data
 */


export interface Stat {
  type: string
  value: number
  unit: string
}

// Represents a single benchmark result for one subject
export interface BenchmarkResult {
  name: string
  workload: string
  subject: string
  stats: Stat[]
}

// Processed chart data structure
export interface ChartData {
  title: string
  statType: string
  statUnit: string
  workloads: string[]
  series: SeriesData[]
  subjectTotals?: Record<string, number>
}

export interface SeriesData {
  subject: string
  values: number[]
  subjectTotals?: Array<{ subject: string; total: number }>
  benchmarkId: string
}

export interface Settings {
  sort: SortOrder
  showLabels: boolean
}

export interface Benchmark {
  name: string
  description: string
  cpu: string
  settings: Settings
  results: BenchmarkResult[]
}

export type SortOrder = 'asc' | 'desc' | ''

export type ChartType = 'bar' | 'line' | 'pie'
