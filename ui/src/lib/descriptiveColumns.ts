import type { DescriptiveStats, StatMath } from '../types'

export type DescriptiveStatCategory = Exclude<StatMath, 'correlations'>

export type DescriptiveColumn = {
  key: keyof DescriptiveStats
  label: string
  category: DescriptiveStatCategory
}

export const STAT_CATEGORY_LABELS: Record<DescriptiveStatCategory, string> = {
  counts: 'Counts',
  center: 'Center',
  spread: 'Spread',
  extremes: 'Extremes',
  shape: 'Shape',
  percentiles: 'Percentiles',
  confidence: 'Confidence',
}

export const STAT_CATEGORY_ORDER = [
  'counts',
  'center',
  'spread',
  'extremes',
  'shape',
  'percentiles',
  'confidence',
] as const satisfies readonly DescriptiveStatCategory[]

export const DESCRIPTIVE_COLUMNS: DescriptiveColumn[] = [
  { key: 'count', label: 'Count', category: 'counts' },
  { key: 'missing', label: 'Missing', category: 'counts' },
  { key: 'unique', label: 'Unique', category: 'counts' },
  { key: 'zeros', label: 'Zeros', category: 'counts' },
  { key: 'negatives', label: 'Negatives', category: 'counts' },
  { key: 'mean', label: 'Mean', category: 'center' },
  { key: 'median', label: 'Median', category: 'center' },
  { key: 'mode', label: 'Mode', category: 'center' },
  { key: 'geoMean', label: 'Geo Mean', category: 'center' },
  { key: 'harmMean', label: 'Harm Mean', category: 'center' },
  { key: 'trimMean', label: 'Trim Mean', category: 'center' },
  { key: 'stdDev', label: 'SD', category: 'spread' },
  { key: 'variance', label: 'Variance', category: 'spread' },
  { key: 'cv', label: 'CV', category: 'spread' },
  { key: 'sem', label: 'SEM', category: 'spread' },
  { key: 'cqv', label: 'CQV', category: 'spread' },
  { key: 'min', label: 'Min', category: 'extremes' },
  { key: 'max', label: 'Max', category: 'extremes' },
  { key: 'range', label: 'Range', category: 'extremes' },
  { key: 'iqr', label: 'IQR', category: 'extremes' },
  { key: 'mad', label: 'MAD', category: 'extremes' },
  { key: 'lowerFence', label: 'Lower Fence', category: 'extremes' },
  { key: 'upperFence', label: 'Upper Fence', category: 'extremes' },
  { key: 'outliers', label: 'Outliers', category: 'extremes' },
  { key: 'skewness', label: 'Skew', category: 'shape' },
  { key: 'kurtosis', label: 'Kurtosis', category: 'shape' },
  { key: 'p1', label: 'P1', category: 'percentiles' },
  { key: 'p5', label: 'P5', category: 'percentiles' },
  { key: 'p10', label: 'P10', category: 'percentiles' },
  { key: 'p25', label: 'P25', category: 'percentiles' },
  { key: 'p75', label: 'P75', category: 'percentiles' },
  { key: 'p90', label: 'P90', category: 'percentiles' },
  { key: 'p95', label: 'P95', category: 'percentiles' },
  { key: 'p99', label: 'P99', category: 'percentiles' },
  { key: 'ci95Lower', label: '95% CI Low', category: 'confidence' },
  { key: 'ci95Upper', label: '95% CI High', category: 'confidence' },
]

export const CATEGORY_KEYS: Record<DescriptiveStatCategory, (keyof DescriptiveStats)[]> =
  STAT_CATEGORY_ORDER.reduce(
    (acc, category) => {
      acc[category] = DESCRIPTIVE_COLUMNS.filter((col) => col.category === category).map(
        (col) => col.key
      )
      return acc
    },
    {} as Record<DescriptiveStatCategory, (keyof DescriptiveStats)[]>
  )

const DESCRIPTIVE_COLUMN_KEYS = new Set(DESCRIPTIVE_COLUMNS.map((col) => col.key))

export function keysForMath(math?: StatMath[]): (keyof DescriptiveStats)[] {
  if (!math || math.length === 0) return DESCRIPTIVE_COLUMNS.map((col) => col.key)

  const keys = math
    .filter((category): category is DescriptiveStatCategory => category !== 'correlations')
    .flatMap((category) => CATEGORY_KEYS[category])

  return keys.length ? keys : DESCRIPTIVE_COLUMNS.map((col) => col.key)
}

export function defaultSelectedKeys(math?: StatMath[]): (keyof DescriptiveStats)[] {
  return keysForMath(math)
}

export function columnsFromKeys(keys: Iterable<keyof DescriptiveStats>): DescriptiveColumn[] {
  const selected = new Set(keys)
  return DESCRIPTIVE_COLUMNS.filter((col) => selected.has(col.key))
}

export function isDescriptiveColumnKey(key: string): key is keyof DescriptiveStats {
  return DESCRIPTIVE_COLUMN_KEYS.has(key as keyof DescriptiveStats)
}
