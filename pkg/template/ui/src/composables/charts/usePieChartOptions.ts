import { computed } from "vue";
import type { EChartsOption } from "echarts";
import { type BaseChartConfig, getBaseOptions } from "./baseChartOptions";
import { getNextColorFor, hasXAxis, hasYAxis } from "../../lib/utils";
import { getChartStyling, createPieSeriesConfig } from "./shared";
import { fontSize, sortByTotal, sortByValue } from "./shared/common";
import type { TitleOption } from "echarts/types/dist/shared";

export function usePieChartOptions(config: BaseChartConfig) {
  const { chartData, sort, showLabels, isDark } = config;

  const sortedData = computed(() => {
    // Build a single pie dataset based on totals per xAxis
    const seriesWithTotals = chartData.value.series.map((series) => ({
      ...series,
      // For single category (no y-axis), use the first value; for dual categories, sum all values
      total: hasYAxis(chartData)
        ? series.values.reduce((sum, val) => sum + val, 0)
        : series.values[0] || 0,
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
      const percent = Number(params.percent).toFixed(2);
      return `${params.name} (${percent}%)`;
    };

    // Pie chart for x-axis data
    const xAxisPieData = sorted.series.map((seriesData) => ({
      name: seriesData.xAxis,
      value: seriesData.total || 0,
      itemStyle: { color: getNextColorFor(seriesData.xAxis) },
    }));

    const options: EChartsOption = {
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

    // For single category: show only one pie chart
    if (!hasYAxis(chartData)) {
      return options;
    }

    // For multiple categories: show two pie charts side by side
    // Calculate totals for y-axis data
    const yAxisTotals = new Map<string, number>();
    chartData.value.yAxis.forEach((yAxis, index) => {
      yAxisTotals.set(
        yAxis,
        sorted.series.reduce((sum, series) => {
          return sum + (series.values[index] || 0);
        }, 0)
      );
    });

    const yAxisPieData = chartData.value.yAxis.map((yAxis) => ({
      name: yAxis,
      value: yAxisTotals.get(yAxis) || 0,
      itemStyle: { color: getNextColorFor(yAxis) },
    }));

    if (sort.value.enabled) {
      yAxisPieData.sort(sortByValue(sort.value.order));
    }

    if (!hasXAxis(chartData)) {
      options.series = [
        createPieSeriesConfig(
          chartData.value.statType,
          yAxisPieData,
          showLabels.value,
          styling,
          formatter
        ),
      ];

      return options;
    }

    const titleStyle: TitleOption = {
      text: "X-Axis",
      left: "25%",
      top: "5%",
      textAlign: "center",
      textStyle: {
        color: styling.textColor,
        fontSize,
        fontWeight: "bold",
      },
    };

    options.title = [
      titleStyle,
      {
        ...titleStyle,
        text: "Y-Axis",
        left: "75%",
      },
    ];

    options.series = [
      // Left pie chart: x-axis data
      createPieSeriesConfig(
        `By X-Axis`,
        xAxisPieData,
        showLabels.value,
        styling,
        formatter,
        ["30%", "60%"],
        ["25%", "50%"]
      ),
      // Right pie chart: y-axis data
      createPieSeriesConfig(
        `By Y-Axis`,
        yAxisPieData,
        showLabels.value,
        styling,
        formatter,
        ["30%", "60%"],
        ["75%", "50%"]
      ),
    ];

    return options;
  });

  return { options };
}
