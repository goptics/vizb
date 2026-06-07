import { watch } from 'vue'
import { useDataPoint } from './useDataPoint'
import { useSettingsStore } from './useSettingsStore'
import { useUrlRouter } from './useUrlRouter'

// Owns the init-orchestration watchers and document-title side effect,
// keeping Dashboard.vue focused on wiring and layout only.
export function useDashboardInit() {
  const { benchmarks, activeDataSet } = useDataPoint()
  const { initializeFromDataSet } = useSettingsStore()
  const { initFromUrl } = useUrlRouter()

  let urlInitialized = false

  watch(
    activeDataSet,
    (b) => {
      if (b?.name) document.title = `Vizb | ${b.name}`
      if (b?.settings) initializeFromDataSet(b.settings)
    },
    { immediate: true }
  )

  watch(
    benchmarks,
    (b) => {
      if (b.length && !urlInitialized) {
        initFromUrl()
        urlInitialized = true
      }
    },
    { immediate: true }
  )
}
