import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { getNextColorFor } from '../../lib/utils'
import { getChartStyling, getTooltipTheme, tooltipDivider, type ChartStyling } from './shared'
import { sortByAxisTotal } from './shared/common'
import type { Point3D } from '../../types'

type Series3DType = 'bar3D' | 'line3D'

const round2 = (v: number) => Math.round(v * 100) / 100

function mapPoints(
  points: Point3D[],
  z: string,
  xIndex: Map<string, number>,
  yIndex: Map<string, number>,
  fillGrid: boolean
) {
  // Aggregate per (x,y) cell: generic tabular data has many rows sharing the
  // same (x,y,z), and emitting one bar per row stacks coplanar boxes at the
  // same position → WebGL z-fighting (stippled patches). Sum into one bar/cell.
  const cells = new Map<string, number>()
  for (const p of points) {
    if (p.zAxis !== z) continue
    const xi = xIndex.get(p.xAxis)
    const yi = yIndex.get(p.yAxis)
    if (xi === undefined || yi === undefined) continue
    cells.set(`${xi},${yi}`, (cells.get(`${xi},${yi}`) ?? 0) + p.value)
  }
  // bar3D: emit a full (x,y) grid, filling absent cells with 0. Stacked bar3D
  // derives each segment's base from the previous series at the same cell; a
  // sparse series misaligns that base → floating segments. A complete grid keeps
  // stacks seated (0-height bars are invisible, so empty columns stay empty).
  // line3D: keep sparse — a 0-grid would drag every line down to the floor.
  if (!fillGrid) {
    return Array.from(cells, ([key, value]) => {
      const [xi, yi] = key.split(',').map(Number)
      return { value: [xi, yi, value] }
    })
  }
  const grid: { value: number[] }[] = []
  for (const xi of xIndex.values()) {
    for (const yi of yIndex.values()) {
      grid.push({ value: [xi, yi, cells.get(`${xi},${yi}`) ?? 0] })
    }
  }
  return grid
}

function makeAxisCommon(styling: ChartStyling) {
  return {
    axisLabel: { color: styling.textColor },
    axisLine: { lineStyle: { color: styling.axisColor } },
  }
}

// Axis name block for a 3D axis (bold + gap so it clears the tick labels and
// reads as a title, not another tick). Returns {} when no label is known.
function axis3DName(label: string | undefined, styling: ChartStyling) {
  if (!label) return {}
  return {
    name: label,
    nameGap: 25,
    nameTextStyle: { color: styling.textColor, fontSize: 14, fontWeight: 'bold' as const },
  }
}

export function use3DChartOptions(config: BaseChartConfig, seriesType: Series3DType) {
  const { chartData, sort, isDark, autoRotate, visibleZ } = config

  const options = computed<EChartsOption>(() => {
    const styling = getChartStyling(isDark.value)
    const base = getBaseOptions(config)
    const points: Point3D[] = chartData.value.points ?? []

    // Legend can toggle z series off. Aggregates (tooltip sums + bar-top labels)
    // must reflect only the currently-visible z, so sum over the filtered set.
    // echarts treats a missing legend key as selected → default everything on.
    const sel = visibleZ?.value ?? {}
    const aggPoints = points.filter((p) => sel[p.zAxis] !== false)

    let xValues = Array.from(new Set(points.map((p) => p.xAxis)))
    let yValues = Array.from(new Set(points.map((p) => p.yAxis)))
    let zValues = chartData.value.zAxis.filter((z) => z !== '')

    if (sort.value.enabled) {
      xValues = sortByAxisTotal(xValues, 'xAxis', points, sort.value.order)
      yValues = sortByAxisTotal(yValues, 'yAxis', points, sort.value.order)
      zValues = sortByAxisTotal(zValues, 'zAxis', points, sort.value.order)
    }

    const xIndex = new Map(xValues.map((v, i) => [v, i]))
    const yIndex = new Map(yValues.map((v, i) => [v, i]))

    const series = zValues.map((z) => ({
      name: z,
      type: seriesType,
      ...(seriesType === 'bar3D'
        ? { stack: 'z', bevelSize: 0.4, bevelSmoothness: 4 }
        : { lineStyle: { width: 2 } }),
      data: mapPoints(points, z, xIndex, yIndex, seriesType === 'bar3D'),
      itemStyle: { color: getNextColorFor(z) },
      shading: 'lambert',
      // No data labels: the z axis always exists here, so every value is already in
      // the tooltip — numbers on bars/points only clutter the 3D scene (and stacked
      // bar3D labels paint unreliably on sparse 4D anyway).
      label: { show: false },
      // echarts-gl ignores emphasis.disabled and shows the value label on hover by
      // default; kill it explicitly so the tooltip stays the only source of values.
      emphasis: { label: { show: false } },
    }))

    // line3D can't draw its own vertex markers → overlay scatter3D dots so the data
    // points are visible (labels off; values live in the tooltip). bar3D needs no
    // overlay.
    const labelSeries =
      seriesType === 'line3D'
        ? zValues.map((z) => ({
            name: z,
            type: 'scatter3D',
            data: mapPoints(points, z, xIndex, yIndex, false),
            symbolSize: 10,
            itemStyle: { color: getNextColorFor(z) },
            label: { show: false },
            emphasis: { label: { show: false } },
          }))
        : []

    const axisCommon = makeAxisCommon(styling)

    const tooltipFormatter = (params: { value: number[] }) => {
      const [xi = 0, yi = 0] = params.value
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
      for (const p of aggPoints) {
        const onX = p.xAxis === xName
        const onY = p.yAxis === yName
        if (onX) xMarginal += p.value
        if (onY) yMarginal += p.value
        if (onX && onY) {
          zmap.set(p.zAxis, (zmap.get(p.zAxis) ?? 0) + p.value)
          cellTotal += p.value
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
      const zSumLine = zmap.size > 1 ? `Σ z: <b>${round2(cellTotal)}</b><br/>` : ''

      // Marginal totals: sum over the other two axes for this x / this y.
      const margins =
        tooltipDivider(isDark.value) +
        zSumLine +
        `Σ ${xName}: <b>${round2(xMarginal)}</b><br/>` +
        `Σ ${yName}: <b>${round2(yMarginal)}</b>`

      return `<b>${xName} / ${yName}</b><br/>${rows.join('<br/>')}${margins}`
    }

    return {
      ...base,
      legend: {
        ...base.legend,
        show: zValues.length > 1,
        orient: 'vertical',
        left: 'left',
        top: 'middle',
        textStyle: { color: styling.textColor },
        // Controlled selection: persist toggles across recomputes (without this,
        // re-applying the option would reset every z back to visible).
        selected: sel,
      },
      tooltip: {
        ...base.tooltip,
        ...getTooltipTheme(isDark.value),
        formatter: tooltipFormatter,
      },
      xAxis3D: {
        type: 'category',
        data: xValues,
        ...axisCommon,
        ...axis3DName(chartData.value.axisLabels?.x, styling),
      },
      yAxis3D: {
        type: 'category',
        data: yValues,
        ...axisCommon,
        ...axis3DName(chartData.value.axisLabels?.y, styling),
      },
      zAxis3D: {
        type: 'value',
        ...axisCommon,
        // The vertical axis is the only free spatial axis for the z group, so its
        // label rides here inside the canvas — matching x/y axis names. The ticks
        // remain the metric value; the legend still lists the z categories.
        ...axis3DName(chartData.value.axisLabels?.z, styling),
      },
      grid3D: {
        boxWidth: 100,
        boxDepth: 100,
        axisLine: { lineStyle: { color: styling.axisColor } },
        splitLine: { lineStyle: { color: styling.axisColor, opacity: styling.opacity } },
        viewControl: {
          distance: 200,
          autoRotate: autoRotate.value,
          ...(seriesType === 'line3D' ? { projection: 'orthographic' } : {}),
        },
        light: {
          main: { intensity: 0.3, shadow: false },
          ambient: { intensity: 0.9 },
        },
      },
      series: [...series, ...labelSeries],
    } as unknown as EChartsOption
  })

  return { options }
}
