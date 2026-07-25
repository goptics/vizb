import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { ref, type Ref } from 'vue'
import type { Dataset, ChartConfig } from '../types'
import { ds } from '@/test-utils'

// vi.hoisted runs before the import statements — store the holder in a
// closure that the vi.mock factory can read at any time, and replace the ref
// in beforeEach so each test sees a fresh dataset.
const holder = vi.hoisted(() => ({
  ref: undefined as Ref<Dataset | undefined> | undefined,
}))

vi.mock('./useDataPoint', () => ({
  get activeDataset() {
    if (!holder.ref) {
      throw new Error('test forgot to set holder.ref in beforeEach')
    }
    return holder.ref
  },
}))

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

  afterEach(() => vi.unstubAllGlobals())

  it('initializes from the dataset, persists viewer choice, and does not overwrite it', async () => {
    const values = new Map<string, string>()
    const storage = {
      getItem: vi.fn((key: string) => values.get(key) ?? null),
      setItem: vi.fn((key: string, value: string) => values.set(key, value)),
    }
    vi.stubGlobal('localStorage', storage)
    vi.stubGlobal('window', { matchMedia: () => ({ matches: false }) })
    vi.stubGlobal('document', { documentElement: { classList: { toggle: vi.fn() } } })

    const { useSettingsStore } = await import('./useSettingsStore')
    const { themeName, initializeTheme, setTheme } = useSettingsStore()
    initializeTheme('vintage')
    expect(themeName.value).toBe('vintage')

    setTheme('roma')
    expect(themeName.value).toBe('roma')
    expect(storage.setItem).toHaveBeenCalledWith('color-theme', 'roma')

    initializeTheme('chalk')
    expect(themeName.value).toBe('roma')
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
  it.each([
    {
      label: 'setThreeDRotate',
      config: {
        type: 'bar',
        sort: { enabled: false, order: 'asc' },
        scale: 'linear',
      } as ChartConfig,
      apply: async () => {
        const { useSettingsStore } = await import('./useSettingsStore')
        const store = useSettingsStore()
        store.setThreeDRotate(true)
        return store.activeConfig
      },
      field: 'threeDRotate' as const,
      value: true as const,
    },
    {
      label: 'setScale',
      config: {
        type: 'bar',
        sort: { enabled: false, order: 'asc' },
      } as ChartConfig,
      apply: async () => {
        const { useSettingsStore } = await import('./useSettingsStore')
        const store = useSettingsStore()
        store.setScale('log')
        return store.activeConfig
      },
      field: 'scale' as const,
      value: 'log' as const,
    },
  ])(
    '$label writes even when the field is absent on the config',
    async ({ config, apply, field, value }) => {
      holder.ref = ref(ds([config]))
      const activeConfig = await apply()
      expect((activeConfig.value as Record<string, unknown> | undefined)?.[field]).toBe(value)
    }
  )

  it('setSmooth writes only to line configs', async () => {
    holder.ref = ref(
      ds([
        { type: 'line', sort: { enabled: false, order: 'asc' } },
        { type: 'bar', sort: { enabled: false, order: 'asc' } },
      ])
    )
    const { useSettingsStore } = await import('./useSettingsStore')
    const { activeConfig, setActiveChartIndex, setSmooth } = useSettingsStore()

    setSmooth(true)
    expect((activeConfig.value as { smooth?: boolean } | undefined)?.smooth).toBe(true)

    setActiveChartIndex(1)
    setSmooth(true)
    expect((activeConfig.value as { smooth?: boolean } | undefined)?.smooth).toBeUndefined()
  })

  it('setStack writes even when the field is absent on the config', async () => {
    holder.ref = ref(
      ds([
        // No stack field — mimics a config the user hasn't toggled stacking on yet.
        { type: 'bar', sort: { enabled: false, order: 'asc' }, scale: 'log' } as ChartConfig,
      ])
    )
    const { useSettingsStore } = await import('./useSettingsStore')
    const { activeConfig, setStack } = useSettingsStore()
    setStack(true)
    expect((activeConfig.value as { stack?: boolean } | undefined)?.stack).toBe(true)
    expect((activeConfig.value as { scale?: string } | undefined)?.scale).toBe('linear')
  })

  it('covers every remaining setter and chart-type helpers', async () => {
    holder.ref = ref(
      ds([
        {
          type: 'bar',
          sort: { enabled: false, order: 'asc' },
          scale: 'linear',
        } as ChartConfig,
        {
          type: 'line',
          sort: { enabled: false, order: 'asc' },
        } as ChartConfig,
        {
          type: 'scatter',
          sort: { enabled: false, order: 'asc' },
        } as ChartConfig,
      ])
    )

    const { useSettingsStore } = await import('./useSettingsStore')
    const store = useSettingsStore()

    store.setShowLabels(true)
    expect(store.activeConfig.value?.showLabels).toBe(true)

    store.setHorizontal(true)
    expect((store.activeConfig.value as { horizontal?: boolean }).horizontal).toBe(true)

    store.setSwap('yx')
    expect(store.activeConfig.value?.swap).toBe('yx')

    store.setThreeD(true)
    expect((store.activeConfig.value as { threeD?: boolean }).threeD).toBe(true)

    store.setThreeDVisualMap(true)
    expect((store.activeConfig.value as { threeDVisualMap?: boolean }).threeDVisualMap).toBe(true)

    store.setChartType('scatter')
    expect(store.chartType.value).toBe('scatter')
    expect(store.activeChartIndex.value).toBe(2)

    store.setVisualMap(true)
    expect((store.activeConfig.value as { visualMap?: boolean }).visualMap).toBe(true)

    store.setChartType('missing' as never)
    expect(store.chartType.value).toBe('scatter')

    store.setActiveChartIndex(99)
    expect(store.activeChartIndex.value).toBe(2)

    store.setActiveChartIndex(1)
    store.setHorizontal(true)
    expect((store.activeConfig.value as { horizontal?: boolean }).horizontal).toBeUndefined()

    store.setVisualMap(true)
    expect((store.activeConfig.value as { visualMap?: boolean }).visualMap).toBeUndefined()
  })

  it('clamps activeChartIndex when settings shrink and defaults chartType', async () => {
    holder.ref = ref(
      ds([
        { type: 'bar', sort: { enabled: false, order: 'asc' } },
        { type: 'pie', sort: { enabled: false, order: 'asc' } },
      ])
    )
    const { useSettingsStore } = await import('./useSettingsStore')
    const { activeChartIndex, setActiveChartIndex, chartType } = useSettingsStore()
    setActiveChartIndex(1)
    expect(chartType.value).toBe('pie')

    holder.ref!.value = ds([{ type: 'bar', sort: { enabled: false, order: 'asc' } }])
    await Promise.resolve()
    expect(activeChartIndex.value).toBe(0)
    expect(chartType.value).toBe('bar')

    holder.ref!.value = undefined
    await Promise.resolve()
    expect(chartType.value).toBe('bar')
  })

  it('initializes dark mode and theme preference from localStorage and toggles dark', async () => {
    const values = new Map<string, string>([
      ['dark-mode', 'true'],
      ['color-theme', 'macarons'],
    ])
    const storage = {
      getItem: vi.fn((key: string) => values.get(key) ?? null),
      setItem: vi.fn((key: string, value: string) => values.set(key, value)),
    }
    const classList = { toggle: vi.fn() }
    vi.stubGlobal('localStorage', storage)
    vi.stubGlobal('window', {
      matchMedia: () => ({ matches: false }),
    })
    vi.stubGlobal('document', { documentElement: { classList } })

    const { useSettingsStore } = await import('./useSettingsStore')
    const { isDark, themeName, toggleDark, initializeTheme } = useSettingsStore()

    expect(isDark.value).toBe(true)
    expect(themeName.value).toBe('macarons')
    initializeTheme('vintage')
    expect(themeName.value).toBe('macarons')

    toggleDark()
    expect(isDark.value).toBe(false)
    expect(storage.setItem).toHaveBeenCalledWith('dark-mode', 'false')
    expect(classList.toggle).toHaveBeenCalled()
  })

  it('setStack without enabling keeps scale when stack is false', async () => {
    holder.ref = ref(
      ds([{ type: 'bar', sort: { enabled: false, order: 'asc' }, scale: 'log' } as ChartConfig])
    )
    const { useSettingsStore } = await import('./useSettingsStore')
    const { activeConfig, setStack } = useSettingsStore()
    setStack(false)
    expect((activeConfig.value as { stack?: boolean }).stack).toBe(false)
    expect((activeConfig.value as { scale?: string }).scale).toBe('log')
  })

  it('no-ops setters when there is no active config', async () => {
    holder.ref = ref(undefined)
    const { useSettingsStore } = await import('./useSettingsStore')
    const store = useSettingsStore()

    store.setSort({ enabled: true, order: 'desc' })
    store.setScale('log')
    store.setStack(true)
    store.setShowLabels(true)
    store.setSmooth(true)
    store.setHorizontal(true)
    store.setThreeDRotate(true)
    store.setSwap('yx')
    store.setThreeD(true)
    store.setThreeDVisualMap(true)
    store.setVisualMap(true)
    store.setChartType('bar')
    store.setActiveChartIndex(0)

    expect(store.activeConfig.value).toBeUndefined()
    expect(store.chartType.value).toBe('bar')
  })

  it('prefers system dark scheme when no localStorage preference exists', async () => {
    const storage = {
      getItem: vi.fn(() => null),
      setItem: vi.fn(),
    }
    const classList = { toggle: vi.fn() }
    vi.stubGlobal('localStorage', storage)
    vi.stubGlobal('window', {
      matchMedia: () => ({ matches: true }),
    })
    vi.stubGlobal('document', { documentElement: { classList } })

    const { useSettingsStore } = await import('./useSettingsStore')
    const { isDark } = useSettingsStore()
    expect(isDark.value).toBe(true)
  })

  it('is import-safe without browser globals', async () => {
    vi.unstubAllGlobals()
    const g = globalThis as typeof globalThis & {
      window?: unknown
      document?: unknown
    }
    const hadWindow = Object.prototype.hasOwnProperty.call(g, 'window')
    const hadDocument = Object.prototype.hasOwnProperty.call(g, 'document')
    const prevWindow = g.window
    const prevDocument = g.document
    Reflect.deleteProperty(g, 'window')
    Reflect.deleteProperty(g, 'document')

    try {
      vi.resetModules()
      const { useSettingsStore } = await import('./useSettingsStore')
      const store = useSettingsStore()
      store.toggleDark()
      store.setTheme('roma')
      expect(store.themeName.value).toBe('roma')
      expect(typeof store.isDark.value).toBe('boolean')
    } finally {
      if (hadWindow) g.window = prevWindow
      else Reflect.deleteProperty(g, 'window')
      if (hadDocument) g.document = prevDocument
      else Reflect.deleteProperty(g, 'document')
    }
  })
})
