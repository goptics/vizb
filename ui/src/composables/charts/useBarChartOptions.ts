import { computed } from "vue";
import type { EChartsOption } from "echarts";
import { type BaseChartConfig, getBaseOptions } from "./baseChartOptions";
import { getNextColorFor, hasXAxis } from "@/lib/utils";
import {
  createAxisConfig,
  createDataZoomConfig,
  createGridConfig,
  createHorizontalAxisConfig,
  createHorizontalDataZoomConfig,
  createLabelConfig,
  createLegendConfig,
  createTooltipConfig,
  getChartStyling,
  horizontalLegendBottom,
  isLargeXAxis,
  makeLegendTitle,
  LARGE_DATA_THRESHOLD,
} from "./shared";
import {
  useSortedSeriesData,
  getEffectiveScale,
  computeSeriesTotals,
} from "./shared/common";
import { buildValueAxes2DOptions } from "./shared/valueMode";
import { buildMixedAxes2DOptions } from "./shared/mixedMode";

/**
 * Applies borderRadius to a chart result (mixed/value modes).
 * @param result - ECharts option object
 * @param radius - Border radius in pixels
 * @param isHorizontal - Whether bars are horizontal
 */

const applyBorderRadiusToResult = (result: EChartsOption, radius: number, isHorizontal: boolean) => {
  if (radius <= 0) return result;
  const series = result.series as any[];
  if (!series) return result;
  const borderRadius = isHorizontal ? [0, radius, radius, 0] : [radius, radius, 0, 0];
  series.forEach((s: any) => {
    s.itemStyle = {
      ...s.itemStyle,
      borderRadius,
    };
  });
  return result;
};

const barNullable = (val: number | null, scale: string): number | null =>
  val === null ? null : scale === "log" && val <= 0 ? null : val;

// Mirrors internal/charts/bar/bar.go's GetCornerValues: [topLeft, topRight, bottomRight, bottomLeft].
const cornerValues = (radius: number, isHorizontal: boolean): number[] =>
  isHorizontal ? [0, radius, radius, 0] : [radius, radius, 0, 0];

export function useBarChartOptions(config: BaseChartConfig) {
  const {
    chartData,
    sort,
    showLabels,
    isDark,
    scale,
    stack,
    horizontal,
    borderRadius,
  } = config;

  const sortedData = useSortedSeriesData(chartData, sort);
  const options = computed<EChartsOption>(() => {
    const isHorizontal = horizontal?.value ?? false;
    const radius = borderRadius?.value ?? 0;

    if (chartData.value.mixedTuples?.length) {
  const result = buildMixedAxes2DOptions(config, 'bar');
  applyBorderRadiusToResult(result, radius, isHorizontal);
  return result;
}
if (chartData.value.valueTuples?.length) {
  const result = buildValueAxes2DOptions(config, 'bar');
  applyBorderRadiusToResult(result, radius, isHorizontal);
  return result;
}

    const { series, xAxisData, hasYAxis } = sortedData.value;
    const baseOptions = getBaseOptions(config);
    const styling = getChartStyling(isDark.value);
    const { minValue, effectiveScale } = getEffectiveScale(
      series,
      scale?.value ?? "linear",
    );
    const largeX = isLargeXAxis(xAxisData);
    const xLabel = chartData.value.axisLabels?.x;
    const useStack = stack?.value === true && effectiveScale !== "log";

    // Helper to apply borderRadius to series
    const applyBorderRadius = (
      seriesItem: any,
      isTopSeries: boolean = true,
    ) => {
      if (radius <= 0) return seriesItem;
      const isTop = !useStack || isTopSeries;
      seriesItem.itemStyle = {
        ...seriesItem.itemStyle,
        borderRadius: isTop ? cornerValues(radius, isHorizontal) : [0, 0, 0, 0],
      };
      return seriesItem;
    };

    if (!hasYAxis && isHorizontal) {
      const singleSeries = {
        name: chartData.value.title,
        type: "bar" as const,
        data: series.map((s) =>
          barNullable(s.values[0] ?? null, effectiveScale),
        ),
        label: createLabelConfig(showLabels.value, styling, "horizontal"),
        large: true,
        largeThreshold: LARGE_DATA_THRESHOLD,
        itemStyle: { color: getNextColorFor(chartData.value.title) },
      };
      applyBorderRadius(singleSeries, true);

      return {
        ...baseOptions,
        grid: {
          left: xLabel ? 70 : "3%",
          right: largeX ? 44 : 24,
          bottom: "3%",
          top: 8,
          containLabel: true,
        },
        tooltip: createTooltipConfig(false, isDark.value),
        legend: { show: false },
        ...createHorizontalAxisConfig(
          styling,
          xAxisData,
          effectiveScale,
          minValue,
          xLabel,
          largeX,
        ),
        ...(largeX
          ? { dataZoom: createHorizontalDataZoomConfig(styling) }
          : {}),
        series: [singleSeries],
      } as EChartsOption;
    }

    if (!hasYAxis) {
      const singleSeries = {
        name: chartData.value.title,
        type: "bar" as const,
        data: series.map((s) =>
          barNullable(s.values[0] ?? null, effectiveScale),
        ),
        label: createLabelConfig(showLabels.value, styling),
        large: true,
        largeThreshold: LARGE_DATA_THRESHOLD,
        itemStyle: { color: getNextColorFor(chartData.value.title) },
      };
      applyBorderRadius(singleSeries, true);

      return {
        ...baseOptions,
        grid: createGridConfig(1, largeX),
        tooltip: createTooltipConfig(false, isDark.value),
        legend: { show: false },
        ...createAxisConfig(
          styling,
          xAxisData,
          effectiveScale,
          minValue,
          xLabel,
          largeX,
        ),
        ...(largeX
          ? { dataZoom: createDataZoomConfig(xAxisData, styling) }
          : {}),
        series: [singleSeries],
      } as EChartsOption;
    }

    // Dual categories: transpose — each y-axis value becomes a bar group
const yAxisLabels = chartData.value.yAxis;
const transposedSeries = yAxisLabels.map((yAxisLabel, yIndex) => ({
  name: yAxisLabel,
  type: "bar" as const,
  data: series.map((s) =>
    barNullable(s.values[yIndex] ?? null, effectiveScale),
  ),
  label: createLabelConfig(
    showLabels.value,
    styling,
    isHorizontal ? "horizontal" : "vertical",
    useStack,
  ),
  large: true,
  largeThreshold: LARGE_DATA_THRESHOLD,
  ...(useStack ? { stack: "total" } : {}),
  itemStyle: { color: getNextColorFor(yAxisLabel) },
}));

// Secondary sort when there is only one x-group (sort before applying borderRadius)
if (sort.value.enabled && xAxisData.length === 1) {
  transposedSeries.sort((a, b) => {
    const valA = a.data[0] ?? 0;
    const valB = b.data[0] ?? 0;
    return sort.value.order === "asc" ? valA - valB : valB - valA;
  });
}

// Apply borderRadius after sorting (top series = last element after sort)
if (radius > 0) {
  transposedSeries.forEach((seriesItem, index) => {
    const isTopSeries = useStack && index === transposedSeries.length - 1;
    const isTop = !useStack || isTopSeries;
    // Use type assertion to tell TypeScript this is allowed
    (seriesItem.itemStyle as any).borderRadius = isTop
      ? cornerValues(radius, isHorizontal)
      : [0, 0, 0, 0];
  });
}

    const hasMultipleSeries = transposedSeries.length > 1;
    const seriesTotals = computeSeriesTotals(transposedSeries);
    const yLabel = chartData.value.axisLabels?.y;
    const showLegendTitle = hasMultipleSeries && !!yLabel;

    if (isHorizontal) {
      return {
        ...baseOptions,
        grid: {
          left: xLabel ? 70 : "3%",
          right: largeX ? 44 : 24,
          bottom: hasMultipleSeries
            ? horizontalLegendBottom(transposedSeries.length)
            : "3%",
          top: 8,
          containLabel: true,
        },
        tooltip: createTooltipConfig(
          hasXAxis(chartData),
          isDark.value,
          seriesTotals,
        ),
        legend: createLegendConfig(
          transposedSeries.map((s) => ({ xAxis: s.name })),
          styling,
          hasMultipleSeries,
          { bottom: 0 },
        ),
        ...createHorizontalAxisConfig(
          styling,
          xAxisData,
          effectiveScale,
          minValue,
          xLabel,
          largeX,
        ),
        ...(largeX
          ? { dataZoom: createHorizontalDataZoomConfig(styling) }
          : {}),
        series: transposedSeries,
      } as EChartsOption;
    }

    return {
      ...baseOptions,
      ...(showLegendTitle ? { title: makeLegendTitle(yLabel!, styling) } : {}),
      grid: createGridConfig(transposedSeries.length, largeX),
      tooltip: createTooltipConfig(
        hasXAxis(chartData),
        isDark.value,
        seriesTotals,
      ),
      legend: createLegendConfig(
        transposedSeries.map((s) => ({ xAxis: s.name })),
        styling,
        hasMultipleSeries,
        showLegendTitle ? { top: 24 } : undefined,
      ),
      ...createAxisConfig(
        styling,
        xAxisData,
        effectiveScale,
        minValue,
        xLabel,
        largeX,
      ),
      ...(largeX ? { dataZoom: createDataZoomConfig(xAxisData, styling) } : {}),
      series: transposedSeries,
    } as EChartsOption;
  });

  return { options };
}
