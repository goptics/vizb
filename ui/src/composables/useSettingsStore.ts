import { reactive, computed } from 'vue'
import type {
  Sort,
  ChartType,
  ChartSettings,
  ScaleType,
  Settings as DataSetSettings,
} from '../types'
import { DEFAULT_SETTINGS } from './constants'
import { isValidIndex } from '../lib/utils'

type ResolvedKey = 'sort' | 'showLabels' | 'scale' | 'autoRotate'
type ResolvedValue = Sort | boolean | ScaleType

type StoreSettings = {
  sort: Sort
  showLabels: boolean
  charts: ChartType[]
  activeChartIndex: number
  // Key: "${benchmarkId}:${chartType}" — per-chart swap index
  selectedSwapIndexMap: Map<string, number>
  isDark: boolean
  scale: ScaleType
  autoRotate: boolean
  chartSettings: Partial<Record<ChartType, ChartSettings>>
}

const settings = reactive<StoreSettings>({
  sort: { enabled: false, order: 'asc' },
  showLabels: false,
  charts: DEFAULT_SETTINGS.charts,
  activeChartIndex: 0,
  selectedSwapIndexMap: new Map(),
  isDark: false,
  scale: 'linear',
  autoRotate: false,
  chartSettings: {},
})

const chartType = computed<ChartType>(() => settings.charts[settings.activeChartIndex] ?? 'bar')

let initialized = false

// Initialize dark mode from localStorage or system preference
const initializeDarkMode = () => {
  const saved = localStorage.getItem('dark-mode')

  if (saved !== null) {
    settings.isDark = saved === 'true'
  } else {
    // Check system preference
    settings.isDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  }

  updateHtmlClass()
}

// Update HTML class based on dark mode state
const updateHtmlClass = () => {
  const html = document.documentElement
  let [addClass, removeClass] = ['light', 'dark']

  if (settings.isDark) {
    ;[addClass, removeClass] = [removeClass, addClass]
  }

  html.classList.add(addClass)
  html.classList.remove(removeClass)
}

// Toggle dark mode
const toggleDark = () => {
  settings.isDark = !settings.isDark
  localStorage.setItem('dark-mode', settings.isDark.toString())
  updateHtmlClass()
}

// Initialize on module load
initializeDarkMode()

export function useSettingsStore() {
  const setSort = (sort: Sort) => {
    settings.sort = sort
  }

  const setScale = (scale: ScaleType) => {
    settings.scale = scale
  }

  const setShowLabels = (show: boolean) => {
    settings.showLabels = show
  }

  const setAutoRotate = (rotate: boolean) => {
    settings.autoRotate = rotate
  }

  const setCharts = (list: ChartType[]) => {
    // Constrain to the charts actually bundled at generation time (--charts).
    // In remote mode settings.charts comes from the fetched JSON, which may list
    // charts whose renderer chunks were pruned; surfacing those tabs would fail
    // the lazy import(). Fall back to all charts when VIZB_CHARTS is absent.
    const allowed = window.VIZB_CHARTS?.length ? window.VIZB_CHARTS : DEFAULT_SETTINGS.charts
    const filtered = list.filter((c) => allowed.includes(c))
    settings.charts = filtered.length ? filtered : allowed

    if (!isValidIndex(settings.activeChartIndex, settings.charts.length)) {
      settings.activeChartIndex = 0
    }
  }

  const setActiveChartIndex = (index: number) => {
    if (isValidIndex(index, settings.charts.length)) {
      settings.activeChartIndex = index
    }
  }

  const setChartType = (type: ChartType) => {
    const idx = settings.charts.indexOf(type)
    if (idx !== -1) {
      settings.activeChartIndex = idx
    }
  }

  const initializeFromDataSet = (inputSettings: DataSetSettings, force = false) => {
    if (!initialized || force) {
      settings.sort = inputSettings.sort
      settings.showLabels = inputSettings.showLabels
      settings.scale = inputSettings.scale || 'linear'
      // Seed per-chart overrides from the dataset's baked-in chartSettings.
      settings.chartSettings = inputSettings.chartSettings
        ? { ...inputSettings.chartSettings }
        : {}

      setCharts(inputSettings.charts ?? DEFAULT_SETTINGS.charts)
      setActiveChartIndex(0)
      initialized = true
    }
  }

  // resolved returns the effective value for key for the active chart type:
  // per-chart override (if set) or the global default.
  const resolved = (key: ResolvedKey): ResolvedValue => {
    const ct = chartType.value
    const perChart = settings.chartSettings[ct]
    if (perChart !== undefined && perChart[key] !== undefined) {
      return perChart[key] as ResolvedValue
    }
    return settings[key] as ResolvedValue
  }

  // setForActiveChart writes per-chart overrides for the currently active chart type.
  const setForActiveChart = (update: Partial<Omit<ChartSettings, 'swap'>>) => {
    const ct = chartType.value
    settings.chartSettings[ct] = { ...settings.chartSettings[ct], ...update }
  }

  // Swap index is keyed by (benchmarkId, chartType) so each chart keeps its own arrangement.
  const setSelectedSwapIndex = (benchmarkId: number, ct: ChartType, index: number) => {
    settings.selectedSwapIndexMap.set(`${benchmarkId}:${ct}`, index)
  }

  const getSelectedSwapIndex = (benchmarkId: number, ct: ChartType): number | undefined => {
    return settings.selectedSwapIndexMap.get(`${benchmarkId}:${ct}`)
  }

  return {
    settings,
    chartType,
    setSort,
    setScale,
    setShowLabels,
    setAutoRotate,
    setCharts,
    setActiveChartIndex,
    setChartType,
    toggleDark,
    initializeFromDataSet,
    resolved,
    setForActiveChart,
    setSelectedSwapIndex,
    getSelectedSwapIndex,
  }
}
