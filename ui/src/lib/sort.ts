import type { Better, ResolvedSort, ResolvedSortOrder, Sort, SortOrder } from '@/types'

/**
 * Concrete direction for a chart. `best` puts the winner first: ascending for
 * lower-is-better, descending for higher-is-better and unknown metrics (most
 * unlabeled metrics are higher-is-better; matches the ★ marker rule).
 */
export function effectiveSortOrder(order: SortOrder, better?: Better): ResolvedSortOrder {
  if (order !== 'best') return order
  return better === 'low' ? 'asc' : 'desc'
}

/** Sort with `best` resolved against one chart's direction. */
export function resolveSort(sort: Sort, better?: Better): ResolvedSort {
  return { enabled: sort.enabled, order: effectiveSortOrder(sort.order, better) }
}
