import { computed } from "vue";
import type { EChartsOption } from "echarts";
import { type BaseChartConfig, getBaseOptions } from "./baseChartOptions";
import { getNextColorFor, hasXAxis, hasYAxis } from "../../lib/utils";
import {
  createAxisConfig,
  createGridConfig,
  createLegendConfig,
  createTooltipConfig,
  getChartStyling,
  getDataZoomConfig,
} from "./shared";
import { sortByTotal } from "./shared/common";

export function useLineChartOptions(config: BaseChartConfig) {
  const { chartData, sort, showLabels, isDark } = config;

  const sortedData = computed(() => {
    if (!sort.value.enabled) {
      return {
        series: chartData.value.series,
        xAxisData: chartData.value.series.map((s) => s.xAxis), // Always use framework names on x-axis
        hasYAxis: hasYAxis(chartData),
      };
    }

    // Sort series by total values
    const seriesWithTotals = chartData.value.series.map((series) => ({
      ...series,
      total: series.values.reduce((sum, val) => sum + val, 0),
    }));

    if (sort.value.enabled) {
      seriesWithTotals.sort(sortByTotal(sort.value.order));
    }

    return {
      series: seriesWithTotals,
      xAxisData: seriesWithTotals.map((s) => s.xAxis), // Always use framework names on x-axis
      hasYAxis: hasYAxis(chartData),
    };
  });

  const options = computed<EChartsOption>(() => {
    const { series, xAxisData, hasYAxis } = sortedData.value;
    const baseOptions = getBaseOptions(config);
    const styling = getChartStyling(isDark.value);

    // For single category: create one series with each x-axis value as a data point
    // For dual categories: create multiple series (one per x-axis value)
    if (!hasYAxis) {
      // Single category case: one series with multiple x-axis points
      const opt = {
        ...baseOptions,
        grid: createGridConfig(1),
        tooltip: createTooltipConfig(false),
      };

      return {
        ...opt,
        ...createAxisConfig(styling, xAxisData),
        dataZoom: getDataZoomConfig(xAxisData.length, styling),
        legend: { show: false },
        series: [
          {
            name: chartData.value.title,
            type: "line",
            data: series.map((seriesData) => ({
              value: seriesData.values[0], // Take the first (and only) value
              label: {
                show: showLabels.value,
                position: "top",
                formatter: "{c}",
                fontSize: 10,
                color: styling.textColor,
              },
            })),
            itemStyle: { color: getNextColorFor(chartData.value.title) },
            lineStyle: {
              width: 2,
              type: "solid",
            },
            smooth: false,
            symbol: "circle",
            symbolSize: 6,
            emphasis: {
              focus: "series",
              itemStyle: {
                borderWidth: 2,
                borderColor: "#fff",
              },
            },
          },
        ],
      };
    }

    // Dual categories case: transpose data to show y-axis values as series
    // Each y-axis value becomes a line, with x-axis values (frameworks) as points
    const yAxisLabels = chartData.value.yAxis;
    const transposedSeries = yAxisLabels.map((yAxisLabel, yIndex) => ({
      name: yAxisLabel,
      type: "line" as const,
      data: series.map((seriesData) => ({
        value: seriesData.values[yIndex] || 0,
        label: {
          show: showLabels.value,
          position: "top",
          formatter: "{c}",
          fontSize: 10,
          color: styling.textColor,
        },
      })),
      itemStyle: { color: getNextColorFor(yAxisLabel) },
      lineStyle: {
        width: 2,
        type: "solid" as const,
      },
      smooth: false,
      symbol: "circle",
      symbolSize: 6,
      emphasis: {
        focus: "series",
        itemStyle: {
          borderWidth: 2,
          borderColor: "#fff",
        },
      },
    }));

    const hasMultipleSeries = transposedSeries.length > 1;
    const opt = {
      ...baseOptions,
      grid: createGridConfig(transposedSeries.length),
      tooltip: createTooltipConfig(true),
    };

    return {
      ...opt,
      ...createAxisConfig(styling, xAxisData),
      dataZoom: getDataZoomConfig(xAxisData.length, styling),
      legend: createLegendConfig(
        transposedSeries.map((s) => ({ xAxis: s.name })),
        styling,
        hasMultipleSeries
      ),
      series: transposedSeries as any,
    };
  });

  return { options };
}
