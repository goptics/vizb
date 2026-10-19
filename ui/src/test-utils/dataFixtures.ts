import type { Axis, DataPoint, ResolvedSort } from '@/types'

/** Default sort fixtures used across transform/worker/pipeline tests. */
export const noSort: ResolvedSort = { enabled: false, order: 'asc' }
export const ascSort: ResolvedSort = { enabled: true, order: 'asc' }
export const descSort: ResolvedSort = { enabled: true, order: 'desc' }

/** Minimal DataPoint row for transform / worker / pipeline suites. */
export function dp(xAxis: string, yAxis = '', zAxis = '', type = 'val', value = 1): DataPoint {
  return { xAxis, yAxis, zAxis, stats: [{ type, value }] }
}

/** Continuous numeric axes used by value-mode charts. */
export const VALUE_AXES: Axis[] = [
  { key: 'x', label: 'price', type: 'value' },
  { key: 'y', label: 'latency', type: 'value' },
]

/** Category x + value y axes used by mixed-mode charts. */
export const MIXED_AXES: Axis[] = [
  { key: 'x', label: 'region' },
  { key: 'y', label: 'latency', type: 'value' },
]
