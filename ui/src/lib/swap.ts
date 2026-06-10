// Framework-free axis-swap helpers, shared by the swap UI (AxisSwapper.vue) and
// the transform Web Worker. Swapping rotates which dataset field (name/x/y/z)
// each axis value lives on; the worker applies the same O(n) field rename to its
// cached dataset that the main-thread store applies, keeping the two copies in
// sync without re-cloning the rows. No Vue, no echarts — pure data in/out.
import type { DataPoint, AxisLabels } from '../types'

export type AxisKey = 'name' | 'xAxis' | 'yAxis' | 'zAxis'

// Translate a compact arrangement string (e.g. "nxy") into the DataPoint field
// keys it maps to (e.g. ['name','xAxis','yAxis']).
export const translateAxisKey = (key: string): AxisKey[] => {
  const keyMap = {
    x: 'xAxis',
    y: 'yAxis',
    n: 'name',
    z: 'zAxis',
  }
  return key.split('').map((k) => keyMap[k as keyof typeof keyMap]) as AxisKey[]
}

// Rename axis fields in place across every row: read each row's current-key
// values, drop the current keys, then re-assign those values onto the target
// keys by position. Keys are pre-translated by the caller. No-op on length
// mismatch (the arrangements always share length, but guard anyway).
export const swapAxisFields = (
  data: DataPoint[],
  currentKeys: AxisKey[],
  targetKeys: AxisKey[]
): void => {
  if (currentKeys.length !== targetKeys.length) return

  for (const bench of data) {
    const values = currentKeys.map((k) => bench[k])
    for (const k of currentKeys) delete bench[k]
    targetKeys.forEach((k, i) => {
      bench[k] = values[i]
    })
  }
}

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
