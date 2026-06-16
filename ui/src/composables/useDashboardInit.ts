import { watch } from 'vue'
import { useDataPoint } from './useDataPoint'
import { useUrlRouter } from './useUrlRouter'

// Owns the init-orchestration watchers and document-title side effect,
// keeping Dashboard.vue focused on wiring and layout only. With the new
// per-chart-config store, the dataset itself is the source of truth — there is
// no flat settings shape to seed at startup, so the only remaining init step
// is restoring state from the URL on first load.
export function useDashboardInit() {
  const { dataSets, activeDataSet } = useDataPoint()
  const { initFromUrl } = useUrlRouter()

  let urlInitialized = false

  watch(
    activeDataSet,
    (d) => {
      if (d?.name) document.title = `Vizb | ${d.name}`
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
