import { computed } from "vue";
import type { EChartsOption } from "echarts";
import { type BaseChartConfig, getBaseOptions } from "./baseChartOptions";
import { getNextColorFor } from "../../lib/utils";
import { getChartStyling, createPieSeriesConfig } from "./shared";
import { sortByTotal } from "./shared/common";

export function usePieChartOptions(config: BaseChartConfig) {
  const { chartData, sortOrder, showLabels, isDark } = config;

  const sortedData = computed(() => {
    // Check if we have subjectTotals (multiple workloads case)
    if (chartData.value.subjectTotals) {
      // Prepare subject data
      const subjects = Object.entries(chartData.value.subjectTotals)
        .map(([subject, total]) => ({
          subject,
          values: [total],
          total,
        }))
        .sort(sortByTotal(sortOrder.value));

      // Prepare workload data
      const workloadTotals = new Map<string, number>();
      chartData.value.series.forEach((series) => {
        const workloadTotal = series.values.reduce((sum, val) => sum + val, 0);
        workloadTotals.set(series.subject, workloadTotal);
      });

      const workloads = Array.from(workloadTotals.entries())
        .map(([workload, total]) => ({
          subject: workload,
          values: [total],
          total,
        }))
        .sort(sortByTotal(sortOrder.value));

      return { subjects, workloads };
    }

    // Single workload case: use series directly
    const seriesWithTotals = chartData.value.series
      .map((series) => ({
        ...series,
        total: series.values.reduce((sum, val) => sum + val, 0),
      }))
      .sort(sortByTotal(sortOrder.value));

    return { series: seriesWithTotals };
  });

  const options = computed<EChartsOption>(() => {
    const sorted = sortedData.value;
    const styling = getChartStyling(isDark.value);
    const baseOptions = getBaseOptions(config);

    const formatter = (params: any) => {
      const value = Number(params.value).toFixed(2);
      const percent = Number(params.percent).toFixed(2);
      return `${params.name}\n${value} (${percent}%)`;
    };

    // Check if we have both subjects and workloads (multiple workloads case)
    if (sorted.subjects && sorted.workloads) {
      // Prepare subject pie chart data
      const subjectPieData = sorted.subjects.map((seriesData) => ({
        name: seriesData.subject,
        value: seriesData.total || 0,
        itemStyle: { color: getNextColorFor(seriesData.subject) },
      }));

      // Prepare workload pie chart data
      const workloadPieData = sorted.workloads.map((seriesData) => ({
        name: seriesData.subject,
        value: seriesData.total || 0,
        itemStyle: { color: getNextColorFor(seriesData.subject) },
      }));

      const subjectTitle = {
        text: "Subjects",
        left: "25%",
        top: "5%",
        textAlign: "center" as const,
        textStyle: {
          fontSize: 16,
          fontWeight: "bold" as const,
          color: styling.textColor,
        },
      };

      // Show two pie charts
      return {
        ...baseOptions,
        legend: { show: false },
        title: [
          subjectTitle,
          {
            ...subjectTitle,
            left: "75%",  // Align with the right pie chart
            text: "Workloads",
          },
        ],
        series: [
          createPieSeriesConfig(
            `${chartData.value.statType} (Subjects)`,
            subjectPieData,
            showLabels.value,
            styling,
            formatter,
            ['30%', '60%'],  // Smaller radius for left pie
            ['25%', '50%']   // Position left pie
          ),
          createPieSeriesConfig(
            `${chartData.value.statType} (Workloads)`,
            workloadPieData,
            showLabels.value,
            styling,
            formatter,
            ['30%', '60%'],  // Smaller radius for right pie
            ['75%', '50%']   // Position right pie
          ),
        ],
      };
    }

    // Single pie chart for subjects only
    const singlePieData = sorted.series.map((seriesData) => ({
      name: seriesData.subject,
      value: seriesData.total || 0,
      itemStyle: { color: getNextColorFor(seriesData.subject) },
    }));

    return {
      ...baseOptions,
      grid: {
        top: "10%",
        bottom: "10%",
        left: "10%",
        right: "10%",
      },
      legend: { show: false },
      series: [
        createPieSeriesConfig(
          chartData.value.statType,
          singlePieData,
          showLabels.value,
          styling,
          formatter
        ),
      ],
    };
  });

  return { options };
}
