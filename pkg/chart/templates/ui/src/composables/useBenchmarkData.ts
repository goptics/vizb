import { ref, computed } from 'vue'
import type { Benchmark, BenchmarkResult } from '../types/benchmark'
import sampleData from '../data/sample.json'

/**
 * Composable for loading and managing benchmark data
 * Groups results by benchmark name to create separate benchmark groups
 */
export function useBenchmarkData() {
  const rawData = sampleData as Benchmark
  const loading = ref(false)
  const error = ref<Error | null>(null)

  /**
   * Group results by their name field to create separate benchmarks
   * E.g., StaticAll results become one benchmark, DynamicRoutes another
   */
  const benchmarks = computed<Benchmark[]>(() => {
    const groupMap = new Map<string, BenchmarkResult[]>()

    // Group results by name
    rawData.results.forEach(result => {
      if (!groupMap.has(result.name)) {
        groupMap.set(result.name, [])
      }
      groupMap.get(result.name)!.push(result)
    })

    // Convert groups to separate benchmarks
    return Array.from(groupMap.entries()).map(([name, results]) => ({
      name: name,
      description: rawData.description,
      cpu: rawData.cpu,
      settings: rawData.settings,
      results: results
    }))
  })

  const activeBenchmarkId = ref(0)

  const activeBenchmark = computed(() => {
    return benchmarks.value[activeBenchmarkId.value] || benchmarks.value[0]
  })

  /**
   * Load benchmark data from a source
   * Currently simulates async loading for demo purposes
   */
  const loadBenchmarks = async () => {
    loading.value = true
    error.value = null

    try {
      // Simulate network delay
      await new Promise(resolve => setTimeout(resolve, 300))
      // Data already loaded from import
      loading.value = false
    } catch (err) {
      error.value = err as Error
      loading.value = false
    }
  }

  const selectBenchmark = (id: number) => {
    if (id >= 0 && id < benchmarks.value.length) {
      activeBenchmarkId.value = id
    }
  }

  return {
    benchmarks,
    activeBenchmark,
    activeBenchmarkId,
    loading,
    error,
    loadBenchmarks,
    selectBenchmark
  }
}
