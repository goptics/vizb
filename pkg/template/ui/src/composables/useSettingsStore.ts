import { reactive, computed } from 'vue'
import type { Sort, ChartType, Settings as BenchmarkSettings } from '../types'
import { DEFAULT_SETTINGS } from './constants'

type StoreSettings = {
  sort: Sort
  showLabels: boolean
  charts: ChartType[]
  activeChartIndex: number
  selectedSwapIndexMap: Map<number, number>
  isDark: boolean
}

const settings = reactive<StoreSettings>({
  sort: { enabled: false, order: 'asc' },
  showLabels: false,
  charts: DEFAULT_SETTINGS.charts,
  activeChartIndex: 0,
  selectedSwapIndexMap: new Map(),
  isDark: false,
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

  const setShowLabels = (show: boolean) => {
    settings.showLabels = show
  }

  const setCharts = (list: ChartType[]) => {
    const filtered = list.filter((c) => DEFAULT_SETTINGS.charts.includes(c))
    settings.charts = filtered.length ? filtered : DEFAULT_SETTINGS.charts

    // clamp index
    if (settings.activeChartIndex < 0 || settings.activeChartIndex >= settings.charts.length) {
      settings.activeChartIndex = 0
    }
  }

  const setActiveChartIndex = (index: number) => {
    if (index >= 0 && index < settings.charts.length) {
      settings.activeChartIndex = index
    }
  }

  const setChartType = (type: ChartType) => {
    const idx = settings.charts.indexOf(type)
    if (idx !== -1) {
      settings.activeChartIndex = idx
    }
  }

  const initializeFromBenchmark = (inputSettings: BenchmarkSettings, force = false) => {
    if (!initialized || force) {
      settings.sort = inputSettings.sort
      settings.showLabels = inputSettings.showLabels

      setCharts(inputSettings.charts ?? DEFAULT_SETTINGS.charts)
      setActiveChartIndex(0)
      initialized = true
    }
  }

  const setSelectedSwapIndex = (benchmarkId: number, index: number) => {
    settings.selectedSwapIndexMap.set(benchmarkId, index)
  }

  const getSelectedSwapIndex = (benchmarkId: number): number | undefined => {
    return settings.selectedSwapIndexMap.get(benchmarkId)
  }

  return {
    settings,
    chartType,
    setSort,
    setShowLabels,
    setCharts,
    setActiveChartIndex,
    setChartType,
    toggleDark,
    initializeFromBenchmark,
    setSelectedSwapIndex,
    getSelectedSwapIndex,
  }
}
