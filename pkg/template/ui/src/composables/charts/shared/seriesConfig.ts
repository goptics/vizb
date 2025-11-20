/**
 * Creates pie chart label configuration
 */
export function createPieLabelConfig(
  showLabels: boolean,
  styling: { textColor: string },
  customFormatter?: (params: any) => string
): any {
  const defaultFormatter = (params: any) => {
    const name = params.name
    return `${name}\n${params.value}\n${params.percent}%`
  }

  return {
    show: showLabels,
    formatter: customFormatter || defaultFormatter,
    fontSize: 9,
    color: styling.textColor,
    fontWeight: 'bold',
  }
}

/**
 * Creates pie chart label line configuration
 */
export function createPieLabelLineConfig(
  showLabels: boolean,
  styling: { textColor: string }
): any {
  return {
    show: showLabels,
    length: 8,
    length2: 4,
    lineStyle: {
      color: styling.textColor,
      width: 1,
    },
  }
}

/**
 * Creates pie chart emphasis configuration
 */
export function createPieEmphasisConfig(showLabels: boolean) {
  return {
    itemStyle: {
      shadowBlur: 10,
      shadowOffsetX: 0,
      shadowColor: 'rgba(0, 0, 0, 0.5)',
    },
    label: {
      show: showLabels,
      fontSize: 11,
      fontWeight: 'bold',
    },
  }
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
  radius: [string, string] = ['40%', '70%'],
  center: [string, string] = ['50%', '50%'],
): any {
  return {
    name,
    type: "pie",
    radius,
    center,
    data,
    label: createPieLabelConfig(showLabels, styling, customFormatter),
    labelLine: createPieLabelLineConfig(showLabels, styling),
    emphasis: createPieEmphasisConfig(showLabels),
  }
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
        fontSize: 10,
        color: styling.textColor,
      },
    })),
    itemStyle: { color },
    lineStyle: {
      width: 2,
      type: "solid",
    },
    smooth: false,
    symbol: "circle",
    symbolSize: 6,
    emphasis: createEmphasisConfig('series'),
    ...customConfig,
  }
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
        fontSize: 10,
        color: styling.textColor,
      },
    })),
    itemStyle: { color },
    emphasis: createEmphasisConfig('series'),
    ...customConfig,
  }
}

/**
 * Helper function to create emphasis configuration
 */
function createEmphasisConfig(focusType: 'series' | 'self' = 'series'): any {
  return {
    focus: focusType,
    itemStyle: {
      borderWidth: 2,
      borderColor: "#fff",
    },
  }
}
