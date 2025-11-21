export interface ChartStyling {
  textColor: string;
  axisColor: string;
  opacity: number;
}

/**
 * Gets consistent styling colors based on dark mode
 */
export function getChartStyling(isDark: boolean): ChartStyling {
  return {
    textColor: isDark ? "#e5e7eb" : "#374151",
    axisColor: isDark ? "#4b5563" : "#d1d5db",
    opacity: isDark ? 0.2 : 0.8,
  };
}

/**
 * Creates common axis configuration
 */
export function createAxisConfig(
  styling: ChartStyling,
  xAxisData: string[]
): { xAxis: any; yAxis: any } {
  return {
    xAxis: {
      type: "category",
      data: xAxisData,
      axisLabel: {
        interval: 0,
        rotate: xAxisData.length > 15 ? 30 : 0,
        fontSize: 10,
        color: styling.textColor,
      },
      axisLine: {
        lineStyle: { color: styling.axisColor },
      },
    },
    yAxis: {
      type: "value",
      splitLine: {
        lineStyle: {
          opacity: styling.opacity,
        },
      },
      axisLabel: {
        color: styling.textColor,
      },
      axisLine: {
        lineStyle: { color: styling.axisColor },
      },
    },
  };
}

/**
 * Creates common tooltip configuration
 */
export function createTooltipConfig(hasXYAxis: boolean): any {
  if (hasXYAxis) {
    return {
      trigger: "axis",
      axisPointer: { type: "shadow" },
      formatter: (params: any) => {
        if (!Array.isArray(params)) return "";

        let result = `<strong>${params[0].axisValue}</strong><br/>`;
        params.forEach((param: any) => {
          result += `${param.marker} ${param.seriesName}: ${param.value}<br/>`;
        });
        return result;
      },
    };
  }

  return {
    trigger: "item",
    formatter: (params: any) => {
      const param = Array.isArray(params) ? params[0] : params;
      return `${param.marker} <strong>${
        param.name || param.seriesName
      }</strong><br/>${param.value}`;
    },
  };
}

/**
 * Creates common legend configuration
 */
export function createLegendConfig(
  series: any[],
  styling: ChartStyling,
  hasMultipleSeries: boolean,
  customConfig?: any
): any {
  if (!hasMultipleSeries) {
    return { show: false };
  }

  return {
    show: true,
    left: "center",
    itemWidth: 10,
    itemHeight: 10,
    textStyle: { fontSize: 12, color: styling.textColor },
    data: series.map((s) => s.xAxis),
    ...customConfig,
  };
}

/**
 * Creates common grid configuration
 */
export function createGridConfig(seriesLength = 1): any {
  const legendSpace = Math.min(
    15 + Math.floor((seriesLength - 1) / 15) * 2,
    35
  );

  return {
    left: "3%",
    right: "3%",
    bottom: "10%",
    top: `${legendSpace}%`,
    containLabel: true,
  };
}
