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
  minValue?: number,
  xAxisName?: string
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
      // Group/axis label (e.g. "category") under the category axis, when known.
      // Bold + extra gap so it clears the (often rotated) tick labels.
      ...(xAxisName
        ? {
            name: xAxisName,
            nameLocation: 'middle',
            nameGap: 45,
            nameTextStyle: { color: styling.textColor, fontSize, fontWeight: 'bold' },
          }
        : {}),
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
export function formatTooltipValue(value: any): string {
  if (value === null || value === undefined) return '0'
  if (typeof value === 'number') return String(Math.round(value * 100) / 100)
  return String(value)
}

/**
 * Shared tooltip box theme (background, border, text) for light/dark mode.
 * Single source so 2D and 3D tooltips look identical.
 */
export function getTooltipTheme(isDark: boolean) {
  return {
    backgroundColor: isDark ? '#1f2937' : '#ffffff',
    borderColor: isDark ? '#4b5563' : '#e5e7eb',
    textStyle: { color: getChartStyling(isDark).textColor },
  }
}

/** Horizontal divider used inside tooltip HTML, themed to match the box. */
export function tooltipDivider(isDark: boolean): string {
  return `<hr style="border:none;border-top:1px solid ${getChartStyling(isDark).axisColor};margin:4px 0"/>`
}

// Inline SVG donut for tooltips. Non-positive slices are dropped (can't show
// negative share). Returns '' when fewer than 2 positive slices — nothing to compare.
export function renderDonutSvg(slices: { value: number; color: string }[], size = 96): string {
  const pos = slices.filter((s) => s.value > 0)
  if (pos.length < 2) return ''
  const total = pos.reduce((s, p) => s + p.value, 0)
  if (total <= 0) return ''

  const cx = size / 2
  const cy = size / 2
  const outerR = size / 2 - 2
  const innerR = outerR * 0.55
  const GAP = 1.5

  const xy = (r: number, deg: number) => {
    const rad = ((deg - 90) * Math.PI) / 180
    return { x: cx + r * Math.cos(rad), y: cy + r * Math.sin(rad) }
  }

  const path = (start: number, end: number) => {
    const s1 = xy(outerR, start), e1 = xy(outerR, end)
    const s2 = xy(innerR, end),   e2 = xy(innerR, start)
    const lg = end - start > 180 ? 1 : 0
    const f = (n: number) => n.toFixed(2)
    return [
      `M${f(s1.x)},${f(s1.y)}`,
      `A${outerR},${outerR},0,${lg},1,${f(e1.x)},${f(e1.y)}`,
      `L${f(s2.x)},${f(s2.y)}`,
      `A${innerR.toFixed(2)},${innerR.toFixed(2)},0,${lg},0,${f(e2.x)},${f(e2.y)}`,
      'Z',
    ].join(' ')
  }

  let angle = 0
  const paths = pos.map((s) => {
    const sweep = (s.value / total) * 360
    const start = angle + GAP / 2
    const end = angle + sweep - GAP / 2
    angle += sweep
    return `<path d="${path(start, end)}" fill="${s.color}"/>`
  })

  return `<svg width="${size}" height="${size}" viewBox="0 0 ${size} ${size}" style="display:block;margin:4px auto">${paths.join('')}</svg>`
}

/**
 * Creates common tooltip configuration
 * @param hasXYAxis - Whether the chart has both X and Y axes
 * @param seriesCount - Number of series in the chart (defaults to 1)
 * @param isDark - Dark mode flag for tooltip theming
 * @param seriesTotals - Per-series totals across all x, appended after each name
 */
export function createTooltipConfig(
  hasXYAxis: boolean,
  seriesCount = 1,
  isDark = false,
  seriesTotals?: Map<string, number>
): EChartsOption['tooltip'] {
  const theme = getTooltipTheme(isDark)

  // Use item trigger if there are too many series (>10) to avoid overwhelming tooltip
  if (hasXYAxis && seriesCount <= 10) {
    return {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      ...theme,
      formatter: (params) => {
        if (!Array.isArray(params)) return ''

        const body = params.reduce((acc, cur) => {
          const seriesSum = seriesTotals?.get(cur.seriesName ?? '')
          const sumTag = seriesSum === undefined ? '' : ` (Σ${Math.round(seriesSum * 100) / 100})`
          return `${acc}${cur.marker} ${cur.seriesName}${sumTag}: ${formatTooltipValue(cur.value)}<br/>`
        }, `<strong>${params[0]?.name}</strong><br/>`)

        // Σ across all series at this x = the x marginal. Label with the x name
        // to match the 3D tooltip's "Σ <name>" lines. Only meaningful with >1 series.
        if (params.length <= 1) return body
        const total = params.reduce(
          (sum, cur) => sum + (typeof cur.value === 'number' ? cur.value : 0),
          0
        )
        const xName = params[0]?.name ?? ''
        const sumLine = `${tooltipDivider(isDark)}Σ ${xName}: <b>${Math.round(total * 100) / 100}</b>`
        const donut = renderDonutSvg(
          params.map((p) => ({
            value: typeof p.value === 'number' ? p.value : 0,
            color: typeof p.color === 'string' ? p.color : String(p.color),
          }))
        )
        return `${body}${sumLine}${donut ? tooltipDivider(isDark) + donut : ''}`
      },
    }
  }

  return {
    trigger: 'item',
    ...theme,
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
// Header centered above the legend, naming the dimension the legend encodes
// (the y group on 2D bar/line). ECharts legends have no native title.
export function makeLegendTitle(text: string, styling: ChartStyling): any {
  return {
    text,
    left: 'center',
    top: 0,
    textStyle: { color: styling.textColor, fontSize, fontWeight: 'bold' },
  }
}

export function createGridConfig(seriesLength = 1): any {
  const legendSpace = Math.min(15 + Math.floor((seriesLength - 1) / 15) * 2, 35)

  return {
    left: '3%',
    right: '3%',
    // Extra bottom room so a category-axis name (group label) isn't clipped.
    bottom: '13%',
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
