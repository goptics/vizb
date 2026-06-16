import { ref, shallowRef, markRaw, reactive, computed, nextTick } from 'vue'
import type { DataSet, DataPoint, ChartType } from '../types'
import type { Arrangement } from './useChartPipeline'
import { resetColor, isValidIndex } from '../lib/utils'
import { useSettingsStore } from './useSettingsStore'

const getStatDimensions = (points: DataPoint[]) => {
  let dimension = 0

  if (points.some((b) => b.name)) dimension++
  if (points.some((b) => b.xAxis)) dimension++
  if (points.some((b) => b.yAxis)) dimension++
  if (points.some((b) => b.zAxis)) dimension++

  return dimension
}

// Canonical axis order; the identity arrangement is the present source axes in
// this order (e.g. a dataset with name+xAxis → "nx").
const AXIS_ORDER = ['n', 'x', 'y', 'z'] as const

// The present source axes of a dataset, as a compact arrangement string. Cheap
// (a few `.some()` passes, no grouping). Mirrors AxisSwapper's identity check.
const presentKeys = (data: DataPoint[] | undefined): string => {
  if (!data?.length) return ''
  const fieldFor = { n: 'name', x: 'xAxis', y: 'yAxis', z: 'zAxis' } as const
  return AXIS_ORDER.filter((k) => data.some((d) => d[fieldFor[k]])).join('')
}

const getDataSets = async (): Promise<DataSet[]> => {
  if (import.meta.env.DEV) {
    const data = await import('../data/sample.json')
    return data.default as unknown as DataSet[]
  }

  const url = window.VIZB_DATA_URL
  if (url) {
    const res = await fetch(url)
    if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
    return res.json()
  }

  return window.VIZB_DATA ?? []
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
  .then((data) => {
    // markRaw keeps the rows plain (clone-safe) even if a future code path tries
    // to wrap them in reactive() — the worker postMessage clones them natively.
    dataSets.value = markRaw(Array.isArray(data) ? data : [data])
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

const activeDataSetDimension = computed(() =>
  getStatDimensions(dataSetsProcessed.value[activeDataSetId.value]?.data ?? [])
)

const activeDataSet = computed(
  () => dataSetsProcessed.value[activeDataSetId.value] || dataSetsProcessed.value[0]
)

export { activeDataSet }

const { initializeFromDataSet, chartType } = useSettingsStore()

// Derive identity from axes[] key order if present, else fall back to presentKeys(data).
// axes[] preserves the serial dimension order from --group-pattern / --group-regex.
const identityFromDataSet = (ds: DataSet | undefined): string => {
  if (ds?.settings?.axes?.length) {
    return ds.settings.axes.map((a) => (a.key === 'name' ? 'n' : a.key.charAt(0))).join('')
  }
  return presentKeys(ds?.data)
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

const selectDataSet = (id: number) => {
  if (isValidIndex(id, dataSets.value.length)) {
    activeDataSetId.value = id

    const benchmark = dataSets.value[id]
    if (benchmark?.settings) {
      initializeFromDataSet(benchmark.settings, true)
    }

    // Group names repopulate asynchronously from the worker's `ready` for the new
    // dataset; reset to the first group until they arrive.
    activeGroupId.value = 0

    nextTick(() => resetColor())
  }
}

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
    activeDataSetDimension,
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
  }
}
