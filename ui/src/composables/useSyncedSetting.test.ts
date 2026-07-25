import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useSyncedSetting } from './useSyncedSetting'

describe('useSyncedSetting', () => {
  it('reads via getter and writes via setter', () => {
    const store = ref('a')
    const write = vi.fn((val: string) => {
      store.value = val
    })
    const synced = useSyncedSetting(() => store.value, write)

    expect(synced.value).toBe('a')
    synced.value = 'b'
    expect(write).toHaveBeenCalledWith('b')
    expect(synced.value).toBe('b')
  })
})
