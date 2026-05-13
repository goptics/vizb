import { ref, computed } from 'vue'
import type { Benchmark, BenchmarkData } from '../types'
import { resetColor } from '../lib/utils'
import { useSettingsStore } from './useSettingsStore'
import { DEFAULT_SETTINGS } from './constants'

const getStatDimensions = (benchmarks: BenchmarkData[]) => {
  let dimension = 0
  if (benchmarks.some((b) => b.name)) dimension++
  if (benchmarks.some((b) => b.xAxis)) dimension++
  if (benchmarks.some((b) => b.yAxis)) dimension++
  return dimension
}

const getBenchmarks = async (): Promise<Benchmark[]> => {
  if (import.meta.env.DEV) {
    const data = await import('../data/sample.json')
    return data.default as unknown as Benchmark[]
  }

return window.VIZB_DATA ?? []
}

// Global state
const benchmarks = ref<Benchmark[]>([])
const activeBenchmarkId = ref(0)
const activeGroupId = ref(0)

// Load data immediately
getBenchmarks().then((data) => {
  benchmarks.value = data
})

// Process and group all benchmarks
const benchmarksProcessed = computed<Benchmark[]>(() => {
  if (!Array.isArray(benchmarks.value)) {
    benchmarks.value = [benchmarks.value]
  }

  if (!benchmarks.value.length) return []
  
  return benchmarks.value.map((benchmark) => {
    for (const result of benchmark.data) {
      for (const stat of result.stats) {
        const { value = 0 } = stat
        stat.value = Number(value.toFixed(2))
      }
    }
    return benchmark
  })
})

const activeBenchmarkDimension = computed(() =>
  getStatDimensions(benchmarksProcessed.value[activeBenchmarkId.value]?.data ?? [])
)

const activeBenchmark = computed(
  () => benchmarksProcessed.value[activeBenchmarkId.value] || benchmarksProcessed.value[0]
)

const grouped = computed(() => {
  if (!activeBenchmark.value) return new Map<string, BenchmarkData[]>()
  const groupMap = new Map<string, BenchmarkData[]>()
  for (const benchmarkData of activeBenchmark.value.data) {
    const name = benchmarkData.name || 'Default'
    if (!groupMap.has(name)) {
      groupMap.set(name, [])
    }
    groupMap.get(name)!.push(benchmarkData)
  }
  return groupMap
})

const benchmarkGroups = computed(() =>
  Array.from(grouped.value.entries()).map(([name, data]) => ({
    name,
    description: activeBenchmark.value?.description || '',
    cpu: {
      name: activeBenchmark.value?.cpu?.name || '',
      cores: activeBenchmark.value?.cpu?.cores || 0,
    },
    settings: activeBenchmark.value?.settings || DEFAULT_SETTINGS,
    data,
  }))
)

const activeGroup = computed(
  () => benchmarkGroups.value[activeGroupId.value] || benchmarkGroups.value[0]
)

const { initializeFromBenchmark } = useSettingsStore()

const selectBenchmark = (id: number) => {
  if (id >= 0 && id < benchmarks.value.length) {
    resetColor()
    const currentGroupName = activeGroup.value?.name
    activeBenchmarkId.value = id
    const benchmark = benchmarks.value[id]
    if (benchmark?.settings) {
      initializeFromBenchmark(benchmark.settings, true)
    }
    const newGroupIndex = benchmarkGroups.value.findIndex((g) => g.name === currentGroupName)
    if (newGroupIndex !== -1) {
      activeGroupId.value = newGroupIndex
    } else {
      activeGroupId.value = 0
    }
  }
}

const selectGroup = (id: number) => {
  if (id >= 0 && id < benchmarkGroups.value.length) {
    activeGroupId.value = id
  }
}

export function useBenchmarkData() {
  return {
    benchmarks,
    activeBenchmark,
    activeBenchmarkId,
    activeBenchmarkDimension,
    selectBenchmark,
    resultGroups: benchmarkGroups,
    activeGroup,
    activeGroupId,
    selectGroup,
  }
}
