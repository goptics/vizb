import { ref, computed } from 'vue'
import type { Benchmark, BenchmarkResult } from '../types/benchmark'
import { resetColor } from '../lib/utils'


const getBenchmarks = async (): Promise<Benchmark[]> => {
  if (import.meta.env.DEV) {
    const data = await import('../data/sample.json')
    return data.default as Benchmark[]
  }
  
  try {
    const parsed = JSON.parse(window.VIZB_DATA)
    return parsed as Benchmark[]
  } catch (error) {
    console.error('Failed to parse VIZB_DATA:', error)
    return []
  }
}

/**
 * Composable for loading and managing benchmark data
 * Groups results by benchmark name to create separate benchmark groups
 */
export function useBenchmarkData() {
  const benchmarksData = ref<Benchmark[]>([])
  
  // Load data immediately
  getBenchmarks().then(data => {
    benchmarksData.value = data
  })

  // Process and group all benchmarks
  const benchmarks = computed<Benchmark[]>(() => {
    if (!benchmarksData.value.length) return []

    return benchmarksData.value.map(benchmark => {
      // Process stats for each result
      for (const result of benchmark.results) {
        for (const stat of result.stats) {
          stat.value = parseFloat(stat.value.toFixed(2))
        }
      }
      return benchmark
    })
  })

  // Track active benchmark selection
  const activeBenchmarkId = ref(0)
  const activeBenchmark = computed(() => benchmarks.value[activeBenchmarkId.value] || benchmarks.value[0])

  // Group results within the active benchmark
  const groupedResults = computed(() => {
    if (!activeBenchmark.value) return new Map<string, BenchmarkResult[]>()

    const groupMap = new Map<string, BenchmarkResult[]>()
    
    for (const result of activeBenchmark.value.results) {
      if (!groupMap.has(result.name)) {
        groupMap.set(result.name, [])
      }
      groupMap.get(result.name)!.push(result)
    }

    return groupMap
  })


  // Convert grouped results to array format for easier consumption
  const resultGroups = computed(() => {
    return Array.from(groupedResults.value.entries()).map(([name, results]) => ({
      name,
      results,
      description: activeBenchmark.value?.description || '',
      cpu: activeBenchmark.value?.cpu || '',
      settings: activeBenchmark.value?.settings || { sort: '', showLabels: false }
    }))
  })


  const activeGroupId = ref(0)
  const activeGroup = computed(() => resultGroups.value[activeGroupId.value] || resultGroups.value[0])

  const selectBenchmark = (id: number) => {
    if (id >= 0 && id < benchmarks.value.length) {
      // Reset color mapping when benchmark changes
      resetColor()
      
      activeBenchmarkId.value = id
      // Reset group selection when benchmark changes
      activeGroupId.value = 0
    }
  }


  const selectGroup = (id: number) => {
    if (id >= 0 && id < resultGroups.value.length) {
      activeGroupId.value = id
    }
  }

  return {
    // Top level benchmark selection
    benchmarks,
    activeBenchmark,
    activeBenchmarkId,
    selectBenchmark,

    // Inner level group selection
    resultGroups,
    activeGroup,
    activeGroupId,
    selectGroup
  }
}
