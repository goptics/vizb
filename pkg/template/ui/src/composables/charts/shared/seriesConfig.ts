import { fontSize } from "./common";

/**
 * Creates pie chart label configuration
 */
export function createPieLabelConfig(
  showLabels: boolean,
  styling: { textColor: string },
  customFormatter?: (params: any) => string
): any {
  return {
    show: showLabels,
    formatter: customFormatter,
    fontSize,
    color: styling.textColor,
  };
}

/**
 * Creates pie chart series configuration
 */
export function createPieSeriesConfig(
  name: string,
  data: any[],
  showLabels: boolean,
  styling: { textColor: string },
  customFormatter?: (params: any) => string,
  radius: [string, string] = ["40%", "70%"],
  center: [string, string] = ["50%", "50%"]
): any {
  return {
    name,
    type: "pie",
    radius,
    center,
    data,
    label: createPieLabelConfig(showLabels, styling, customFormatter),
  };
}

/**
 * Creates line chart series configuration
 */
export function createLineSeriesConfig(
  seriesData: any,
  showLabels: boolean,
  styling: { textColor: string },
  color: string,
  customConfig?: any
): any {
  return {
    name: seriesData.subject,
    type: "line",
    data: seriesData.values.map((value: number) => ({
      value,
      label: {
        show: showLabels,
        position: "top",
        formatter: "{c}",
        fontSize,
        color: styling.textColor,
      },
    })),
    itemStyle: { color },
    ...customConfig,
  };
}

/**
 * Creates bar chart series configuration
 */
export function createBarSeriesConfig(
  seriesData: any,
  showLabels: boolean,
  styling: { textColor: string },
  color: string,
  customConfig?: any
): any {
  return {
    name: seriesData.subject,
    type: "bar",
    data: seriesData.values.map((value: number) => ({
      value,
      label: {
        show: showLabels,
        position: "top",
        formatter: "{c}",
        fontSize,
        color: styling.textColor,
      },
    })),
    itemStyle: { color },
    ...customConfig,
  };
}
