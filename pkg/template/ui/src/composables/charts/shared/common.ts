import type { SortOrder, ScaleType } from '../../../types'

export const fontSize = 12

export const sortBy =
  <K extends string>(key: K) =>
  <T extends Record<K, number>>(sortOrder: SortOrder) => {
    if (sortOrder === 'asc') {
      return (a: T, b: T) => a[key] - b[key]
    }

    return (a: T, b: T) => b[key] - a[key]
  }

export const sortByTotal = sortBy('total')

export const sortByValue = sortBy('value')

// For line charts: use null for zero values to create gaps instead of dropping below axis
export const adjustForLogScaleLine = (value: number, scale: ScaleType): number | null => {
  if (scale !== 'log') return value
  return value <= 0 ? null : value
}
