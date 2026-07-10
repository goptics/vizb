import { computed, ref, watch } from 'vue'
import type { ChartConfig, ChartType, Sort, ScaleType } from '../types'
import { activeDataSet } from './useDataPoint'
import { isValidIndex } from '../lib/utils'
import { activeThemeName, applyTheme, normalizeTheme } from '../lib/themes'

// Module-level singleton state. `activeChartIndex` is the cursor into
// `activeDataSet.value.settings`; the dataset IS the source of truth — no flat
// global state anymore. `setSort` / `setScale` / etc. mutate the active
// config in place, which the Vue reactivity propagates everywhere.
const activeChartIndex = ref(0)
const isDark = ref(false)
const themeName = activeThemeName

// Dark mode is gated so the module is import-safe in node/test environments
// (no `localStorage` / `window` access at load time).
const isBrowser = typeof window !== 'undefined' && typeof document !== 'undefined'

const updateHtmlClass = () => {
  if (!isBrowser) return
  const html = document.documentElement
  html.classList.toggle('dark', isDark.value)
  html.classList.toggle('light', !isDark.value)
}

const initializeDarkMode = () => {
  if (!isBrowser) return
  const saved = localStorage.getItem('dark-mode')
  if (saved !== null) {
    isDark.value = saved === 'true'
  } else {
    isDark.value = window.matchMedia('(prefers-color-scheme: dark)').matches
  }
  updateHtmlClass()
}

initializeDarkMode()

let hasThemePreference = false
if (isBrowser) {
  const savedTheme = localStorage.getItem('color-theme')
  if (savedTheme !== null) {
    hasThemePreference = true
    applyTheme(savedTheme)
  }
}

const toggleDark = () => {
  isDark.value = !isDark.value
  if (isBrowser) {
    localStorage.setItem('dark-mode', isDark.value.toString())
  }
  updateHtmlClass()
}

const initializeTheme = (datasetTheme?: string) => {
  if (!hasThemePreference) applyTheme(datasetTheme)
}

const setTheme = (theme: string) => {
  const normalized = normalizeTheme(theme)
  applyTheme(normalized)
  hasThemePreference = true
  if (isBrowser) localStorage.setItem('color-theme', normalized)
}

export function useSettingsStore() {
  const activeConfig = computed<ChartConfig | undefined>(
    () => activeDataSet.value?.settings[activeChartIndex.value]
  )

  const chartType = computed<ChartType>(() => activeConfig.value?.type ?? 'bar')

  // Clamp the active index when the active dataset's chart list shrinks (e.g.
  // a settings change drops one of the bundled charts).
  watch(
    () => activeDataSet.value?.settings.length,
    (len) => {
      if (len !== undefined && activeChartIndex.value >= len) {
        activeChartIndex.value = 0
      }
    }
  )

  const setActiveChartIndex = (index: number) => {
    const len = activeDataSet.value?.settings.length ?? 0
    if (isValidIndex(index, len)) {
      activeChartIndex.value = index
    }
  }

  // setChartType locates the first config with the requested type and makes it
  // active. No-op if the active dataset has no config of that type.
  const setChartType = (type: ChartType) => {
    const settings: ChartConfig[] = activeDataSet.value?.settings ?? []
    const idx = settings.findIndex((s: ChartConfig) => s.type === type)
    if (idx !== -1) {
      activeChartIndex.value = idx
    }
  }

  // Per-field setters write back to the active config in place. Each config
  // shape carries only the fields that apply to its chart type, so `scale` and
  // `threeDRotate` are only set on bar/line configs (the others have no such
  // field). The narrowing uses TypeScript's optional-field semantics — no
  // runtime type-guard. The SettingsPanel already gates by `appliesTo` in
  // `fieldRegistry`, so the setters are only ever called for the chart types
  // that carry the field. The Go migration does NOT pre-populate `threeDRotate`
  // (it didn't exist in v0.12.0), so an `'threeDRotate' in cfg` guard here would
  // silently no-op the first toggle on a freshly migrated config.
  const setSort = (sort: Sort) => {
    const cfg = activeConfig.value
    if (cfg) cfg.sort = { ...sort }
  }

  const setScale = (scale: ScaleType) => {
    const cfg = activeConfig.value as { scale?: ScaleType } | undefined
    if (cfg) cfg.scale = scale
  }

  const setStack = (stack: boolean) => {
    const cfg = activeConfig.value as { stack?: boolean; scale?: ScaleType } | undefined
    if (cfg) {
      cfg.stack = stack
      if (stack) cfg.scale = 'linear'
    }
  }

  const setShowLabels = (show: boolean) => {
    const cfg = activeConfig.value
    if (cfg) cfg.showLabels = show
  }

  const setSmooth = (smooth: boolean) => {
    const cfg = activeConfig.value
    if (cfg?.type === 'line') cfg.smooth = smooth
  }

  const setHorizontal = (horizontal: boolean) => {
    const cfg = activeConfig.value
    if (cfg?.type === 'bar') cfg.horizontal = horizontal
  }

  const setThreeDRotate = (rotate: boolean) => {
    const cfg = activeConfig.value as { threeDRotate?: boolean } | undefined
    if (cfg) cfg.threeDRotate = rotate
  }

  const setSwap = (swap: string | undefined) => {
    const cfg = activeConfig.value
    if (cfg) cfg.swap = swap
  }

  const setThreeD = (enabled: boolean) => {
    const cfg = activeConfig.value as { threeD?: boolean } | undefined
    if (cfg) cfg.threeD = enabled
  }

  const setThreeDVisualMap = (enabled: boolean) => {
    const cfg = activeConfig.value as { threeDVisualMap?: boolean } | undefined
    if (cfg) cfg.threeDVisualMap = enabled
  }

  const setVisualMap = (enabled: boolean) => {
    const cfg = activeConfig.value
    if (cfg?.type === 'scatter') cfg.visualMap = enabled
  }

  return {
    activeChartIndex,
    activeConfig,
    chartType,
    isDark,
    themeName,
    setActiveChartIndex,
    setChartType,
    setSort,
    setScale,
    setStack,
    setShowLabels,
    setSmooth,
    setHorizontal,
    setThreeDRotate,
    setSwap,
    setThreeD,
    setThreeDVisualMap,
    setVisualMap,
    initializeTheme,
    setTheme,
    toggleDark,
  }
}
