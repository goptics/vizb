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
}

export interface SeriesData {
  subject: string
  values: number[]
}

export interface Settings {
  sort: 'asc' | 'desc' | ''
  showLabels: boolean
}

export interface Benchmark {
  name: string
  description: string
  cpu: string
  settings: Settings
  results: BenchmarkResult[]
}

export type SortOrder = 'asc' | 'desc' | 'default'
