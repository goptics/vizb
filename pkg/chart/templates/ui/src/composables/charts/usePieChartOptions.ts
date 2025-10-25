import { computed } from "vue";
import type { EChartsOption } from "echarts";
import { type BaseChartConfig, getBaseOptions } from "./baseChartOptions";
import { getNextColorFor, hasYAxis } from "../../lib/utils";
import { getChartStyling, createPieSeriesConfig } from "./shared";
import { sortByTotal } from "./shared/common";

export function usePieChartOptions(config: BaseChartConfig) {
  const { chartData, sort, showLabels, isDark } = config;

  const sortedData = computed(() => {
    // Check if we have y-axis data (dual categories)
    const hasYAxis = chartData.value.yAxis && chartData.value.yAxis.length > 0 && chartData.value.yAxis[0] !== "";
    
    // Build a single pie dataset based on totals per xAxis
    const seriesWithTotals = chartData.value.series.map((series) => ({
      ...series,
      // For single category (no y-axis), use the first value; for dual categories, sum all values
      total: hasYAxis ? series.values.reduce((sum, val) => sum + val, 0) : (series.values[0] || 0),
    }));
    
    if (sort.value.enabled) {
      seriesWithTotals.sort(sortByTotal(sort.value.order));
    }

    return { series: seriesWithTotals };
  });

  const options = computed<EChartsOption>(() => {
    const sorted = sortedData.value;
    const styling = getChartStyling(isDark.value);
    const baseOptions = getBaseOptions(config);
    
    // Check if we have y-axis data (dual categories)

    const formatter = (params: any) => {
      const value = Number(params.value).toFixed(2);
      const percent = Number(params.percent).toFixed(2);
      return `${params.name}\n${value} (${percent}%)`;
    };

    // Pie chart for x-axis data
    const xAxisPieData = sorted.series.map((seriesData) => ({
      name: seriesData.xAxis,
      value: seriesData.total || 0,
      itemStyle: { color: getNextColorFor(seriesData.xAxis) },
    }));

    // For single category: show only one pie chart
    if (!hasYAxis(chartData)) {
      return {
        ...baseOptions,
        legend: { show: false },
        series: [
          createPieSeriesConfig(
            chartData.value.statType,
            xAxisPieData,
            showLabels.value,
            styling,
            formatter
          ),
        ],
      };
    }

    // For multiple categories: show two pie charts side by side
    // Calculate totals for y-axis data
    const yAxisTotals = new Map<string, number>();
    chartData.value.yAxis.forEach((yAxis, index) => {
      let total = 0;
      sorted.series.forEach((series) => {
        total += series.values[index] || 0;
      });
      yAxisTotals.set(yAxis, total);
    });

    const yAxisPieData = chartData.value.yAxis.map((yAxis) => ({
      name: yAxis,
      value: yAxisTotals.get(yAxis) || 0,
      itemStyle: { color: getNextColorFor(yAxis) },
    }));

    return {
      ...baseOptions,
      title: [
        {
          text: 'X-Axis',
          left: '25%',
          top: '5%',
          textAlign: 'center',
          textStyle: {
            color: styling.textColor,
            fontSize: 12,
            fontWeight: 'bold',
          },
        },
        {
          text: 'Y-Axis',
          left: '75%',
          top: '5%',
          textAlign: 'center',
          textStyle: {
            color: styling.textColor,
            fontSize: 12,
            fontWeight: 'bold',
          },
        },
      ],
      legend: { show: false },
      series: [
        // Left pie chart: x-axis data
        createPieSeriesConfig(
          `By X-Axis`,
          xAxisPieData,
          showLabels.value,
          styling,
          formatter,
          ['30%', '60%'],
          ['25%', '50%']
        ),
        // Right pie chart: y-axis data
        createPieSeriesConfig(
          `By Y-Axis`,
          yAxisPieData,
          showLabels.value,
          styling,
          formatter,
          ['30%', '60%'],
          ['75%', '50%']
        ),
      ],
    };
  });

  return { options };
}
