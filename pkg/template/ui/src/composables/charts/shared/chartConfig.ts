import type { EChartsOption } from 'echarts/types/dist/shared'
import type { ScaleType } from '../../../types'
import { fontSize } from './common'

export interface ChartStyling {
  textColor: string
  axisColor: string
  opacity: number
  backgroundColor: string | undefined
}

/**
 * Gets consistent styling colors based on dark mode
 */
export function getChartStyling(isDark: boolean): ChartStyling {
  return {
    textColor: isDark ? '#e5e7eb' : '#374151',
    axisColor: isDark ? '#4b5563' : '#d1d5db',
    backgroundColor: isDark ? 'transparent' : undefined,
    opacity: isDark ? 0.2 : 0.8,
  }
}

/**
 * Creates common axis configuration
 */
export function createAxisConfig(
  styling: ChartStyling,
  xAxisData: string[],
  scale: ScaleType = 'linear',
  minValue?: number
): { xAxis: any; yAxis: any } {
  const yAxisConfig: any = {
    type: scale === 'log' ? 'log' : 'value',
    logBase: 10,
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
  }

  // For log scale, set a clean minimum to avoid showing 0.1
  if (scale === 'log' && minValue !== undefined) {
    // Round down to nearest power of 10, but minimum is 1
    const minLog = Math.pow(10, Math.floor(Math.log10(minValue)))
    yAxisConfig.min = Math.max(1, minLog)
  }

  return {
    xAxis: {
      type: 'category',
      data: xAxisData,
      axisLabel: {
        interval: 0,
        rotate: xAxisData.reduce((acc, cur) => acc + cur.length, 0) > 100 ? 30 : 0,
        fontSize,
        color: styling.textColor,
      },
      axisLine: {
        lineStyle: { color: styling.axisColor },
      },
    },
    yAxis: yAxisConfig,
  }
}

/**
 * Format a tooltip value, showing 0 for null/undefined values
 */
function formatTooltipValue(value: any): string {
  if (value === null || value === undefined) return '0'
  return String(value)
}

/**
 * Creates common tooltip configuration
 * @param hasXYAxis - Whether the chart has both X and Y axes
 * @param seriesCount - Number of series in the chart (defaults to 1)
 */
export function createTooltipConfig(hasXYAxis: boolean, seriesCount = 1): EChartsOption['tooltip'] {
  // Use item trigger if there are too many series (>10) to avoid overwhelming tooltip
  if (hasXYAxis && seriesCount <= 10) {
    return {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter: (params) => {
        if (!Array.isArray(params)) return ''

        return params.reduce(
          (acc, cur) => `${acc}${cur.marker} ${cur.seriesName}: ${formatTooltipValue(cur.value)}<br/>`,
          `<strong>${params[0]?.name}</strong><br/>`
        )
      },
    }
  }

  return {
    trigger: 'item',
    formatter: (params) => {
      if (Array.isArray(params)) return ''
      let { name, seriesName } = params

      if (hasXYAxis && seriesName) {
        name = seriesName
      }

      if (!name && seriesName) {
        name = seriesName
      }

      return `${params.marker} <strong>${name}</strong><br/>${formatTooltipValue(params.value)}`
    },
  }
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
    return { show: false }
  }

  return {
    show: true,
    left: 'center',
    itemWidth: 10,
    itemHeight: 10,
    textStyle: { fontSize, color: styling.textColor },
    data: series.map((s) => s.xAxis),
    ...customConfig,
  }
}

/**
 * Creates common grid configuration
 */
export function createGridConfig(seriesLength = 1): any {
  const legendSpace = Math.min(15 + Math.floor((seriesLength - 1) / 15) * 2, 35)

  return {
    left: '3%',
    right: '3%',
    bottom: '10%',
    top: `${legendSpace}%`,
    containLabel: true,
  }
}

export const createLabelConfig = (showLabels: boolean, styling: ChartStyling) => ({
  show: showLabels,
  position: 'top' as const,
  formatter: '{c}',
  fontSize,
  color: styling.textColor,
})
