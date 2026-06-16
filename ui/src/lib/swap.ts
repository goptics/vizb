// Framework-free axis-swap helpers, shared by the swap UI (AxisSwapper.vue), the
// chart pipeline and the transform Web Worker. Swapping rotates which dataset
// field (name/x/y/z) each axis value lives on. The worker re-projects its cached
// raw dataset under the new arrangement (see `projectAndGroup`); these helpers
// translate the compact arrangement strings and permute the display labels. No
// Vue, no echarts — pure data in/out.
import type { Axis, AxisLabels } from '../types'

// Concatenate an axis list into the compact arrangement string (e.g.
// [{key:"x"},{key:"y"},{key:"name"}] → "xyn"). The single-char convention is:
// name → 'n', x/y/z → first char of the key. Returns '' for an empty axis list
// — callers default to identity-string fallback when the dataset has no axes.
// Mirrors the Go side's `axisIdentity` in `shared/chart_spec.go`.
export const axisKeyConcat = (axes: Axis[] | undefined): string => {
  if (!axes?.length) return ''
  return axes
    .map((a) => (a.key === 'name' ? 'n' : a.key.charAt(0)))
    .join('')
}

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
