import type { EChartsOption } from 'echarts/types/dist/shared'
import type { ScaleType } from '@/types'
import { fontSize } from './common'
import { describe } from '@/lib/stats'

export const LARGE_X_THRESHOLD = 50

export function isLargeXAxis(xAxisData: string[]): boolean {
  return xAxisData.length > LARGE_X_THRESHOLD
}

// Point count past which bar/line series switch on ECharts' large-data path
// (`large: true`). Below it, normal rendering keeps full per-item interactivity;
// above it the optimized path keeps a 100k-point dataset's draw on one frame.
export const LARGE_DATA_THRESHOLD = 2000

export function createHeatmapDataZoomConfig(
  largeX: boolean,
  largeY: boolean,
  xLen: number,
  yLen: number,
  styling: ChartStyling
): any[] {
  const result: any[] = []
  if (largeX) {
    const end = Math.max(5, Math.ceil((30 / xLen) * 100))
    result.push(
      { type: 'inside', xAxisIndex: 0, start: 0, end, filterMode: 'filter' },
      {
        type: 'slider',
        xAxisIndex: 0,
        start: 0,
        end,
        bottom: 55,
        height: 28,
        filterMode: 'filter',
        textStyle: { color: styling.textColor },
      }
    )
  }
  if (largeY) {
    const end = Math.max(5, Math.ceil((30 / yLen) * 100))
    result.push(
      { type: 'inside', yAxisIndex: 0, start: 0, end, filterMode: 'filter' },
      {
        type: 'slider',
        yAxisIndex: 0,
        start: 0,
        end,
        left: 10,
        width: 20,
        filterMode: 'filter',
        textStyle: { color: styling.textColor },
      }
    )
  }
  return result
}

export function createDataZoomConfig(xAxisData: string[], styling: ChartStyling): any[] {
  const end = Math.max(5, Math.ceil((30 / xAxisData.length) * 100))
  return [
    { type: 'inside', start: 0, end },
    // Slider sits between the (auto-thinned) tick labels and the category-axis
    // name, which is pushed below it via a larger nameGap. Heights coordinated
    // with createGridConfig's fixed px bottom so spacing is stable across sizes.
    // textStyle colors the left/right boundary labels to match the theme text
    // (ECharts' default gray is too dim in dark mode).
    {
      type: 'slider',
      start: 0,
      end,
      bottom: 34,
      height: 28,
      textStyle: { color: styling.textColor },
    },
  ]
}

export interface ChartStyling {
  textColor: string
  axisColor: string
  opacity: number
  backgroundColor: string | undefined
}

export function createToolboxConfig(isDark: boolean, title: string, pixelRatio: number): any {
  const { textColor } = getChartStyling(isDark)
  return {
    show: true,
    feature: {
      saveAsImage: {
        show: true,
        type: 'jpeg',
        title: 'Save',
        pixelRatio,
        name: title,
      },
    },
    iconStyle: { borderColor: textColor },
    emphasis: { iconStyle: { borderColor: textColor } },
  }
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
            // Large axis: extra gap drops the name below the dataZoom slider.
            nameGap: isLargeXAxis(xAxisData) ? 88 : 45,
            nameTextStyle: { color: styling.textColor, fontSize, fontWeight: 'bold' },
          }
        : {}),
      axisLabel: {
        // Large axis: let ECharts auto-thin ticks (dataZoom drives navigation)
        // so labels don't overlap. Small axis: force every label visible.
        interval: isLargeXAxis(xAxisData) ? 'auto' : 0,
        rotate: isLargeXAxis(xAxisData)
          ? 0
          : xAxisData.reduce((acc, cur) => acc + cur.length, 0) > 100
            ? 30
            : 0,
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

/** Item-trigger radar tooltip: spoke rows + Σ + spread + donut (parity with 3D/heatmap). */
export function formatRadarItemTooltip(
  params: {
    data?: { name?: string; value?: number[] }
    name?: string
    seriesName?: string
    color?: string
  },
  indicatorNames: string[],
  isDark: boolean,
  colorFor?: (name: string) => string | undefined
): string {
  if (!params?.data) return ''

  const vals: number[] = Array.isArray(params.data.value) ? params.data.value : []
  const dataName = params.data.name
  const seriesName = params.seriesName

  const title =
    seriesName && dataName && seriesName !== dataName
      ? `${seriesName} / ${dataName}`
      : (dataName ?? seriesName ?? params.name ?? '')

  const seriesColor = typeof params.color === 'string' ? params.color : undefined
  const color = (name: string) => colorFor?.(name) ?? seriesColor ?? '#888'

  const rows = indicatorNames
    .map((name, i) => {
      const dot = `<span style="display:inline-block;width:10px;height:10px;border-radius:50%;background:${color(name)};margin-right:6px"></span>`
      return `${dot}${name}: <b>${formatTooltipValue(vals[i])}</b>`
    })
    .join('<br/>')

  const finiteVals = vals.filter((v) => Number.isFinite(v))
  const sigmaBlock =
    finiteVals.length >= 2
      ? `${tooltipDivider(isDark)}Σ ${title}: <b>${formatTooltipValue(finiteVals.reduce((a, b) => a + b, 0))}</b>`
      : ''

  const spread = tooltipSpreadRows(vals, isDark)

  const donut = renderDonutSvg(
    indicatorNames.map((name, i) => ({
      name,
      value: typeof vals[i] === 'number' ? vals[i]! : 0,
      color: color(name),
    }))
  )
  const donutBlock = donut ? `${tooltipDivider(isDark)}${donut}` : ''

  return `<b>${title}</b><br/>${rows}${sigmaBlock}${spread}${donutBlock}`
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

// Lean spread summary for the hovered vector (the series values at a category /
// the z-values in a 3D cell): median, IQR, CV. The full descriptive set lives in
// the stats panel — this is just the at-a-glance trio. Returns '' for <2 finite
// values (a single value has no spread). Shared by the 2D and 3D tooltips.
export function tooltipSpreadRows(values: number[], isDark: boolean): string {
  const finite = values.filter((v) => Number.isFinite(v))
  if (finite.length < 2) return ''
  const s = describe(finite)
  const num = (v: number) => (Number.isFinite(v) ? Math.round(v * 100) / 100 : '—')
  const cv = Number.isFinite(s.cv) ? `${(s.cv * 100).toFixed(1)}%` : '—'
  return (
    `${tooltipDivider(isDark)}` +
    `Median: <b>${num(s.median)}</b><br/>` +
    `IQR: <b>${num(s.iqr)}</b><br/>` +
    `CV: <b>${cv}</b>`
  )
}

// Inline SVG donut + side legend (swatch, name, %) for tooltips. Non-positive
// slices are dropped (can't show negative share). Returns '' when fewer than 2
// positive slices — nothing to compare.
export function renderDonutSvg(
  slices: { value: number; color: string; name: string }[],
  size = 96
): string {
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
    const s1 = xy(outerR, start),
      e1 = xy(outerR, end)
    const s2 = xy(innerR, end),
      e2 = xy(innerR, start)
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

  const svg = `<svg width="${size}" height="${size}" viewBox="0 0 ${size} ${size}" style="flex:none">${paths.join('')}</svg>`

  // Side legend: one row per slice (swatch + name + share %). Scrolls with the
  // tooltip when there are many slices.
  const legendRows = pos
    .map((s) => {
      const pct = ((s.value / total) * 100).toFixed(1)
      const swatch = `<span style="display:inline-block;width:10px;height:10px;border-radius:2px;background:${s.color};margin-right:6px;flex:none"></span>`
      return `<div style="display:flex;align-items:center;white-space:nowrap">${swatch}<span>${s.name}</span><b style="margin-left:6px">${pct}%</b></div>`
    })
    .join('')
  // Flow into balanced columns when there are many slices, so a long legend
  // stays compact instead of one tall stack. ≤12 slices → single column.
  const cols = Math.ceil(pos.length / 12)
  const rowsPerCol = Math.ceil(pos.length / cols)
  const legend = `<div style="display:grid;grid-auto-flow:column;grid-template-rows:repeat(${rowsPerCol},auto);gap:2px 12px;font-size:11px">${legendRows}</div>`

  return `<div style="display:flex;align-items:center;gap:8px;margin-top:4px">${svg}${legend}</div>`
}

/**
 * Single-series axis tooltip pinned near the top, following the cursor's x.
 * Used by the no-Y line chart (only an x axis): with dense/large data an
 * item-trigger tooltip is hard to hit, so we trigger on the axis and pin the
 * box to a fixed height so it never jumps over the line.
 */
export function createPinnedAxisTooltip(isDark = false): EChartsOption['tooltip'] {
  const theme = getTooltipTheme(isDark)
  return {
    trigger: 'axis',
    position: (pt: number[]) => [pt[0] ?? 0, '10%'],
    ...theme,
    formatter: (params) => {
      if (!Array.isArray(params)) return ''
      const p = params[0]
      if (!p) return ''
      return `<strong>${p.name}</strong><br/>${p.marker} ${formatTooltipValue(p.value)}`
    },
  }
}

/**
 * Creates common tooltip configuration
 * @param hasXYAxis - Whether the chart has both X and Y axes
 * @param isDark - Dark mode flag for tooltip theming
 * @param seriesTotals - Per-series totals across all x, appended after each name
 */
export function createTooltipConfig(
  hasXYAxis: boolean,
  isDark = false,
  seriesTotals?: Map<string, number>
): EChartsOption['tooltip'] {
  const theme = getTooltipTheme(isDark)

  if (hasXYAxis) {
    return {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      // Many-series charts can produce a long row list + donut legend; cap the
      // height and let the user scroll into the tooltip rather than overflow.
      enterable: true,
      extraCssText: 'max-height:60vh;overflow:auto;',
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
        const spread = tooltipSpreadRows(
          params.map((p) => (typeof p.value === 'number' ? p.value : NaN)),
          isDark
        )
        const donut = renderDonutSvg(
          params.map((p) => ({
            value: typeof p.value === 'number' ? p.value : 0,
            color: typeof p.color === 'string' ? p.color : String(p.color),
            name: p.seriesName ?? '',
          }))
        )
        return `${body}${sumLine}${spread}${donut ? tooltipDivider(isDark) + donut : ''}`
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

export function createGridConfig(seriesLength = 1, hasDataZoom = false): any {
  const legendSpace = Math.min(15 + Math.floor((seriesLength - 1) / 15) * 2, 35)

  // With a dataZoom slider we turn containLabel OFF and reserve label space in
  // fixed px. containLabel pins both the tick labels and the axis name to the
  // container bottom, where they collide with a bottom-pinned slider. Fixed px
  // lets the axis line sit at exactly `bottom`, so the slider (bottom:34) and
  // the axis name (nameGap:88 → ~8px from bottom) land in predictable, non-
  // overlapping bands: ticks → slider → name, top to bottom.
  if (hasDataZoom) {
    return {
      left: 55,
      right: 24,
      bottom: 100,
      top: `${legendSpace}%`,
      containLabel: false,
    }
  }

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
