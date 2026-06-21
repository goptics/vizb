// Framework-free axis-swap helpers, shared by the swap UI (SwapControl.vue), the
// chart pipeline and the transform Web Worker. Swapping rotates which dataset
// field (name/x/y/z) each axis value lives on. The worker re-projects its cached
// raw dataset under the new arrangement (see `projectAndGroup`); these helpers
// translate the compact arrangement strings and permute the display labels. No
// Vue, no echarts — pure data in/out.
import type { AxisLabels, DataPoint } from '../types'

const AXIS_ORDER = ['n', 'x', 'y', 'z'] as const
const VALUE_MODE_POOL = ['x', 'y', 'z'] as const

const kPermutations = (pool: readonly string[], k: number): string[] => {
  if (k <= 0) return ['']
  const result: string[] = []
  pool.forEach((key, i) => {
    const rest = [...pool.slice(0, i), ...pool.slice(i + 1)]
    for (const perm of kPermutations(rest, k - 1)) result.push(key + perm)
  })
  return result
}

export const presentAxisKeys = (data: DataPoint[] | undefined): string[] => {
  if (!data?.length) return []
  const fieldFor = { n: 'name', x: 'xAxis', y: 'yAxis', z: 'zAxis' } as const
  return AXIS_ORDER.filter((k) => data.some((d) => d[fieldFor[k]]))
}

/** Swap dropdown options for SwapControl; value mode permutes x/y/z only. */
export const swapOptionKeys = (data: DataPoint[] | undefined, valueMode = false): string[] => {
  if (!data?.length) return []
  const present = presentAxisKeys(data)
  const pool = valueMode ? VALUE_MODE_POOL : AXIS_ORDER
  const k = valueMode ? present.filter((key) => key !== 'n').length : present.length
  return kPermutations(pool, k)
    .filter((key) => !key.includes('z') || (key.includes('x') && key.includes('y')))
    .filter((key) => key !== 'n')
}

export type AxisKey = 'name' | 'xAxis' | 'yAxis' | 'zAxis'

// Translate a compact arrangement string (e.g. "nxy") into the DataPoint field
// keys it maps to (e.g. ['name','xAxis','yAxis']).
// Which raw field feeds a chart axis under the current identity → target mapping.
export const sourceFieldForChartAxis = (
  identityKeys: AxisKey[],
  targetKeys: AxisKey[],
  chartAxis: 'xAxis' | 'yAxis' | 'zAxis'
): AxisKey | undefined => {
  const i = targetKeys.indexOf(chartAxis)
  if (i < 0 || i >= identityKeys.length) return undefined
  return identityKeys[i]
}

export const translateAxisKey = (key: string): AxisKey[] => {
  const keyMap = {
    x: 'xAxis',
    y: 'yAxis',
    n: 'name',
    z: 'zAxis',
  }
  return key.split('').map((k) => keyMap[k as keyof typeof keyMap]) as AxisKey[]
}

// True when the active swap places z on a chart axis (grouped 3D). Value-mode
// "3D view" is only offered when z is folded off the chart (e.g. xyn vs xyz).
export const arrangementHasChartZ = (targetKey: string): boolean =>
  translateAxisKey(targetKey).includes('zAxis')

// Axis values move between dimensions on swap; the dataset's axisLabels are keyed
// by dimension, so permute them by the same currentKeys → targetKeys mapping or
// they'd point at the wrong axis. Returns a fresh object so the chart computeds
// re-read it. No-op when there are no labels (benchmark inputs).
const LABEL_KEY_FOR: Record<AxisKey, keyof AxisLabels> = {
  name: 'name',
  xAxis: 'x',
  yAxis: 'y',
  zAxis: 'z',
}

export const swapAxisLabels = (
  currentKey: string,
  targetKey: string,
  labels: AxisLabels | undefined
): AxisLabels | undefined => {
  if (!labels) return labels

  const currentKeys = translateAxisKey(currentKey)
  const targetKeys = translateAxisKey(targetKey)
  if (currentKeys.length !== targetKeys.length) return labels

  const values = currentKeys.map((k) => labels[LABEL_KEY_FOR[k]])
  const next: AxisLabels = { ...labels }
  for (const k of currentKeys) delete next[LABEL_KEY_FOR[k]]
  targetKeys.forEach((k, i) => {
    next[LABEL_KEY_FOR[k]] = values[i]
  })

  return next
}
