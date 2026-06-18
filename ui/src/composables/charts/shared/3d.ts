import type { Render3D, Point3D } from '../../../types'
import { getNextColorFor } from '../../../lib/utils'
import { tooltipDivider, tooltipSpreadRows, renderDonutSvg, type ChartStyling } from './chartConfig'

export const round2 = (v: number) => Math.round(v * 100) / 100

/** Blue-to-red gradient for value-mode 3D visualMap (metric height). */
export const VALUE_3D_COLOR_RANGE = [
  '#313695',
  '#4575b4',
  '#74add1',
  '#abd9e9',
  '#e0f3f8',
  '#ffffbf',
  '#fee090',
  '#fdae61',
  '#f46d43',
  '#d73027',
  '#a50026',
]

export function maxFrom3DData(series: { data: { value: number[] }[] }[]): number {
  let max = 0
  for (const s of series) {
    for (const item of s.data) {
      const v = item.value[2] ?? 0
      if (v > max) max = v
    }
  }
  return max
}

export function create3DVisualMap(max: number, styling: ChartStyling) {
  return {
    show: true,
    min: 0,
    max: max || 1,
    dimension: 2,
    calculable: true,
    orient: 'vertical' as const,
    right: '0%',
    top: 'center',
    inRange: { color: VALUE_3D_COLOR_RANGE },
    textStyle: { color: styling.textColor },
  }
}

/** ECharts merge keeps omitted visualMap — pass `[]` when off (with replaceMerge). */
export function resolve3DVisualMap(
  enabled: boolean,
  series: { data: { value: number[] }[] }[],
  styling: ChartStyling
) {
  return enabled ? create3DVisualMap(maxFrom3DData(series), styling) : []
}

/** @deprecated Use create3DVisualMap */
export const createValue3DVisualMap = create3DVisualMap

export function createValue3DTooltipFormatter(params: {
  xValues: string[]
  yValues: string[]
  xAxisLabel?: string
  yAxisLabel?: string
  valueLabel?: string
}) {
  const { xValues, yValues, xAxisLabel, yAxisLabel, valueLabel } = params
  const xLabel = xAxisLabel ?? 'x'
  const yLabel = yAxisLabel ?? 'y'
  const zLabel = valueLabel ?? 'value'

  return (p: { value: number[] }) => {
    const [xi = 0, yi = 0, v = 0] = p.value
    const xName = xValues[xi] ?? String(xi)
    const yName = yValues[yi] ?? String(yi)
    return `<b>${xLabel}: ${xName}</b><br/>${yLabel}: ${yName}<br/>${zLabel}: <b>${round2(v)}</b>`
  }
}

/** grid3D boxWidth / boxDepth tier from axis category count (max 200). */
export function boxSizeForAxisCount(len: number): number {
  if (len < 5) return 80
  if (len < 15) return 100
  return 200
}

// Empty render payload when a chart lacks precomputed 3D data (shouldn't happen:
// the worker attaches render3D to every 3D chart).
export const EMPTY_RENDER: Render3D = {
  xValues: [],
  yValues: [],
  zValues: [],
  barSeries: [],
  lineSeries: [],
  cellTotals: {},
}

export function makeAxis3DCommon(styling: ChartStyling) {
  return {
    axisLabel: { color: styling.textColor },
    axisLine: { lineStyle: { color: styling.axisColor } },
  }
}

// Axis name block for a 3D axis (bold + gap so it clears the tick labels and
// reads as a title, not another tick). Returns {} when no label is known.
export function axis3DName(label: string | undefined, styling: ChartStyling) {
  if (!label) return {}
  return {
    name: label,
    nameGap: 25,
    nameTextStyle: { color: styling.textColor, fontSize: 14, fontWeight: 'bold' as const },
  }
}

/**
 * Pure factory for the complex per-cell 3D tooltip.
 * Shared by bar3D and line3D so the aggregation, Σz, marginals, spread, and donut
 * logic never diverges.
 */
export function create3DTooltipFormatter(params: {
  xValues: string[]
  yValues: string[]
  zValues: string[]
  aggPoints: Point3D[]
  isDark: boolean
  xAxisLabel?: string
  yAxisLabel?: string
  zAxisLabel?: string
}) {
  const { xValues, yValues, zValues, aggPoints, isDark, xAxisLabel, yAxisLabel, zAxisLabel } =
    params
  const xLabel = xAxisLabel ?? 'x'
  const yLabel = yAxisLabel ?? 'y'
  const zSumLabel = zAxisLabel ?? 'z'

  return (p: { value: number[] }) => {
    const [xi = 0, yi = 0] = p.value
    const xName = xValues[xi]
    const yName = yValues[yi]

    // Aggregate on demand for the hovered cell only — one pass over the visible
    // points yields this cell's per-z breakdown, its Σz, and the x/y marginals.
    // Hovers are rare next to cell count, so computing here beats precomputing
    // every cell's sums up front.
    const zmap = new Map<string, number>()
    let cellTotal = 0
    let xMarginal = 0
    let yMarginal = 0
    for (const pt of aggPoints) {
      const onX = pt.xAxis === xName
      const onY = pt.yAxis === yName
      if (onX) xMarginal += pt.value
      if (onY) yMarginal += pt.value
      if (onX && onY) {
        zmap.set(pt.zAxis, (zmap.get(pt.zAxis) ?? 0) + pt.value)
        cellTotal += pt.value
      }
    }

    const rows = zValues
      .map((z) => {
        const v = zmap.get(z)
        if (v === undefined) return ''
        const dot = `<span style="display:inline-block;width:10px;height:10px;border-radius:50%;background:${getNextColorFor(z)};margin-right:6px"></span>`
        return `${dot}${z}: <b>${round2(v)}</b>`
      })
      .filter(Boolean)

    // Σ over z = stacked bar height at this (x,y). First line under the divider,
    // above the x/y marginals, when there's more than one z to sum.
    const zSumLine = zmap.size > 1 ? `Σ ${zSumLabel}: <b>${round2(cellTotal)}</b><br/>` : ''

    // Marginal totals: sum over the other two axes for this x / this y.
    const margins =
      tooltipDivider(isDark) +
      zSumLine +
      `Σ ${xLabel}(${xName}): <b>${round2(xMarginal)}</b><br/>` +
      `Σ ${yLabel}(${yName}): <b>${round2(yMarginal)}</b>`

    // Spread of the z-values in this cell (median / IQR / CV), mirroring the 2D
    // tooltip. Only meaningful with >1 z.
    const spread = tooltipSpreadRows(Array.from(zmap.values()), isDark)

    const donut =
      zmap.size > 1
        ? renderDonutSvg(
            zValues
              .filter((z) => zmap.has(z))
              .map((z) => ({ value: zmap.get(z)!, color: getNextColorFor(z) ?? '', name: z }))
          )
        : ''
    return `<b>${xLabel}: ${xName} / ${yLabel}: ${yName}</b><br/>${rows.join('<br/>')}${margins}${spread}${donut ? tooltipDivider(isDark) + donut : ''}`
  }
}

/**
 * Z-axis legend config used by both 3D bar and line when there are multiple z groups.
 * Selection state is passed through so legend toggles survive recomputes.
 */
export function createZLegendConfig(
  zValues: string[],
  styling: ChartStyling,
  selected: Record<string, boolean>
) {
  return {
    show: zValues.length > 1,
    orient: 'vertical',
    left: 'left',
    top: 'middle',
    textStyle: { color: styling.textColor },
    // Pin palette swatches so visualMap (value-based bar colors) does not
    // rewrite legend markers; tooltips use the same getNextColorFor keys.
    data: zValues.map((z) => ({
      name: z,
      itemStyle: { color: getNextColorFor(z) },
    })),
    // Controlled selection: persist toggles across recomputes (without this,
    // re-applying the option would reset every z back to visible).
    selected,
  }
}

/**
 * Build the common grid3D block. The caller supplies the one or two differing
 * pieces (autoRotate + optional orthographic projection for line3D).
 */
export function create3DGridConfig(opts: {
  styling: ChartStyling
  autoRotate: boolean
  orthographic?: boolean
  xCount: number
  yCount: number
}) {
  const { styling, autoRotate, orthographic, xCount, yCount } = opts
  const xWidth = boxSizeForAxisCount(xCount)
  const yWidth = boxSizeForAxisCount(yCount)

  return {
    boxWidth: xWidth,
    boxDepth: yWidth,
    axisLine: { lineStyle: { color: styling.axisColor } },
    splitLine: { lineStyle: { color: styling.axisColor, opacity: styling.opacity } },
    viewControl: {
      distance: xWidth + yWidth,
      // `autoRotate` is optional on BaseChartConfig (relaxed in Task 7) —
      // pie/heatmap/radar pass a config without it. Default to off at the
      // call site; 3D bar/line are the only consumers.
      autoRotate,
      ...(orthographic ? { projection: 'orthographic' } : {}),
    },
    light: {
      main: { intensity: 0.3, shadow: false },
      ambient: { intensity: 0.9 },
    },
  }
}

/**
 * Label config for the "top of stack" cell total.
 * Used both for the topmost bar3D series label and for the scatter3D overlay label on line3D.
 */
export function create3DCellLabel(
  show: boolean,
  cellTotals: Record<string, number>,
  textColor: string
) {
  if (!show) return { show: false }
  return {
    show: true,
    formatter: (p: { value: number[] }) => {
      const [xi = 0, yi = 0] = p.value
      const total = cellTotals[`${xi},${yi}`]
      return total === undefined ? '' : String(round2(total))
    },
    textStyle: { fontSize: 12, color: textColor },
  }
}
