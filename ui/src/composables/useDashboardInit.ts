import { watch } from 'vue'
import { useDataPoint } from './useDataPoint'
import { useSettingsStore } from './useSettingsStore'
import { useUrlRouter } from './useUrlRouter'

// Owns the init-orchestration watchers and document-title side effect,
// keeping Dashboard.vue focused on wiring and layout only.
export function useDashboardInit() {
  const { dataSets, activeDataSet } = useDataPoint()
  const { initializeFromDataSet } = useSettingsStore()
  const { initFromUrl } = useUrlRouter()

  let urlInitialized = false

  watch(
    activeDataSet,
    (d) => {
      if (d?.name) document.title = `Vizb | ${d.name}`
      if (d?.settings) initializeFromDataSet(d.settings)
    },
    { immediate: true }
  )

  watch(
    dataSets,
    (d) => {
      if (d.length && !urlInitialized) {
        initFromUrl()
        urlInitialized = true
      }
    },
    { immediate: true }
  )
}
