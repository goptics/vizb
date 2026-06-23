import type { EChartsOption } from 'echarts'
import type { Render3D, Point3D, ScaleType, Series3DData } from '@/types'
import { COLOR_PALETTE, getNextColorFor } from '@/lib/utils'
import {
  formatTooltipValue,
  tooltipDivider,
  tooltipSpreadRows,
  renderDonutSvg,
  getTooltipTheme,
  type ChartStyling,
} from './chartConfig'

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
  seriesData: { value: number[] }[]
  isDark: boolean
  xAxisLabel?: string
  yAxisLabel?: string
  valueLabel?: string
  seriesColor?: string
}) {
  const {
    xValues,
    yValues,
    seriesData,
    isDark,
    xAxisLabel,
    yAxisLabel,
    valueLabel,
    seriesColor = COLOR_PALETTE[0]!,
  } = params
  const xLabel = xAxisLabel ?? 'x'
  const yLabel = yAxisLabel ?? 'y'
  const metricLabel = valueLabel ?? 'value'

  const xMarginals = new Map<string, number>()
  const yMarginals = new Map<string, number>()
  for (const item of seriesData) {
    const [xi = 0, yi = 0, v = 0] = item.value
    const xName = xValues[xi] ?? String(xi)
    const yName = yValues[yi] ?? String(yi)
    xMarginals.set(xName, (xMarginals.get(xName) ?? 0) + v)
    yMarginals.set(yName, (yMarginals.get(yName) ?? 0) + v)
  }

  const dot = `<span style="display:inline-block;width:10px;height:10px;border-radius:50%;background:${seriesColor};margin-right:6px"></span>`

  return (p: { value: number[] }) => {
    const [xi = 0, yi = 0, v = 0] = p.value
    const xName = xValues[xi] ?? String(xi)
    const yName = yValues[yi] ?? String(yi)
    const xMarginal = xMarginals.get(xName) ?? 0
    const yMarginal = yMarginals.get(yName) ?? 0
    const xyTotal = xMarginal + yMarginal

    const margins =
      tooltipDivider(isDark) +
      `Σ ${xLabel}(${xName}): <b>${round2(xMarginal)}</b><br/>` +
      `Σ ${yLabel}(${yName}): <b>${round2(yMarginal)}</b><br/>` +
      `Σ (${xLabel}+${yLabel}): <b>${round2(xyTotal)}</b>`

    return (
      `<b>${xLabel}: ${xName} / ${yLabel}: ${yName}</b><br/>` +
      `${dot}${metricLabel}: <b>${round2(v)}</b><br/>` +
      margins
    )
  }
}

/** grid3D boxWidth / boxDepth tier from axis category count (max 200). */
export function boxSizeForAxisCount(len: number): number {
  if (len < 5) return 80
  if (len < 15) return 100
  return 200
}

/** Fixed grid footprint for pseudo value 3D (`--3d` on x+y category data). */
export const VALUE_MODE_3D_BOX_SIZE = 100
export const VALUE_MODE_3D_VIEW_DISTANCE = 200

/** Largest grid3D edge — used to scale camera framing. */
export function boxExtent3D(boxWidth: number, boxDepth: number, boxHeight: number): number {
  return Math.max(boxWidth, boxDepth, boxHeight)
}

/**
 * Perspective camera distance (bar3D). ECharts default is 200 for a 100³ box → 2× extent.
 */
export function viewDistanceFor3DBox(
  boxWidth: number,
  boxDepth: number,
  boxHeight: number
): number {
  return boxExtent3D(boxWidth, boxDepth, boxHeight) * 2
}

/**
 * Orthographic viewing volume (scatter3D / line3D). ECharts default is 150 for a 100³ box → 1.5× extent.
 */
export function orthographicSizeFor3DBox(
  boxWidth: number,
  boxDepth: number,
  boxHeight: number
): number {
  return boxExtent3D(boxWidth, boxDepth, boxHeight) * 1.5
}

const BAND_FILL_MIN = 0.45
const BAND_FILL_MAX = 0.92
const BAND_FILL_LO = 2
const BAND_FILL_HI = 40

/** Fraction of each category band filled by bars/markers (sparse → low, dense → high). */
export function bandFillRatioForCount(count: number): number {
  const c = Math.max(1, count)
  if (c <= BAND_FILL_LO) return BAND_FILL_MIN
  if (c >= BAND_FILL_HI) return BAND_FILL_MAX
  const t = (c - BAND_FILL_LO) / (BAND_FILL_HI - BAND_FILL_LO)
  return BAND_FILL_MIN + t * (BAND_FILL_MAX - BAND_FILL_MIN)
}

function clamp3DSpacing(v: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, v))
}

/** bar3D footprint on category grids — replaces ECharts' fixed bandWidth × 0.7 default. */
export function barSizeFor3DGrid(
  xCount: number,
  yCount: number,
  boxWidth: number,
  boxDepth: number
): [number, number] {
  const ratio = bandFillRatioForCount(Math.max(xCount, yCount, 1))
  const bandX = boxWidth / Math.max(xCount, 1)
  const bandY = boxDepth / Math.max(yCount, 1)
  return [bandX * ratio, bandY * ratio]
}

/** scatter3D marker size on category grids — diameter tracks cell size × fill ratio. */
export function symbolSizeFor3DGrid(
  xCount: number,
  yCount: number,
  boxWidth: number,
  boxDepth: number
): number {
  const ratio = bandFillRatioForCount(Math.max(xCount, yCount, 1))
  const minBand = Math.min(boxWidth / Math.max(xCount, 1), boxDepth / Math.max(yCount, 1))
  return clamp3DSpacing(minBand * ratio, 2, 24)
}

/** bar3D footprint for continuous value-axis point clouds. */
export function barSizeForContinuous3D(
  pointCount: number,
  boxWidth: number,
  boxDepth: number
): [number, number] {
  const n = Math.max(pointCount, 1)
  const ratio = bandFillRatioForCount(n)
  const baseX = Math.round(boxWidth / Math.sqrt(n))
  const baseY = Math.round(boxDepth / Math.sqrt(n))
  return [baseX * ratio, baseY * ratio]
}

/** scatter3D marker size for continuous value-axis point clouds. */
export function symbolSizeForContinuous3D(
  pointCount: number,
  boxWidth: number,
  boxDepth: number
): number {
  const { xCount, yCount } = continuous3DGridCounts(pointCount)
  return symbolSizeFor3DGrid(xCount, yCount, boxWidth, boxDepth)
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

/** Value axes for --axes x,y,z continuous 3D (swap-driven, not category indices). */
export function createContinuous3DAxes(
  styling: ChartStyling,
  xLabel?: string,
  yLabel?: string,
  zLabel?: string,
  scale: ScaleType = 'linear'
) {
  const valueType = scale === 'log' ? ('log' as const) : ('value' as const)
  const log = scale === 'log' ? { logBase: 10 } : {}
  const axisCommon = makeAxis3DCommon(styling)
  return {
    xAxis3D: {
      type: valueType,
      ...log,
      ...axisCommon,
      ...axis3DName(xLabel, styling),
    },
    yAxis3D: {
      type: valueType,
      ...log,
      ...axisCommon,
      ...axis3DName(yLabel, styling),
    },
    zAxis3D: {
      type: valueType,
      ...log,
      ...axisCommon,
      ...axis3DName(zLabel, styling),
    },
  }
}

export function createContinuous3DTooltipFormatter(
  _isDark: boolean,
  labels: { x?: string; y?: string; z?: string }
) {
  const xName = labels.x ?? 'x'
  const yName = labels.y ?? 'y'
  const zName = labels.z ?? 'z'
  return (p: { value: number[] }) => {
    const [x = 0, y = 0, z = 0] = p.value
    return (
      `<b>${xName}: ${formatTooltipValue(x)}</b><br/>` +
      `${yName}: ${formatTooltipValue(y)}<br/>` +
      `${zName}: ${formatTooltipValue(z)}`
    )
  }
}

export function continuous3DGridCounts(pointCount: number): { xCount: number; yCount: number } {
  const tier = pointCount < 50 ? 10 : pointCount < 500 ? 50 : 100
  return { xCount: tier, yCount: tier }
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

  // Per-legend total across all visible cells — mirrors 2D "(Σtotal)" tag.
  const legendTotals = new Map<string, number>()
  for (const pt of aggPoints) {
    legendTotals.set(pt.zAxis, (legendTotals.get(pt.zAxis) ?? 0) + pt.value)
  }

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
        const zTotal = legendTotals.get(z)
        const sumTag = zTotal !== undefined ? ` (Σ${round2(zTotal)})` : ''
        return `${dot}${z}${sumTag}: <b>${round2(v)}</b>`
      })
      .filter(Boolean)

    // Σ over z = stacked bar height at this (x,y), when there's more than one z.
    const zSumLine = zmap.size > 1 ? `Σ ${zSumLabel}: <b>${round2(cellTotal)}</b><br/>` : ''

    const xyTotal = xMarginal + yMarginal
    const xyzTotal = xyTotal + cellTotal

    // Per-axis marginals first, then combined axis sums.
    const margins =
      tooltipDivider(isDark) +
      `Σ ${xLabel}(${xName}): <b>${round2(xMarginal)}</b><br/>` +
      `Σ ${yLabel}(${yName}): <b>${round2(yMarginal)}</b><br/>` +
      zSumLine +
      `Σ (${xLabel}+${yLabel}): <b>${round2(xyTotal)}</b><br/>` +
      `Σ (${xLabel}+${yLabel}+${zSumLabel}): <b>${round2(xyzTotal)}</b>`

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
    const legendBlock = rows.length > 0 ? `${rows.join('<br/>')}<br/>` : ''

    return `<b>${xLabel}: ${xName} / ${yLabel}: ${yName}</b><br/>${legendBlock}${margins}${spread}${donut ? tooltipDivider(isDark) + donut : ''}`
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

export type Continuous3DParams = {
  base: EChartsOption
  styling: ChartStyling
  isDark: boolean
  showLabels: boolean
  useVisualMap: boolean
  defaultColor: string
  threeDRotate: boolean
  scale: ScaleType
  seriesData: Series3DData[]
  axisLabels?: { x?: string; y?: string; z?: string }
}

export type Continuous3DContext = Omit<Continuous3DParams, 'seriesData'>

export function valuePoints3DToSeries(
  points: [number, number, number][],
  title: string
): Series3DData[] {
  return [{ name: title, data: points.map(([x, y, z]) => ({ value: [x, y, z] })) }]
}

export function makeContinuous3DParams(
  ctx: Continuous3DContext,
  seriesData: Series3DData[]
): Continuous3DParams {
  return { ...ctx, seriesData }
}

/** Continuous [x,y,z] point cloud rendered as scatter3D, line3D, or bar3D. */
export function buildContinuous3DOptions(
  params: Continuous3DParams,
  seriesType: 'scatter3D' | 'line3D' | 'bar3D' = 'scatter3D'
): EChartsOption {
  const {
    base,
    styling,
    isDark,
    showLabels,
    useVisualMap,
    defaultColor,
    threeDRotate,
    scale,
    seriesData,
    axisLabels,
  } = params
  const pointCount = seriesData[0]?.data.length ?? 0
  const { xCount, yCount } = continuous3DGridCounts(pointCount)
  const axes3D = createContinuous3DAxes(styling, axisLabels?.x, axisLabels?.y, axisLabels?.z, scale)
  const grid3D = create3DGridConfig({
    styling,
    autoRotate: threeDRotate,
    orthographic: true,
    xCount,
    yCount,
    mode: 'continuous',
  })

  const lineStyle = seriesType === 'line3D' ? { width: 3, color: defaultColor } : undefined
  const isBar = seriesType === 'bar3D'
  const barSize = barSizeForContinuous3D(pointCount, grid3D.boxWidth, grid3D.boxDepth)
  const symbolSize = symbolSizeForContinuous3D(pointCount, grid3D.boxWidth, grid3D.boxDepth)
  const barProps = isBar
    ? { bevelSize: 0.3, bevelSmoothness: 3, shading: 'lambert' as const, barSize }
    : {}

  return {
    ...base,
    legend: { show: false },
    visualMap: resolve3DVisualMap(useVisualMap, seriesData, styling),
    tooltip: {
      ...base.tooltip,
      ...getTooltipTheme(isDark),
      formatter: createContinuous3DTooltipFormatter(isDark, {
        x: axisLabels?.x,
        y: axisLabels?.y,
        z: axisLabels?.z,
      }),
    },
    ...axes3D,
    grid3D,
    series: seriesData.map((s: Series3DData) => ({
      name: s.name,
      type: seriesType,
      data: s.data,
      symbolSize: isBar ? undefined : seriesType === 'line3D' ? 0 : symbolSize,
      ...(lineStyle ? { lineStyle } : {}),
      ...(useVisualMap ? {} : { itemStyle: { color: defaultColor } }),
      ...barProps,
      label: {
        show: showLabels,
        formatter: (p: { value: number[] }) => {
          const z = p.value[2]
          return z === undefined ? '' : String(round2(z))
        },
        textStyle: { fontSize: 12, color: styling.textColor },
      },
      emphasis: { label: { show: false } },
    })),
  } as unknown as EChartsOption
}

/**
 * Build the common grid3D block. The caller supplies the one or two differing
 * pieces (autoRotate + optional orthographic projection for line3D).
 */
export type Grid3DLayoutMode = 'grouped' | 'value' | 'continuous'

export function create3DGridConfig(opts: {
  styling: ChartStyling
  autoRotate: boolean
  orthographic?: boolean
  xCount: number
  yCount: number
  /** grouped = legacy sizing; value = fixed 100³ box; continuous = cubic auto-value grid. */
  mode?: Grid3DLayoutMode
}) {
  const { styling, autoRotate, orthographic, xCount, yCount } = opts
  const mode = opts.mode ?? 'grouped'
  const xWidth = mode === 'value' ? VALUE_MODE_3D_BOX_SIZE : boxSizeForAxisCount(xCount)
  const yWidth = mode === 'value' ? VALUE_MODE_3D_BOX_SIZE : boxSizeForAxisCount(yCount)

  const shell = {
    boxWidth: xWidth,
    boxDepth: yWidth,
    axisLine: { lineStyle: { color: styling.axisColor } },
    splitLine: { lineStyle: { color: styling.axisColor, opacity: styling.opacity } },
    light: {
      main: { intensity: 0.3, shadow: false },
      ambient: { intensity: 0.9 },
    },
  }

  if (mode === 'grouped') {
    return {
      ...shell,
      viewControl: {
        distance: xWidth + yWidth,
        autoRotate,
        ...(orthographic ? { projection: 'orthographic' as const } : {}),
      },
    }
  }

  if (mode === 'value') {
    const boxHeight = VALUE_MODE_3D_BOX_SIZE
    const orthographicSize = orthographicSizeFor3DBox(xWidth, yWidth, boxHeight)
    return {
      ...shell,
      boxHeight,
      viewControl: {
        distance: VALUE_MODE_3D_VIEW_DISTANCE,
        autoRotate,
        ...(orthographic
          ? {
              projection: 'orthographic' as const,
              orthographicSize,
              maxOrthographicSize: Math.max(400, orthographicSize * 2),
            }
          : {}),
      },
    }
  }

  const boxHeight = Math.max(xWidth, yWidth)
  const distance = viewDistanceFor3DBox(xWidth, yWidth, boxHeight)
  const orthographicSize = orthographicSizeFor3DBox(xWidth, yWidth, boxHeight)
  return {
    ...shell,
    boxHeight,
    viewControl: {
      distance,
      autoRotate,
      ...(orthographic
        ? {
            projection: 'orthographic' as const,
            orthographicSize,
            maxOrthographicSize: Math.max(400, orthographicSize * 2),
          }
        : { maxDistance: Math.max(400, distance * 1.5) }),
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
