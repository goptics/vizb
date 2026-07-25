import type { LabelMode } from '@/types'

export const formatPercentageLabel = (value: number, total: number): string => {
  if (!Number.isFinite(value) || !Number.isFinite(total) || total === 0) return ''
  return `${Math.round((value / total) * 10000) / 100}%`
}

export const percentageFormatter = <T>(
  mode: LabelMode,
  total: number,
  value: (params: T) => number | null | undefined
) =>
  mode === 'percentage'
    ? (params: T) => formatPercentageLabel(value(params) ?? NaN, total)
    : undefined
