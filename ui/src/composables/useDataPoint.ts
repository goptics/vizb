import { ref, shallowRef, markRaw, reactive, computed, nextTick } from 'vue'
import type { DataSet, DataPoint, ChartType } from '../types'
import type { Arrangement } from './useChartPipeline'
import { filterDataSetSettings } from '../lib/filterDataSetSettings'
import {
  resetColor,
  isValidIndex,
  datasetDimension,
  isValueMode as checkValueMode,
  isScatterTransformMode,
} from '../lib/utils'
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
// (a few `.some()` passes, no grouping). Mirrors SwapControl's identity check.
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
    // Each DataSet is wrapped in reactive() so settings mutations (sort/scale/
    // showLabels/threeDRotate/swap) propagate to the chart pipeline's watchers.
    // The `data` field (raw rows) is markRaw'd so it stays proxy-free: the
    // transform worker clones it natively via postMessage structured clone,
    // which would otherwise reject Vue's reactive Proxy. Rows are display-only
    // and never mutated in place, so dropping per-row reactivity is the
    // intended perf trade-off.
    const raw = Array.isArray(data) ? data : [data]
    const allowed = window.VIZB_CHARTS
    dataSets.value = raw.map((ds) => {
      const filtered = filterDataSetSettings(ds, allowed)
      // data + axes are markRaw'd so postMessage structured-clone succeeds (Vue
      // reactive Proxies on the worker init payload throw DataCloneError).
      const axes = filtered.axes?.map((a) => ({
        key: a.key,
        label: a.label,
        type: a.type,
      }))
      return reactive({
        ...filtered,
        data: markRaw(filtered.data),
        ...(axes?.length ? { axes: markRaw(axes) } : {}),
      })
    })
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

// Number of present axes in the active dataset (0..4: name, x, y, z). Used
// for the "rows: N cols: M" subtitle in the dashboard.
const activeDataAxisCount = computed(() =>
  getStatDimensions(dataSetsProcessed.value[activeDataSetId.value]?.data ?? [])
)

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

const { chartType } = useSettingsStore()

// True when the active dataset uses pure value-mode axes (--axes x,y[,z]).
// Used to disable sort/swap controls that are no-ops in value mode.
const isValueModeActive = computed(() => checkValueMode(activeDataSet.value?.axes))

// Scatter-only: value or hybrid transform paths are active for this dataset.
const isValueModeDataset = computed(() =>
  isScatterTransformMode(chartType.value, activeDataSet.value?.axes)
)

// Derive identity from axes[] key order if present, else fall back to presentKeys(data).
// axes[] preserves the serial dimension order from --group-pattern / --group-regex.
const identityFromDataSet = (ds: DataSet | undefined): string => {
  if (ds?.axes?.length) {
    return ds.axes.map((a) => (a.key === 'name' ? 'n' : a.key.charAt(0))).join('')
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

    // The new store reads `dataset.value.settings[activeChartIndex]` directly,
    // so no init step is needed — switching the active dataset id is enough.

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
    activeDataAxisCount,
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

    isValueMode: isValueModeActive,
    isValueModeDataset,
  }
}
