import { ref } from 'vue'
import type { SortOrder, Settings, ChartType } from '../types/benchmark'

const sortOrder = ref<SortOrder>('')
const showLabels = ref(false)
const chartType = ref<ChartType>('bar')
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
  const setSortOrder = (order: SortOrder) => {
    sortOrder.value = order
  }

  const setShowLabels = (show: boolean) => {
    showLabels.value = show
  }

  const setChartType = (type: ChartType) => {
    chartType.value = type
  }

  const initializeFromBenchmark = (settings: Settings) => {
    if (!initialized) {
      sortOrder.value = settings.sort
      showLabels.value = settings.showLabels
      initialized = true
    }
  }

  return {
    sortOrder,
    showLabels,
    chartType,
    isDark,
    setSortOrder,
    setShowLabels,
    setChartType,
    toggleDark,
    initializeFromBenchmark
  }
}
