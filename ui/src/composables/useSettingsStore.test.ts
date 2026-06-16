import { describe, it, expect, beforeEach, vi } from 'vitest'
import { ref, type Ref } from 'vue'
import type { DataSet, ChartConfig } from '../types'

// vi.hoisted runs before the import statements — store the holder in a
// closure that the vi.mock factory can read at any time, and replace the ref
// in beforeEach so each test sees a fresh dataset.
const holder = vi.hoisted(() => ({
  ref: undefined as Ref<DataSet | undefined> | undefined,
}))

vi.mock('./useDataPoint', () => ({
  get activeDataSet() {
    if (!holder.ref) {
      throw new Error('test forgot to set holder.ref in beforeEach')
    }
    return holder.ref
  },
}))

const ds = (settings: ChartConfig[]): DataSet => ({
  name: 'test',
  settings,
  data: [],
})

describe('useSettingsStore', () => {
  beforeEach(() => {
    // Reset the store module so the module-level activeChartIndex ref starts
    // fresh each test (it persists across tests otherwise).
    vi.resetModules()
    holder.ref = ref(
      ds([
        { type: 'bar', sort: { enabled: false, order: 'asc' }, scale: 'linear' },
        { type: 'pie', sort: { enabled: false, order: 'asc' } },
      ])
    )
  })

  it('activeConfig returns the config at the active chart index', async () => {
    const { useSettingsStore } = await import('./useSettingsStore')
    const { activeConfig, setActiveChartIndex } = useSettingsStore()
    expect(activeConfig.value?.type).toBe('bar')
    setActiveChartIndex(1)
    expect(activeConfig.value?.type).toBe('pie')
  })

  it('setSort writes back to dataset.settings[i].sort', async () => {
    const { useSettingsStore } = await import('./useSettingsStore')
    const { activeConfig, setSort } = useSettingsStore()
    setSort({ enabled: true, order: 'desc' })
    expect(activeConfig.value?.sort).toEqual({ enabled: true, order: 'desc' })
    expect(holder.ref!.value!.settings[0]!.sort).toEqual({
      enabled: true,
      order: 'desc',
    })
  })
})
