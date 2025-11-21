import { computed, type Ref } from "vue";
import type {
  BenchmarkData,
  ChartData,
  SeriesData,
  Stat,
} from "../types/benchmark";

const createChartTitle = (stat: Stat) => {
  if (stat.unit) {
    return `${stat.type} (${stat.unit}/${stat.per})`;
  }

  return `${stat.type}/${stat.per}`;
};

export function useChartData(results: Ref<BenchmarkData[]> | BenchmarkData[]) {
  const chartData = computed<ChartData[]>(() => {
    const data = Array.isArray(results) ? results : results.value;
    if (!data?.length) return [];

    const firstBenchmarkData = data[0];
    if (!firstBenchmarkData?.stats) return [];

    return firstBenchmarkData.stats.map((stat, statIndex) => {
      const dataMap = new Map<string, Map<string, number>>();
      const xAxisSet = new Set<string>();
      const yAxisSet = new Set<string>();

      data.forEach((benchmarkData) => {
        const { xAxis, yAxis } = benchmarkData;
        const value = benchmarkData.stats[statIndex]?.value || 0;

        yAxisSet.add(yAxis);
        xAxisSet.add(xAxis);

        if (!dataMap.has(yAxis)) {
          dataMap.set(yAxis, new Map());
        }

        dataMap.get(yAxis)!.set(xAxis, value);
      });

      const xAxisValues = Array.from(xAxisSet);
      const yAxisValues = Array.from(yAxisSet);

      // Single workload case - sort by subject values
      const series: SeriesData[] = xAxisValues.map((xAxis) => ({
        xAxis,
        values: yAxisValues.map((yAxis) => dataMap.get(yAxis)?.get(xAxis) || 0),
        benchmarkId: firstBenchmarkData.name, // Add benchmark identifier
      }));

      return {
        title: createChartTitle(stat),
        statType: stat.type,
        statUnit: stat.unit,
        yAxis: yAxisValues,
        series,
      };
    });
  });

  return { chartData };
}
