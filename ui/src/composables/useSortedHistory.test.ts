import { describe, it, expect } from 'vitest'
import { ref } from 'vue'
import type { HistoryEntry } from '@/types'
import { useSortedHistory } from './useSortedHistory'

const entry = (timestamp: string, tag = 't'): HistoryEntry => ({ tag, timestamp })

describe('useSortedHistory', () => {
  it('returns empty when history is undefined', () => {
    const { sortedHistory, hasHistory } = useSortedHistory(ref(undefined))
    expect(sortedHistory.value).toEqual([])
    expect(hasHistory.value).toBe(false)
  })

  it('returns empty when history is empty', () => {
    const { sortedHistory, hasHistory } = useSortedHistory(ref([]))
    expect(sortedHistory.value).toEqual([])
    expect(hasHistory.value).toBe(false)
  })

  it('sorts newest first without filter', () => {
    const history = ref([
      entry('2024-01-01T00:00:00Z', 'old'),
      entry('2024-03-01T00:00:00Z', 'new'),
      entry('2024-02-01T00:00:00Z', 'mid'),
    ])
    const { sortedHistory, hasHistory } = useSortedHistory(history)
    expect(sortedHistory.value.map((e) => e.tag)).toEqual(['new', 'mid', 'old'])
    expect(hasHistory.value).toBe(true)
  })

  it('applies filter before sort', () => {
    const history = ref([
      entry('2024-01-01T00:00:00Z', 'a'),
      entry('2024-03-01T00:00:00Z', 'keep-new'),
      entry('2024-02-01T00:00:00Z', 'keep-mid'),
      entry('2024-04-01T00:00:00Z', 'skip'),
    ])
    const { sortedHistory, hasHistory } = useSortedHistory(history, (e) => e.tag.startsWith('keep'))
    expect(sortedHistory.value.map((e) => e.tag)).toEqual(['keep-new', 'keep-mid'])
    expect(hasHistory.value).toBe(true)
  })

  it('hasHistory is false when filter removes everything', () => {
    const history = ref([entry('2024-01-01T00:00:00Z', 'x')])
    const { sortedHistory, hasHistory } = useSortedHistory(history, () => false)
    expect(sortedHistory.value).toEqual([])
    expect(hasHistory.value).toBe(false)
  })
})
