import { watch } from 'vue'
import { useBenchmarkData } from './useBenchmarkData'
import { useSettingsStore } from './useSettingsStore'
import { useUrlRouter } from './useUrlRouter'

// Owns the init-orchestration watchers and document-title side effect,
// keeping Dashboard.vue focused on wiring and layout only.
export function useDashboardInit() {
  const { benchmarks, activeBenchmark } = useBenchmarkData()
  const { initializeFromBenchmark } = useSettingsStore()
  const { initFromUrl } = useUrlRouter()

  let urlInitialized = false

  watch(
    activeBenchmark,
    (b) => {
      if (b?.name) document.title = `Vizb | ${b.name}`
      if (b?.settings) initializeFromBenchmark(b.settings)
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
