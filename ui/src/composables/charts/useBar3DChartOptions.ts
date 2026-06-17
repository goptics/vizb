import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { getNextColorFor } from '../../lib/utils'
import { getChartStyling, getTooltipTheme } from './shared'
import {
  EMPTY_RENDER,
  makeAxis3DCommon,
  axis3DName,
  create3DTooltipFormatter,
  createZLegendConfig,
  create3DGridConfig,
  create3DCellLabel,
} from './shared'
import type { Series3DData } from '../../types'

export function useBar3DChartOptions(config: BaseChartConfig) {
  const { chartData, isDark, autoRotate, visibleZ, showLabels, scale } = config

  const options = computed<EChartsOption>(() => {
    const styling = getChartStyling(isDark.value)
    const base = getBaseOptions(config)
    const points = chartData.value.points ?? []

    // Sorted axis categories + per-z series data are precomputed off-thread by
    // the transform worker (see lib/transform.ts) and carried on chartData.
    const render = chartData.value.render3D ?? EMPTY_RENDER
    const { xValues, yValues, zValues } = render
    const seriesData = render.barSeries

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

    const series = seriesData.map((s: Series3DData) => {
      const isTop = showLabels.value && s.name === lastVisibleZName
      return {
        name: s.name,
        type: 'bar3D' as const,
        stack: 'z',
        bevelSize: 0.3,
        bevelSmoothness: 3,
        data: s.data,
        itemStyle: { color: getNextColorFor(s.name) },
        shading: 'lambert',
        label: create3DCellLabel(isTop, cellTotals, styling.textColor),
        // echarts-gl ignores emphasis.disabled and shows the value label on hover by
        // default; kill it explicitly so the tooltip stays the only source of values.
        emphasis: { label: { show: false } },
      }
    })

    const axisCommon = makeAxis3DCommon(styling)

    const tooltipFormatter = create3DTooltipFormatter({
      xValues,
      yValues,
      zValues,
      aggPoints,
      isDark: isDark.value,
    })

    return {
      ...base,
      legend: {
        ...base.legend,
        ...createZLegendConfig(zValues, styling, sel),
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
        // Log scale mirrors the 2D value axis; the worker has already dropped
        // non-positive cells so nothing invalid lands on a log axis.
        ...(scale?.value === 'log' ? { type: 'log', logBase: 10 } : { type: 'value' }),
        ...axisCommon,
        // The vertical axis is the only free spatial axis for the z group, so its
        // label rides here inside the canvas — matching x/y axis names. The ticks
        // remain the metric value; the legend still lists the z categories.
        ...axis3DName(chartData.value.axisLabels?.z, styling),
      },
      grid3D: create3DGridConfig({
        styling,
        autoRotate: autoRotate?.value ?? false,
      }),
      series,
    } as unknown as EChartsOption
  })

  return { options }
}
