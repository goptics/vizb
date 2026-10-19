import type { Better, ResolvedSort, Sort } from '@/types'

/**
 * Sort with `best` resolved for one chart: the winner goes first — ascending
 * for lower-is-better, descending for higher-is-better and unknown metrics
 * (most unlabeled metrics are higher-is-better; matches the ★ marker rule).
 */
export function resolveSort(sort: Sort, better?: Better): ResolvedSort {
  const order = sort.order !== 'best' ? sort.order : better === 'low' ? 'asc' : 'desc'
  return { enabled: sort.enabled, order }
}
