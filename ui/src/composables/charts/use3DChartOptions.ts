import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { getNextColorFor } from '../../lib/utils'
import { getChartStyling, getTooltipTheme, tooltipDivider, type ChartStyling } from './shared'
import type { Render3D } from '../../types'

type Series3DType = 'bar3D' | 'line3D'

const round2 = (v: number) => Math.round(v * 100) / 100

// Empty render payload when a chart lacks precomputed 3D data (shouldn't happen:
// the worker attaches render3D to every 3D chart).
const EMPTY_RENDER: Render3D = { xValues: [], yValues: [], zValues: [], barSeries: [], lineSeries: [], cellTotals: {} }

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
  const { chartData, isDark, autoRotate, visibleZ, showLabels } = config

  const options = computed<EChartsOption>(() => {
    const styling = getChartStyling(isDark.value)
    const base = getBaseOptions(config)
    const points = chartData.value.points ?? []

    // Sorted axis categories + per-z series data are precomputed off-thread by
    // the transform worker (see lib/transform.ts) and carried on chartData.
    const render = chartData.value.render3D ?? EMPTY_RENDER
    const { xValues, yValues, zValues } = render
    const seriesData = seriesType === 'bar3D' ? render.barSeries : render.lineSeries

    // Legend can toggle z series off. Aggregates (tooltip sums) must reflect only
    // the currently-visible z, so sum over the filtered set. echarts treats a
    // missing legend key as selected → default everything on.
    const sel = visibleZ?.value ?? {}
    const aggPoints = points.filter((p) => sel[p.zAxis] !== false)

    // Precomputed per-cell totals from the transform worker (all z groups, unfiltered).
    const cellTotals = render.cellTotals ?? {}
    // Only the visual top of the stacked bar gets labels to avoid duplicate text per cell.
    const lastVisibleZName =
      [...zValues].reverse().find((z) => sel[z] !== false) ?? zValues[zValues.length - 1]

    const makeLabel = (isTop: boolean) => {
      if (!isTop) return { show: false }
      return {
        show: true,
        formatter: (params: { value: number[] }) => {
          const [xi = 0, yi = 0] = params.value
          const total = cellTotals[`${xi},${yi}`] ?? 0
          return total > 0 ? String(round2(total)) : ''
        },
        textStyle: { fontSize: 11, color: styling.textColor },
      }
    }

    const series = seriesData.map((s) => {
      // line3D labels go on the scatter3D overlay (visible vertex dots), not here.
      const isTop = seriesType === 'bar3D' && showLabels.value && s.name === lastVisibleZName
      return {
        name: s.name,
        type: seriesType,
        ...(seriesType === 'bar3D'
          ? { stack: 'z', bevelSize: 0.4, bevelSmoothness: 4 }
          : { lineStyle: { width: 2 } }),
        data: s.data,
        itemStyle: { color: getNextColorFor(s.name) },
        shading: 'lambert',
        label: makeLabel(isTop),
        // echarts-gl ignores emphasis.disabled and shows the value label on hover by
        // default; kill it explicitly so the tooltip stays the only source of values.
        emphasis: { label: { show: false } },
      }
    })

    // line3D can't draw its own vertex markers → overlay scatter3D dots so the data
    // points are visible. When showLabels is on, the last visible z series shows
    // cell totals via series-level formatter. bar3D needs no overlay.
    const labelSeries =
      seriesType === 'line3D'
        ? render.lineSeries.map((s) => ({
            name: s.name,
            type: 'scatter3D',
            data: s.data,
            symbolSize: 10,
            itemStyle: { color: getNextColorFor(s.name) },
            label: makeLabel(showLabels.value && s.name === lastVisibleZName),
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
          // No camera tween: echarts-gl otherwise animates the camera in to
          // `distance` on (re)render, which reads as a jarring zoom flash when
          // switching chart type. Snap straight to the final position.
          animation: false,
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
