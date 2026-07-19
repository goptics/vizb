import { ref, shallowRef, markRaw, reactive, computed, nextTick, watch } from 'vue'
import type { DataSet, ChartType } from '../types'
import type { Arrangement } from './useChartPipeline'
import { filterDataSetSettings } from '../lib/filterDataSetSettings'
import {
  resetColor,
  isValidIndex,
  datasetDimension,
  isValueMode as checkValueMode,
  isMixedMode as checkMixedMode,
} from '../lib/utils'
import { presentAxisString } from '../lib/swap'
import { useSettingsStore } from './useSettingsStore'
import { classifyRemotePayload, fetchDatasetDetail, type RemotePayload } from '../lib/remoteData'

const dataUrl = window.VIZB_DATA_URL
const getDataSets = async (): Promise<RemotePayload> => {
  const url = window.VIZB_DATA_URL
  if (url) {
    const res = await fetch(url)
    if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
    return classifyRemotePayload(await res.json())
  }

  if (import.meta.env.DEV) {
    const data = await import('../data/sample.json')
    return { mode: 'full', datasets: data.default as unknown as DataSet[] }
  }

  return { mode: 'full', datasets: window.VIZB_DATA ?? [] }
}

// Global state. shallowRef (not ref): the rows are display-only and never mutated
// in place, so deep reactivity would only proxy every row for nothing — and that
// proxy is what forced the expensive JSON round-trip when cloning into the worker.
// Top-level `.value =` still triggers reactivity (the selector/dimension/arrangement
// computeds depend on the ref + activeDataSetId, not per-row reactivity).
const dataSets = shallowRef<DataSet[]>([])
const activeDataSetId = ref(0)
const activeGroupId = ref(0)
const loading = ref(true)
const loadError = ref<string | null>(null)
const lazyCatalog = ref(false)
const detailLoading = ref(false)
const detailError = ref<string | null>(null)
const preparedDetails = new Map<string, DataSet>()

const prepareDataSet = (ds: DataSet): DataSet => {
  const filtered = filterDataSetSettings(ds, window.VIZB_CHARTS)
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

getDataSets()
  .then((payload) => {
    // Each DataSet is wrapped in reactive() so settings mutations (sort/scale/
    // showLabels/threeDRotate/swap) propagate to the chart pipeline's watchers.
    // The `data` field (raw rows) is markRaw'd so it stays proxy-free: the
    // transform worker clones it natively via postMessage structured clone,
    // which would otherwise reject Vue's reactive Proxy. Rows are display-only
    // and never mutated in place, so dropping per-row reactivity is the
    // intended perf trade-off.
    if (payload.mode === 'catalog') {
      lazyCatalog.value = true
      dataSets.value = payload.entries.map((entry) =>
        prepareDataSet({ ...entry, data: [], settings: [] })
      )
    } else {
      dataSets.value = payload.datasets.map(prepareDataSet)
    }
  })
  .catch((err: unknown) => {
    loadError.value = err instanceof Error ? err.message : String(err)
  })
  .finally(() => {
    loading.value = false
  })

// Values are normalized once at load (see `normalize`); this just guards the
// array shape.
const dataSetsProcessed = computed<DataSet[]>(() => {
  if (!Array.isArray(dataSets.value)) {
    dataSets.value = [dataSets.value]
  }

  return dataSets.value
})

// Data-shape dimensionality tag for the active dataset, used by the settings
// panel to filter fields (e.g. `threeDRotate` is 3D-only). `undefined` for
// empty/unknown data — the panel treats that as "no dimension constraint".
const activeDataDimension = computed(() =>
  datasetDimension(dataSetsProcessed.value[activeDataSetId.value]?.data)
)

const activeDataSet = computed(
  () => dataSetsProcessed.value[activeDataSetId.value] || dataSetsProcessed.value[0]
)

export { activeDataSet }

const { chartType, initializeTheme } = useSettingsStore()

export type ChartMode = 'grouped' | 'value' | 'mixed'

const chartMode = computed<ChartMode>(() => {
  const axes = activeDataSet.value?.axes
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
const identityFromDataSet = (ds: DataSet | undefined): string => {
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
  const ds = activeDataSet.value
  const identityString = identityFromDataSet(ds)
  const targetString =
    arrangementMap.get(`${activeDataSetId.value}:${chartType.value}`) ?? identityString
  return { identityString, targetString }
})

// Group list as the selector consumes it: a `{ name }[]` over the worker's names.
const resultGroups = computed(() => groupNames.value.map((name) => ({ name })))

watch(
  () => activeDataSet.value?.theme,
  (theme) => {
    initializeTheme(theme)
  },
  { immediate: true }
)

const selectDataSet = async (id: number): Promise<boolean> => {
  if (!isValidIndex(id, dataSets.value.length)) return false

  activeDataSetId.value = id
  activeGroupId.value = 0
  detailError.value = null

  if (!lazyCatalog.value) {
    nextTick(() => resetColor())
    return true
  }

  const dataSetId = dataSets.value[id]?.id
  if (!dataSetId || !dataUrl) return false

  let detail = preparedDetails.get(dataSetId)
  if (detail) {
    const next = [...dataSets.value]
    next[id] = detail
    dataSets.value = next
    detailLoading.value = false
    nextTick(() => resetColor())
    return true
  }

  detailLoading.value = true
  try {
    detail = prepareDataSet(await fetchDatasetDetail(dataUrl, dataSetId))
  } catch (error: unknown) {
    if (activeDataSetId.value !== id) return false
    detailError.value = error instanceof Error ? error.message : String(error)
    detailLoading.value = false
    return false
  }

  if (activeDataSetId.value !== id) return false
  preparedDetails.set(dataSetId, detail)

  const next = [...dataSets.value]
  next[id] = detail
  dataSets.value = next
  detailLoading.value = false
  nextTick(() => resetColor())
  return true
}

const retryActiveDataSet = () => selectDataSet(activeDataSetId.value)

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
    dataSets,
    activeDataSet,
    activeDataSetId,
    activeDataDimension,
    selectDataSet,

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
    retryActiveDataSet,

    chartMode,
    isValueMode,
    isValueModeDataset,
    isMixedMode,
    isMixedModeDataset,
  }
}
