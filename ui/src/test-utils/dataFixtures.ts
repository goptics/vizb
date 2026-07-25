import type { Axis, DataPoint, Sort } from '@/types'

/** Default sort fixtures used across transform/worker/pipeline tests. */
export const noSort: Sort = { enabled: false, order: 'asc' }
export const ascSort: Sort = { enabled: true, order: 'asc' }
export const descSort: Sort = { enabled: true, order: 'desc' }

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
