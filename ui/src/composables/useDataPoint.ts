import { ref, computed, nextTick } from 'vue'
import type { DataSet, DataPoint } from '../types'
import { resetColor, isValidIndex } from '../lib/utils'
import { useSettingsStore } from './useSettingsStore'
import { DEFAULT_SETTINGS } from './constants'

const getStatDimensions = (points: DataPoint[]) => {
  let dimension = 0

  if (points.some((b) => b.name)) dimension++
  if (points.some((b) => b.xAxis)) dimension++
  if (points.some((b) => b.yAxis)) dimension++
  if (points.some((b) => b.zAxis)) dimension++

  return dimension
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

// Normalize stat values once, at load. Done here (not in a computed) so later
// swaps — which mutate the dataset in place — don't re-round every row of every
// dataset on each change.
const normalize = (sets: DataSet[]): DataSet[] => {
  for (const set of sets) {
    for (const result of set.data ?? []) {
      for (const stat of result.stats ?? []) {
        stat.value = Number((stat.value ?? 0).toFixed(2))
      }
    }
  }
  return sets
}

// Global state
const dataSets = ref<DataSet[]>([])
const activeDataSetId = ref(0)
const activeGroupId = ref(0)
const loading = ref(true)
const loadError = ref<string | null>(null)

getDataSets()
  .then((data) => {
    dataSets.value = normalize(Array.isArray(data) ? data : [data])
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

// Group results within the active benchmark
const grouped = computed(() => {
  if (!activeDataSet.value) return new Map<string, DataPoint[]>()

  const groupMap = new Map<string, DataPoint[]>()

  for (const dataPoints of activeDataSet.value.data) {
    const { name = 'Default', ...rest } = dataPoints

    if (!groupMap.has(name)) {
      groupMap.set(name, [])
    }

    groupMap.get(name)!.push(rest)
  }

  return groupMap
})

// Convert grouped results to array format for easier consumption
const dataPointGroups = computed(() =>
  Array.from(grouped.value.entries()).map(([name, data]) => ({
    name,
    description: activeDataSet.value?.description || '',
    cpu: {
      name: activeDataSet.value?.cpu?.name || '',
      cores: activeDataSet.value?.cpu?.cores || 0,
    },
    settings: activeDataSet.value?.settings || DEFAULT_SETTINGS,
    data,
  }))
)

const activeGroup = computed(
  () => dataPointGroups.value[activeGroupId.value] || dataPointGroups.value[0]
)

const { initializeFromDataSet } = useSettingsStore()

const selectDataSet = (id: number) => {
  if (isValidIndex(id, dataSets.value.length)) {
    const currentGroupName = activeGroup.value?.name

    activeDataSetId.value = id

    const benchmark = dataSets.value[id]
    if (benchmark?.settings) {
      initializeFromDataSet(benchmark.settings, true)
    }

    const newGroupIndex = dataPointGroups.value.findIndex((g) => g.name === currentGroupName)

    if (newGroupIndex !== -1) {
      activeGroupId.value = newGroupIndex
    } else {
      activeGroupId.value = 0
    }

    nextTick(() => resetColor())
  }
}

const selectGroup = (id: number) => {
  if (isValidIndex(id, dataPointGroups.value.length)) {
    activeGroupId.value = id
  }
}

export function useDataPoint() {
  return {
    dataSets,
    activeDataSet,
    activeDataSetId,
    activeDataSetDimension,
    selectDataSet,

    resultGroups: dataPointGroups,
    activeGroup,
    activeGroupId,
    selectGroup,

    loading,
    loadError,
  }
}
