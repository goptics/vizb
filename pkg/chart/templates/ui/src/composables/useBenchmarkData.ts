import { ref, computed } from 'vue'
import type { Benchmark, BenchmarkResult } from '../types/benchmark'
import sampleData from '../data/sample.json'

/**
 * Composable for loading and managing benchmark data
 * Groups results by benchmark name to create separate benchmark groups
 */
export function useBenchmarkData() {
  const rawData = sampleData as Benchmark

  const benchmarks = computed<Benchmark[]>(() => {
    const groupMap = new Map<string, BenchmarkResult[]>()

    for (const result of rawData.results) {
      if (!groupMap.has(result.name)) {
        groupMap.set(result.name, []);
      }

      for (const stat of result.stats) {
        stat.value = parseFloat(stat.value.toFixed(2));
      }

      groupMap.get(result.name)!.push(result);
    }


    return Array.from(groupMap.entries()).map(([name, results]) => ({
      name,
      description: rawData.description,
      cpu: rawData.cpu,
      settings: rawData.settings,
      results
    }))
  })

  const activeBenchmarkId = ref(0)
  const activeBenchmark = computed(() => benchmarks.value[activeBenchmarkId.value] || benchmarks.value[0])

  const selectBenchmark = (id: number) => {
    if (id >= 0 && id < benchmarks.value.length) {
      activeBenchmarkId.value = id
    }
  }

  return {
    benchmarks,
    activeBenchmark,
    activeBenchmarkId,
    selectBenchmark
  }
}
