import type { SortOrder } from '../../../types'

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
