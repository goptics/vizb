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

export function useLine3DChartOptions(config: BaseChartConfig) {
  const { chartData, isDark, autoRotate, visibleZ, showLabels, scale } = config

  const options = computed<EChartsOption>(() => {
    const styling = getChartStyling(isDark.value)
    const base = getBaseOptions(config)
    const points = chartData.value.points ?? []

    // Sorted axis categories + per-z series data are precomputed off-thread by
    // the transform worker (see lib/transform.ts) and carried on chartData.
    const render = chartData.value.render3D ?? EMPTY_RENDER
    const { xValues, yValues, zValues } = render
    const seriesData = render.lineSeries

    // Legend can toggle z series off. Aggregates (tooltip sums) must reflect only
    // the currently-visible z, so sum over the filtered set. echarts treats a
    // missing legend key as selected → default everything on.
    const sel = visibleZ?.value ?? {}
    const aggPoints = points.filter((p) => sel[p.zAxis] !== false)

    // Precomputed per-cell totals from the transform worker (all z groups, unfiltered).
    const cellTotals = render.cellTotals ?? {}
    // For line3D the labels live on the scatter3D overlay (visible vertex dots).
    // Only the last visible z series shows the cell totals.
    const lastVisibleZName =
      [...zValues].reverse().find((z) => sel[z] !== false) ?? zValues[zValues.length - 1]

    const series = seriesData.map((s: Series3DData) => ({
      name: s.name,
      type: 'line3D' as const,
      lineStyle: { width: 3 },
      data: s.data,
      itemStyle: { color: getNextColorFor(s.name) },
      shading: 'lambert',
      // line3D labels go on the scatter3D overlay (visible vertex dots), not here.
      label: { show: false },
      // echarts-gl ignores emphasis.disabled and shows the value label on hover by
      // default; kill it explicitly so the tooltip stays the only source of values.
      emphasis: { label: { show: false } },
    }))

    // line3D can't draw its own vertex markers → overlay scatter3D dots so the data
    // points are visible. When showLabels is on, the last visible z series shows
    // cell totals via series-level formatter.
    const labelSeries = render.lineSeries.map((s: Series3DData) => ({
      name: s.name,
      type: 'scatter3D',
      data: s.data,
      symbolSize: 10,
      itemStyle: { color: getNextColorFor(s.name) },
      label: create3DCellLabel(
        showLabels.value && s.name === lastVisibleZName,
        cellTotals,
        styling.textColor
      ),
      emphasis: { label: { show: false } },
    }))

    const axisCommon = makeAxis3DCommon(styling)

    const tooltipFormatter = create3DTooltipFormatter({
      xValues,
      yValues,
      zValues,
      aggPoints,
      isDark: isDark.value,
      xAxisLabel: chartData.value.axisLabels?.x,
      yAxisLabel: chartData.value.axisLabels?.y,
      zAxisLabel: chartData.value.axisLabels?.z,
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
        orthographic: true,
      }),
      series: [...series, ...labelSeries],
    } as unknown as EChartsOption
  })

  return { options }
}
