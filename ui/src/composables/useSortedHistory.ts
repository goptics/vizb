import { computed } from 'vue'
import type { Ref } from 'vue'
import type { HistoryEntry } from '../types'

export function useSortedHistory(
  history: Ref<HistoryEntry[] | undefined>,
  filterFn?: (entry: HistoryEntry) => boolean
) {
  const sortedHistory = computed(() => {
    if (!history.value?.length) return []
    const entries = filterFn ? history.value.filter(filterFn) : [...history.value]
    return entries.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
  })

  const hasHistory = computed(() => sortedHistory.value.length > 0)

  return { sortedHistory, hasHistory }
}
