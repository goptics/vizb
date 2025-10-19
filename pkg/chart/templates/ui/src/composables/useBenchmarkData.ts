import { ref, computed } from "vue";
import {
  DEFAULT_SETTINGS,
  type Benchmark,
  type BenchmarkData,
} from "../types/benchmark";
import { resetColor } from "../lib/utils";

const getBenchmarks = async (): Promise<Benchmark[]> => {
  if (import.meta.env.DEV) {
    const data = await import("../data/sample.json");
    return data.default as Benchmark[];
  }

  try {
    return JSON.parse(window.VIZB_DATA);
  } catch (error) {
    console.error("Failed to parse VIZB_DATA:", error);
    return [];
  }
};

/**
 * Composable for loading and managing benchmark data
 * Groups results by benchmark name to create separate benchmark groups
 */
export function useBenchmarkData() {
  const benchmarks = ref<Benchmark[]>([]);

  // Load data immediately
  getBenchmarks().then((data) => {
    benchmarks.value = data;
  });

  // Process and group all benchmarks
  const benchmarksProcessed = computed<Benchmark[]>(() => {
    if (!benchmarks.value.length) return [];

    return benchmarks.value.map((benchmark) => {
      // Process stats for each result
      for (const result of benchmark.data) {
        for (const stat of result.stats) {
          stat.value = parseFloat(stat.value.toFixed(2));
        }
      }

      return benchmark;
    });
  });

  // Track active benchmark selection
  const activeBenchmarkId = ref(0);
  const activeBenchmark = computed(
    () =>
      benchmarksProcessed.value[activeBenchmarkId.value] ||
      benchmarksProcessed.value[0]
  );

  // Group results within the active benchmark
  const grouped = computed(() => {
    if (!activeBenchmark.value) return new Map<string, BenchmarkData[]>();

    const groupMap = new Map<string, BenchmarkData[]>();

    for (const benchmarkData of activeBenchmark.value.data) {
      if (!groupMap.has(benchmarkData.name)) {
        groupMap.set(benchmarkData.name, []);
      }
      groupMap.get(benchmarkData.name)!.push(benchmarkData);
    }

    return groupMap;
  });

  // Convert grouped results to array format for easier consumption
  const benchmarkGroups = computed(() =>
    Array.from(grouped.value.entries()).map(([name, data]) => ({
      name,
      description: activeBenchmark.value?.description || "",
      cpu: {
        name: activeBenchmark.value?.cpu?.name || "",
        cores: activeBenchmark.value?.cpu?.cores || 0,
      },
      settings: activeBenchmark.value?.settings || DEFAULT_SETTINGS,
      data,
    }))
  );

  const activeGroupId = ref(0);
  const activeGroup = computed(
    () => benchmarkGroups.value[activeGroupId.value] || benchmarkGroups.value[0]
  );

  const selectBenchmark = (id: number) => {
    if (id >= 0 && id < benchmarks.value.length) {
      // Reset color mapping when benchmark changes
      resetColor();

      activeBenchmarkId.value = id;
      // Reset group selection when benchmark changes
      activeGroupId.value = 0;
    }
  };

  const selectGroup = (id: number) => {
    if (id >= 0 && id < benchmarkGroups.value.length) {
      activeGroupId.value = id;
    }
  };

  return {
    // Top level benchmark selection
    benchmarks,
    activeBenchmark,
    activeBenchmarkId,
    selectBenchmark,

    // Inner level group selection
    resultGroups: benchmarkGroups,
    activeGroup,
    activeGroupId,
    selectGroup,
  };
}
