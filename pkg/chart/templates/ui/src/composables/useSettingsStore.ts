import { ref } from 'vue'
import { useDark, useToggle } from '@vueuse/core'
import type { SortOrder, Settings, ChartType } from '../types/benchmark'

const sortOrder = ref<SortOrder>('')
const showLabels = ref(false)
const chartType = ref<ChartType>('bar')
let initialized = false

const isDark = useDark({
  selector: 'html',
  attribute: 'class',
  valueDark: 'dark',
  valueLight: 'light',
})
const toggleDark = useToggle(isDark)

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
