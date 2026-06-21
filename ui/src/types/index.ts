export type SortOrder = 'asc' | 'desc'
export const SORT_ORDERS: SortOrder[] = ['asc', 'desc']

export type StatMath =
  | 'counts'
  | 'center'
  | 'spread'
  | 'extremes'
  | 'shape'
  | 'percentiles'
  | 'confidence'
  | 'correlations'

export type StatConfig = {
  enabled: boolean
  math: StatMath[] // empty = all categories
}

export type ChartType = 'bar' | 'line' | 'scatter' | 'pie' | 'heatmap' | 'radar'

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

// Ordered axis dimension — replaces the flat AxisLabels on DataSet.
// `key` is the canonical short key ("name" | "x" | "y" | "z");
// `label` is the human-readable column name from --group.
export type Axis = {
  key: 'name' | 'x' | 'y' | 'z'
  label?: string
  type?: string // 'value' = continuous numeric axis; absent or '' = category (default)
}

// Per-chart typed configs (wire format: `Dataset.Settings []ChartConfig`).
// Each chart type carries only the fields that apply to it. The `type`
// discriminator narrows the union at the call site — chart-rendering code may
// still use `cfg.type === 'bar' || cfg.type === 'line'` to access `scale` /
// `threeDRotate` (those fields are absent on pie/heatmap/radar). The settings
// panel is fully schema-less: it walks `Object.keys(activeConfig)` and renders
// the registered control for each non-`type` key.
export type BarConfig = {
  type: 'bar'
  swap?: string
  sort?: Sort
  scale?: ScaleType
  showLabels?: boolean
  threeDRotate?: boolean
  threeD?: boolean
  threeDVisualMap?: boolean
  stat?: StatConfig
}

export type LineConfig = {
  type: 'line'
  swap?: string
  sort?: Sort
  scale?: ScaleType
  showLabels?: boolean
  threeDRotate?: boolean
  threeD?: boolean
  threeDVisualMap?: boolean
  stat?: StatConfig
}

export type ScatterConfig = {
  type: 'scatter'
  swap?: string
  sort?: Sort
  scale?: ScaleType
  showLabels?: boolean
  threeDRotate?: boolean
  threeD?: boolean
  threeDVisualMap?: boolean
  stat?: StatConfig
}

export type PieConfig = {
  type: 'pie'
  swap?: string
  sort?: Sort
  showLabels?: boolean
  stat?: StatConfig
}

export type HeatmapConfig = {
  type: 'heatmap'
  swap?: string
  sort?: Sort
  showLabels?: boolean
  stat?: StatConfig
}

export type RadarConfig = {
  type: 'radar'
  swap?: string
  sort?: Sort
  showLabels?: boolean
  stat?: StatConfig
}

export type ChartConfig =
  | BarConfig
  | LineConfig
  | ScatterConfig
  | PieConfig
  | HeatmapConfig
  | RadarConfig

// Human-readable label for each dimension, derived from the --group columns.
// `name` is carried (though not rendered as an axis) so the swap feature can
// rotate it onto x/y/z carrying its label.
export type AxisLabels = {
  name?: string
  x?: string
  y?: string
  z?: string
}

// Machine metadata, nested under `meta` on both the dataset and each history
// entry. `cpu` is absent when there is no CPU info.
export type Meta = {
  cpu?: {
    name?: string
    cores?: number
  }
  os?: string
  arch?: string
  pkg?: string
}

export type HistoryEntry = {
  tag: string
  timestamp: string
  meta?: Meta
}

export type DataSet = {
  name: string
  description?: string
  tag?: string
  timestamp?: string
  history?: HistoryEntry[]
  meta?: Meta
  axes?: Axis[]
  settings: ChartConfig[]
  data: DataPoint[]
}

// Full descriptive-statistics profile of one numeric vector (a series' values
// across categories). Produced by lib/stats.ts `describe`. NaN where undefined
// (e.g. cv when mean is 0, shape stats for n<2).
export type DescriptiveStats = {
  // counts
  count: number
  missing: number
  unique: number
  zeros: number
  negatives: number
  // center
  mean: number
  median: number
  mode: number
  geoMean: number
  harmMean: number
  trimMean: number
  // spread
  variance: number
  stdDev: number
  cv: number
  sem: number
  cqv: number
  // extremes
  min: number
  max: number
  range: number
  iqr: number
  mad: number
  lowerFence: number
  upperFence: number
  outliers: number
  // shape
  skewness: number
  kurtosis: number
  // percentiles
  p1: number
  p5: number
  p10: number
  p25: number
  p75: number
  p90: number
  p95: number
  p99: number
  // confidence
  ci95Lower: number
  ci95Upper: number
}

// One series' descriptive profile (column profile, YData/D-Tale style).
export type SeriesProfile = {
  name: string
  stats: DescriptiveStats
}

// Symmetric correlation matrices across the chart's auto-picked entity axis (the
// series, the category axis, or the z axis — see `selectCorrelationAxis`). `axis`
// names which one so the panel can caption it; `labels` are that axis's values.
// All 4 methods are precomputed so the panel toggles with no recompute.
export type CorrelationMatrix = {
  axis: 'x' | 'y' | 'z'
  labels: string[]
  pearson: number[][]
  spearman: number[][]
  kendall: number[][]
  dcor: number[][]
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
  valueTuples?: [number, number][] // value-mode 2D: chart [x, y] coordinate pairs
  valuePoints3D?: [number, number, number][] // value-mode 3D: chart [x, y, z] coordinates
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
  mode?: 'grouped' | 'value' | 'continuous'
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
  values: (number | null)[] // null = no data for that category (missing cell)
  benchmarkId: string
}

export type Point3D = {
  xAxis: string
  yAxis: string
  zAxis: string
  value: number
}
