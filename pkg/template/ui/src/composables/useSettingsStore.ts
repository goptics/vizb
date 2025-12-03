import { ref, computed } from 'vue'
import type { Sort, Settings, ChartType } from '../types'
import { DEFAULT_SETTINGS } from './constants'

const sortOrder = ref<Sort>({ enabled: false, order: 'asc' })
const showLabels = ref(false)
const charts = ref<ChartType[]>(DEFAULT_SETTINGS.charts)
const activeChartIndex = ref<number>(0)
const chartType = computed<ChartType>(() => charts.value[activeChartIndex.value] ?? 'bar')
const selectedSwapIndexMap = ref(new Map<number, number>())
let initialized = false

// Simple dark mode implementation
const isDark = ref(false)

// Initialize dark mode from localStorage or system preference
const initializeDarkMode = () => {
  const saved = localStorage.getItem('dark-mode')
  if (saved !== null) {
    isDark.value = saved === 'true'
  } else {
    // Check system preference
    isDark.value = window.matchMedia('(prefers-color-scheme: dark)').matches
  }
  updateHtmlClass()
}

// Update HTML class based on dark mode state
const updateHtmlClass = () => {
  const html = document.documentElement

  if (isDark.value) {
    html.classList.add('dark')
    html.classList.remove('light')
  } else {
    html.classList.add('light')
    html.classList.remove('dark')
  }
}

// Toggle dark mode
const toggleDark = () => {
  isDark.value = !isDark.value
  localStorage.setItem('dark-mode', isDark.value.toString())
  updateHtmlClass()
}

// Initialize on module load
initializeDarkMode()

export function useSettingsStore() {
  const setSort = (sort: Sort) => {
    sortOrder.value = sort
  }

  const setShowLabels = (show: boolean) => {
    showLabels.value = show
  }

  const setCharts = (list: ChartType[]) => {
    const filtered = list.filter((c) => DEFAULT_SETTINGS.charts.includes(c))
    charts.value = filtered.length ? filtered : DEFAULT_SETTINGS.charts

    // clamp index
    if (activeChartIndex.value < 0 || activeChartIndex.value >= charts.value.length) {
      activeChartIndex.value = 0
    }
  }

  const setActiveChartIndex = (index: number) => {
    if (index >= 0 && index < charts.value.length) {
      activeChartIndex.value = index
    }
  }

  const setChartType = (type: ChartType) => {
    const idx = charts.value.indexOf(type)
    if (idx !== -1) {
      activeChartIndex.value = idx
    }
  }

  const initializeFromBenchmark = (settings: Settings, force = false) => {
    if (!initialized || force) {
      sortOrder.value = settings.sort
      showLabels.value = settings.showLabels

      setCharts(settings.charts ?? DEFAULT_SETTINGS.charts)
      setActiveChartIndex(0)
      initialized = true
    }
  }

  const setSelectedSwapIndex = (benchmarkId: number, index: number) => {
    selectedSwapIndexMap.value.set(benchmarkId, index)
  }

  const getSelectedSwapIndex = (benchmarkId: number): number | undefined => {
    return selectedSwapIndexMap.value.get(benchmarkId)
  }

  return {
    sortOrder,
    showLabels,
    charts,
    activeChartIndex,
    chartType,
    isDark,
    selectedSwapIndexMap,
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
