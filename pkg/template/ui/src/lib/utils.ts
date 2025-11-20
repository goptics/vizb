import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";
import type { ChartData } from "../types/benchmark";
import type { Ref } from "vue";

/**
 * Utility function to merge Tailwind CSS classes
 * Used for conditional styling with shadcn components
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export const COLOR_PALETTE = [
  "#5470C6", // Blue
  "#3BA272", // Green
  "#FC8452", // Orange
  "#73C0DE", // Light blue
  "#EE6666", // Red
  "#FAC858", // Yellow
  "#9A60B4", // Purple
  "#EA7CCC", // Pink
  "#91CC75", // Lime
  "#FF9F7F", // Coral
];

const colorMap = new Map<string, number>();
let i = 0;

export function getNextColorFor(key: string) {
  if (colorMap.has(key)) {
    return COLOR_PALETTE[colorMap.get(key)!];
  }

  const colorIndex = i % COLOR_PALETTE.length;
  const color = COLOR_PALETTE[colorIndex];
  colorMap.set(key, colorIndex);

  if (i === COLOR_PALETTE.length) {
    i = 0;
  } else {
    i++;
  }

  return color;
}

export const resetColor = () => {
  colorMap.clear();
  i = 0;
};

export const hasXAxis = (chartData: Ref<ChartData, ChartData>) =>
  chartData.value.series.some(
    (series) => series.xAxis && series.xAxis.trim() !== ""
  );

export const hasYAxis = (chartData: Ref<ChartData, ChartData>) =>
  chartData.value.yAxis &&
  chartData.value.yAxis.length > 0 &&
  chartData.value.yAxis[0] !== "";
