import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions, formatValue } from './baseChartOptions'
import { getNextColorFor } from '../../lib/utils'
import { calculateLegendSpace } from './shared/common'
import { createAxisConfig, createGridConfig, createLegendConfig, createTooltipConfig, getChartStyling, getDataZoomConfig } from './shared'

export function useBarChartOptions(config: BaseChartConfig) {
  const { chartData, sortOrder, showLabels, isDark } = config
  
  const sortedData = computed(() => {
    if (sortOrder.value === "") {
      return {
        series: chartData.value.series,
        xAxisData: chartData.value.workloads,
      };
    }

    // Check if we have multiple workloads (subjectTotals will exist)
    const hasSubjectTotals = chartData.value.subjectTotals !== undefined;

    if (hasSubjectTotals) {
      // Sort X-axis (subjects) based on their totals across all workloads
      const sortedSubjects = chartData.value.workloads
        .map((subject) => ({
          subject,
          total: chartData.value.subjectTotals![subject] || 0,
        }))
        .sort((a, b) => {
          if (sortOrder.value === "asc") return a.total - b.total;
          return b.total - a.total;
        })
        .map((item) => item.subject);

      // Rebuild series data with sorted X-axis order
      const sortedSeries = chartData.value.series.map((series) => {
        const subjectIndexMap = new Map(
          chartData.value.workloads.map((subject, idx) => [subject, idx])
        );

        return {
          ...series,
          values: sortedSubjects.map((subject) => {
            const idx = subjectIndexMap.get(subject);
            return idx !== undefined ? series.values[idx] : 0;
          }),
        };
      });

      return {
        series: sortedSeries,
        xAxisData: sortedSubjects,
      };
    }

    // Single workload case - sort series by their total values
    const seriesWithTotals = chartData.value.series.map((series) => ({
      ...series,
      total: series.values.reduce((sum, val) => sum + val, 0),
    }));
    seriesWithTotals.sort((a, b) => {
      if (sortOrder.value === "asc") return (a.total || 0) - (b.total || 0);
      return (b.total || 0) - (a.total || 0);
    });

    return {
      series: seriesWithTotals,
      xAxisData: chartData.value.workloads,
    };
  });

  const options = computed<EChartsOption>(() => {
    const { series, xAxisData } = sortedData.value;
    const hasMultipleSeries = series.length > 1;

    const styling = getChartStyling(isDark.value);
    const hasMultipleWorkloads = chartData.value.workloads.length > 1;

    const baseOptions = getBaseOptions(config);

    return {
      ...baseOptions,
      grid: createGridConfig("bar", series.length),
      tooltip: createTooltipConfig(chartData.value, hasMultipleWorkloads),
      legend: createLegendConfig(series, styling, hasMultipleSeries),
      ...createAxisConfig(styling, xAxisData),
      dataZoom: getDataZoomConfig(xAxisData.length, styling),
      series: series.map((seriesData) => ({
        name: seriesData.subject,
        type: "bar",
        data: seriesData.values.map((value) => ({
          value,
          label: {
            show: showLabels.value,
            position: "top",
            formatter: "{c}",
            fontSize: 10,
            color: styling.textColor,
          },
        })),
        itemStyle: { color: getNextColorFor(seriesData.subject) },
        emphasis: {
          focus: 'series'
        }
      })),
    };
  });

  return { options };
}
