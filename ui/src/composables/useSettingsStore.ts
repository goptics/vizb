import { computed, ref, watch } from 'vue'
import type { ChartConfig, ChartType, Sort, ScaleInput, ScaleType } from '../types'
import { activeDataset } from './useDataPoint'
import { isValidIndex } from '../lib/utils'
import { applyScaleType, parseScale } from '../lib/scale'
import { activeThemeName, applyTheme, isAvailableThemeName, normalizeTheme } from '../lib/themes'

// Module-level singleton state. `activeChartIndex` is the cursor into
// `activeDataset.value.settings`; the dataset IS the source of truth — no flat
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

// Apply a stored theme name only when it is in the available set for the
// current report (at module load that is only UI default until dataset themes
// are registered). Unknown legacy names and custom hex strings are ignored.
if (isBrowser) {
  const savedTheme = localStorage.getItem('color-theme')
  if (savedTheme !== null && isAvailableThemeName(savedTheme)) {
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

/**
 * Pick the active theme for the current dataset.
 * Prefer localStorage `color-theme` when that name is still in the available
 * set for this report; otherwise use the author default (`themes[0]` / default
 * / legacy hex) passed by the caller.
 */
const initializeTheme = (datasetTheme?: string) => {
  if (isBrowser) {
    const saved = localStorage.getItem('color-theme')
    if (saved !== null && isAvailableThemeName(saved)) {
      applyTheme(saved)
      return
    }
  }
  applyTheme(datasetTheme)
}

/** Apply a theme only when its name is in the available set; persist the name. */
const setTheme = (theme: string) => {
  if (!isAvailableThemeName(theme)) return
  const normalized = normalizeTheme(theme)
  applyTheme(normalized)
  if (isBrowser) localStorage.setItem('color-theme', normalized)
}

export function useSettingsStore() {
  const activeConfig = computed<ChartConfig | undefined>(
    () => activeDataset.value?.settings[activeChartIndex.value]
  )

  const chartType = computed<ChartType>(() => activeConfig.value?.type ?? 'bar')

  // Clamp the active index when the active dataset's chart list shrinks (e.g.
  // a settings change drops one of the bundled charts).
  watch(
    () => activeDataset.value?.settings.length,
    (len) => {
      if (len !== undefined && activeChartIndex.value >= len) {
        activeChartIndex.value = 0
      }
    }
  )

  const setActiveChartIndex = (index: number) => {
    const len = activeDataset.value?.settings.length ?? 0
    if (isValidIndex(index, len)) {
      activeChartIndex.value = index
    }
  }

  // setChartType locates the first config with the requested type and makes it
  // active. No-op if the active dataset has no config of that type.
  const setChartType = (type: ChartType) => {
    const settings: ChartConfig[] = activeDataset.value?.settings ?? []
    const idx = settings.findIndex((s: ChartConfig) => s.type === type)
    if (idx !== -1) {
      activeChartIndex.value = idx
    }
  }

  // Per-field setters write the active config in place via `patchActive`.
  // Each config shape carries only the fields that apply to its chart type, so
  // `scale` and `threeDRotate` are only set on bar/line configs (the others
  // have no such field). The narrowing uses TypeScript's optional-field
  // semantics — no runtime `'field' in cfg` guard. SettingsPanel already gates
  // by `appliesTo` in `fieldRegistry`, so these setters are only called for
  // chart types that carry the field. The Go migration does NOT pre-populate
  // `threeDRotate` (it didn't exist in v0.12.0), so a `'threeDRotate' in cfg`
  // guard would silently no-op the first toggle on a freshly migrated config.
  const patchActive = (fields: Record<string, unknown>, when?: (cfg: ChartConfig) => boolean) => {
    const cfg = activeConfig.value
    if (!cfg || (when && !when(cfg))) return
    Object.assign(cfg, fields)
  }

  const currentScale = () => (activeConfig.value as { scale?: ScaleInput } | undefined)?.scale

  const setSort = (sort: Sort) => patchActive({ sort: { ...sort } })
  const setScale = (scale: ScaleType) => {
    const current = currentScale()
    if (parseScale(current).type === scale) return
    patchActive({ scale: applyScaleType(current, scale) })
  }
  const setStack = (stack: boolean) =>
    patchActive(stack ? { stack, scale: applyScaleType(currentScale(), 'linear') } : { stack })
  const setShowLabels = (show: boolean) => patchActive({ showLabels: show })
  const setSmooth = (smooth: boolean) => patchActive({ smooth }, (cfg) => cfg.type === 'line')
  const setHorizontal = (horizontal: boolean) =>
    patchActive({ horizontal }, (cfg) => cfg.type === 'bar')
  const setThreeDRotate = (rotate: boolean) => patchActive({ threeDRotate: rotate })
  const setSwap = (swap: string | undefined) => patchActive({ swap })
  const setThreeD = (enabled: boolean) => patchActive({ threeD: enabled })
  const setThreeDVisualMap = (enabled: boolean) => patchActive({ threeDVisualMap: enabled })
  const setVisualMap = (enabled: boolean) =>
    patchActive({ visualMap: enabled }, (cfg) => cfg.type === 'scatter')

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
