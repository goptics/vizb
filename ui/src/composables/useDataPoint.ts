import { ref, shallowRef, markRaw, reactive, computed, nextTick, watch } from 'vue'
import type { Dataset, ChartType } from '../types'
import type { Arrangement } from './useChartPipeline'
import { filterDatasetSettings } from '../lib/filterDatasetSettings'
import {
  resetColor,
  isValidIndex,
  datasetDimension,
  isValueMode as checkValueMode,
  isMixedMode as checkMixedMode,
} from '../lib/utils'
import { presentAxisString } from '../lib/swap'
import { useSettingsStore } from './useSettingsStore'
import { classifyPayload, fetchDatasetDetail, type DataPayload } from '../lib/remoteData'
import { extractPathDatasetId } from '../lib/pathRoute'

const dataUrl = window.VIZB_DATA_URL
const pathname = window.location.pathname
const pathDatasetId =
  dataUrl && window.location.protocol !== 'file:' && !pathname.endsWith('/')
    ? extractPathDatasetId(pathname)
    : null
const getDatasets = async (): Promise<DataPayload> => {
  if (dataUrl && pathDatasetId) {
    return { mode: 'full', datasets: [await fetchDatasetDetail(dataUrl, pathDatasetId)] }
  }

  if (dataUrl) {
    const res = await fetch(dataUrl)
    if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
    return classifyPayload(await res.json())
  }

  if (import.meta.env.DEV) {
    const data = await import('../data/sample.json')
    // sample.json is always a full-dataset array in dev fixtures.
    return classifyPayload(data.default)
  }

  // Embedded HTML may inject one Dataset object or a Dataset[] (multi-tab).
  // classifyPayload normalizes both shapes to { mode: 'full', datasets: [...] }.
  return classifyPayload(window.VIZB_DATA ?? [])
}

// Global state. shallowRef (not ref): the rows are display-only and never mutated
// in place, so deep reactivity would only proxy every row for nothing — and that
// proxy is what forced the expensive JSON round-trip when cloning into the worker.
// Top-level `.value =` still triggers reactivity (the selector/dimension/arrangement
// computeds depend on the ref + activeDatasetId, not per-row reactivity).
const datasets = shallowRef<Dataset[]>([])
const activeDatasetId = ref(0)
const activeGroupId = ref(0)
const loading = ref(true)
const loadError = ref<string | null>(null)
const lazyCatalog = ref(false)
const detailLoading = ref(false)
const detailError = ref<string | null>(null)
const preparedDetails = new Map<string, Dataset>()

const prepareDataset = (ds: Dataset): Dataset => {
  const filtered = filterDatasetSettings(ds, window.VIZB_CHARTS)
  const axes = filtered.axes?.map((a) => ({
    key: a.key,
    label: a.label,
    type: a.type,
  }))
  return reactive({
    ...filtered,
    settings: filtered.settings ?? [],
    data: markRaw(filtered.data ?? []),
    ...(axes?.length ? { axes: markRaw(axes) } : {}),
  })
}

// Group list now comes from the worker (it owns grouping). Dashboard wires the
// pipeline's `groupNames` into this via `setGroupNames` on every `ready`, so the
// selector + URL router read the worker's grouping without any main-thread pass.
const groupNames = ref<string[]>([])
const setGroupNames = (names: string[]) => {
  groupNames.value = names
}

// Per-(dataset, chartType) target arrangement (e.g. "yx"); absent = identity.
// Key: "${datasetId}:${chartType}". Allows each chart type to keep its own swap.
const arrangementMap = reactive(new Map<string, string>())
const setArrangement = (datasetId: number, ct: ChartType, targetString: string) => {
  arrangementMap.set(`${datasetId}:${ct}`, targetString)
}

const loadDatasets = async () => {
  loading.value = true
  loadError.value = null
  try {
    const payload = await getDatasets()
    // Each Dataset is wrapped in reactive() so settings mutations (sort/scale/
    // showLabels/threeDRotate/swap) propagate to the chart pipeline's watchers.
    // The `data` field (raw rows) is markRaw'd so it stays proxy-free: the
    // transform worker clones it natively via postMessage structured clone,
    // which would otherwise reject Vue's reactive Proxy. Rows are display-only
    // and never mutated in place, so dropping per-row reactivity is the
    // intended perf trade-off.
    if (payload.mode === 'catalog') {
      lazyCatalog.value = true
      datasets.value = payload.entries.map((entry) =>
        prepareDataset({ ...entry, data: [], settings: [] })
      )
    } else {
      lazyCatalog.value = false
      datasets.value = payload.datasets.map(prepareDataset)
    }
  } catch (err: unknown) {
    loadError.value = err instanceof Error ? err.message : String(err)
  } finally {
    loading.value = false
  }
}

void loadDatasets()

// Values are normalized once at load (see `normalize`); this just guards the
// array shape.
const datasetsProcessed = computed<Dataset[]>(() => {
  if (!Array.isArray(datasets.value)) {
    datasets.value = [datasets.value]
  }

  return datasets.value
})

// Data-shape dimensionality tag for the active dataset, used by the settings
// panel to filter fields (e.g. `threeDRotate` is 3D-only). `undefined` for
// empty/unknown data — the panel treats that as "no dimension constraint".
const activeDataDimension = computed(() =>
  datasetDimension(datasetsProcessed.value[activeDatasetId.value]?.data)
)

const activeDataset = computed(
  () => datasetsProcessed.value[activeDatasetId.value] || datasetsProcessed.value[0]
)

export { activeDataset }

const { chartType, initializeTheme } = useSettingsStore()

export type ChartMode = 'grouped' | 'value' | 'mixed'

const chartMode = computed<ChartMode>(() => {
  const axes = activeDataset.value?.axes
  if (!axes?.length) return 'grouped'
  if (checkValueMode(axes)) return 'value'
  if (checkMixedMode(axes)) return 'mixed'
  return 'grouped'
})

const isValueMode = computed(() => chartMode.value === 'value')
const isMixedMode = computed(() => chartMode.value === 'mixed')
const isValueModeDataset = computed(() => chartMode.value === 'value')
const isMixedModeDataset = computed(() => chartMode.value === 'mixed')

// Derive identity from axes[] key order if present, else fall back to data shape.
// axes[] preserves the serial dimension order from --group-pattern / --group-regex.
const identityFromDataset = (ds: Dataset | undefined): string => {
  if (ds?.axes?.length) {
    return ds.axes
      .filter((a) => a.key !== 'metric')
      .map((a) => (a.key === 'name' ? 'n' : a.key.charAt(0)))
      .join('')
  }
  return presentAxisString(ds?.data)
}

// The active arrangement: per-(dataset, chartType) target with identity fallback.
const activeArrangement = computed<Arrangement>(() => {
  const ds = activeDataset.value
  const identityString = identityFromDataset(ds)
  const targetString =
    arrangementMap.get(`${activeDatasetId.value}:${chartType.value}`) ?? identityString
  return { identityString, targetString }
})

// Group list as the selector consumes it: a `{ name }[]` over the worker's names.
const resultGroups = computed(() => groupNames.value.map((name) => ({ name })))

watch(
  () => activeDataset.value?.theme,
  (theme) => {
    initializeTheme(theme)
  },
  { immediate: true }
)

const selectDataset = async (id: number): Promise<boolean> => {
  if (!isValidIndex(id, datasets.value.length)) return false

  activeDatasetId.value = id
  activeGroupId.value = 0
  detailError.value = null

  if (!lazyCatalog.value) {
    nextTick(() => resetColor())
    return true
  }

  const datasetId = datasets.value[id]?.id
  if (!datasetId || !dataUrl) return false

  let detail = preparedDetails.get(datasetId)
  if (detail) {
    const next = [...datasets.value]
    next[id] = detail
    datasets.value = next
    detailLoading.value = false
    nextTick(() => resetColor())
    return true
  }

  detailLoading.value = true
  try {
    detail = prepareDataset(await fetchDatasetDetail(dataUrl, datasetId))
  } catch (error: unknown) {
    if (activeDatasetId.value !== id) return false
    detailError.value = error instanceof Error ? error.message : String(error)
    detailLoading.value = false
    return false
  }

  if (activeDatasetId.value !== id) return false
  preparedDetails.set(datasetId, detail)

  const next = [...datasets.value]
  next[id] = detail
  datasets.value = next
  detailLoading.value = false
  nextTick(() => resetColor())
  return true
}

const retryActiveDataset = () =>
  loadError.value ? loadDatasets() : selectDataset(activeDatasetId.value)

const selectGroup = (id: number) => {
  if (isValidIndex(id, groupNames.value.length)) {
    activeGroupId.value = id
  }
}

export function useDataPoint() {
  const getArrangement = (datasetId: number, ct: ChartType): string | undefined => {
    return arrangementMap.get(`${datasetId}:${ct}`)
  }

  return {
    datasets,
    activeDataset,
    activeDatasetId,
    activeDataDimension,
    selectDataset,

    activeArrangement,
    setArrangement,
    getArrangement,
    arrangementMap,

    resultGroups,
    activeGroupId,
    selectGroup,
    setGroupNames,

    loading,
    loadError,
    lazyCatalog,
    detailLoading,
    detailError,
    retryActiveDataset,
    pathDatasetId,

    chartMode,
    isValueMode,
    isValueModeDataset,
    isMixedMode,
    isMixedModeDataset,
  }
}
