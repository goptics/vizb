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

  // Regression: a freshly migrated config (or a config created in the UI) may
  // not carry `scale` / `threeDRotate` yet — the Go migration does not pre-
  // populate `threeDRotate` (it didn't exist in v0.12.0). The setters used to
  // guard on `'field' in cfg`, which silently no-oped the first toggle. The
  // panel already filters by `appliesTo`, so writing the field is always safe.
  it('setThreeDRotate writes even when the field is absent on the config', async () => {
    holder.ref = ref(
      ds([
        // No threeDRotate field — mimics a Go-migrated v0.12.0 config.
        { type: 'bar', sort: { enabled: false, order: 'asc' }, scale: 'linear' } as ChartConfig,
      ])
    )
    const { useSettingsStore } = await import('./useSettingsStore')
    const { activeConfig, setThreeDRotate } = useSettingsStore()
    setThreeDRotate(true)
    expect((activeConfig.value as { threeDRotate?: boolean } | undefined)?.threeDRotate).toBe(true)
  })

  it('setScale writes even when the field is absent on the config', async () => {
    holder.ref = ref(
      ds([
        // No scale field — mimics a config the user hasn't set the scale on yet.
        { type: 'bar', sort: { enabled: false, order: 'asc' } } as ChartConfig,
      ])
    )
    const { useSettingsStore } = await import('./useSettingsStore')
    const { activeConfig, setScale } = useSettingsStore()
    setScale('log')
    expect((activeConfig.value as { scale?: string } | undefined)?.scale).toBe('log')
  })
})
