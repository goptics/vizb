import { ref } from 'vue'
import { useDark, useToggle } from '@vueuse/core'
import type { SortOrder } from '../types/benchmark'

/**
 * Global settings store using Vue's reactivity
 * Manages sorting, labels visibility, and theme state
 */
const sortOrder = ref<SortOrder>('default')
const showLabels = ref(false)

// Configure dark mode to use 'class' strategy on html element
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

  const toggleLabels = () => {
    showLabels.value = !showLabels.value
  }

  const setShowLabels = (show: boolean) => {
    showLabels.value = show
  }

  return {
    sortOrder,
    showLabels,
    isDark,
    setSortOrder,
    toggleLabels,
    setShowLabels,
    toggleDark
  }
}
