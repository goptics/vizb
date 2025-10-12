import type { ChartData } from '../../../types/benchmark'
import { formatValue } from '../baseChartOptions'

export interface ChartStyling {
  textColor: string
  axisLineColor: string
  splitLineColor: string
}

/**
 * Gets consistent styling colors based on dark mode
 */
export function getChartStyling(isDark: boolean): ChartStyling {
  return {
    textColor: isDark ? "#e5e7eb" : "#374151",
    axisLineColor: isDark ? "#4b5563" : "#d1d5db",
    splitLineColor: isDark ? "#4b5563" : "#d1d5db",
  }
}

/**
 * Creates common axis configuration
 */
export function createAxisConfig(
  styling: ChartStyling,
  _isDark: boolean,
  xAxisData: string[],
  isLineChart = false
): { xAxis: any; yAxis: any } {
  return {
    xAxis: {
      type: "category",
      data: xAxisData,
      axisLabel: {
        interval: isLineChart ? "auto" : 0,
        rotate: isLineChart ? 45 : 0,
        fontSize: isLineChart ? 10 : 11,
        color: styling.textColor,
        margin: isLineChart ? 8 : undefined,
      },
      axisLine: {
        lineStyle: { color: styling.axisLineColor },
      },
    },
    yAxis: {
      type: "value",
      splitLine: {
        show: true,
        lineStyle: {
          type: "solid",
          opacity: 0.2,
          color: styling.splitLineColor,
        },
      },
      axisLabel: {
        color: styling.textColor,
      },
      axisLine: {
        lineStyle: { color: styling.axisLineColor },
      },
    },
  }
}

/**
 * Creates common tooltip configuration
 */
export function createTooltipConfig(
  chartData: ChartData,
  hasMultipleWorkloads: boolean,
  _styling: ChartStyling
): any {
  if (hasMultipleWorkloads) {
    return {
      trigger: "axis",
      axisPointer: { type: "shadow" },
      formatter: (params: any) => {
        if (!Array.isArray(params)) return ""
        let result = `<strong>${params[0].axisValue}</strong><br/>`
        params.forEach((param: any) => {
          const value = formatValue(param.value, chartData.statUnit)
          result += `${param.marker} ${param.seriesName}: ${value}<br/>`
        })
        return result
      },
    }
  }

  return {
    trigger: "item",
    formatter: (params: any) => {
      const param = Array.isArray(params) ? params[0] : params
      const value = formatValue(param.value, chartData.statUnit)
      return `${param.marker} <strong>${param.seriesName}</strong><br/>${value}`
    },
  }
}

/**
 * Creates common legend configuration
 */
export function createLegendConfig(
  series: any[],
  _styling: ChartStyling,
  hasMultipleSeries: boolean,
  customConfig?: any
): any {
  if (!hasMultipleSeries) {
    return { show: false }
  }

  return {
    show: true,
    left: "center",
    itemWidth: 10,
    itemHeight: 10,
    textStyle: { fontSize: 12, color: _styling.textColor },
    data: series.map((s) => s.subject),
    ...customConfig,
  }
}

/**
 * Creates common grid configuration
 */
export function createGridConfig(
  seriesLength: number,
  _customConfig?: any
): any {
  const legendSpace = Math.min(
    15 + Math.floor((seriesLength - 1) / 15) * 4,
    35
  )

  return {
    left: "3%",
    right: "3%",
    bottom: "10%",
    top: `${legendSpace}%`,
    containLabel: true,
  }
}

/**
 * Creates common series item style with label
 */
export function createSeriesItemStyle(
  value: number,
  showLabels: boolean,
  styling: ChartStyling,
  customLabelConfig?: any
): any {
  return {
    value,
    label: {
      show: showLabels,
      position: "top",
      formatter: "{c}",
      fontSize: 10,
      color: styling.textColor,
      ...customLabelConfig,
    },
  }
}

/**
 * Creates common emphasis configuration
 */
export function createEmphasisConfig(
  focusType: 'series' | 'self' = 'series',
  customConfig?: any
): any {
  return {
    focus: focusType,
    itemStyle: {
      borderWidth: 2,
      borderColor: "#fff",
    },
    ...customConfig,
  }
}
