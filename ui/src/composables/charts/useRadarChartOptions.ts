import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { hasXAxis, hasYAxis, hasZAxis } from '../../lib/utils'
import { getChartStyling, getTooltipTheme, formatTooltipValue } from './shared'
import { fontSize, sortByTotal } from './shared/common'

interface AxisView {
  label: string
  names: string[]
  values: number[]
}

const makeIndicators = (names: string[], maxValues: number[]) =>
  names.map((name, i) => ({ name, max: Math.max((maxValues[i] ?? 0) * 1.1, 1) }))

const makeTooltip = (isDark: boolean, indicatorNames: string[]): EChartsOption['tooltip'] =>
  ({
    trigger: 'item',
    ...getTooltipTheme(isDark),
    formatter: (params: any) => {
      if (!params?.data) return ''
      const vals: number[] = Array.isArray(params.data.value) ? params.data.value : []
      const rows = indicatorNames
        .map((name, i) => `${name}: <b>${formatTooltipValue(vals[i])}</b>`)
        .join('<br/>')
      return `<strong>${params.data.name ?? params.name}</strong><br/>${rows}`
    },
  }) as EChartsOption['tooltip']

const radarConfig = (
  indicators: { name: string; max: number }[],
  styling: ReturnType<typeof getChartStyling>,
) => ({
  indicator: indicators,
  axisName: { color: styling.textColor },
  splitLine: { lineStyle: { color: styling.axisColor } },
  splitArea: { areaStyle: { opacity: 0.05 } },
})

export function useRadarChartOptions(config: BaseChartConfig) {
  const { chartData, sort, showLabels, isDark, visibleZ } = config

  const options = computed<EChartsOption>(() => {
    const cd = chartData.value
    const styling = getChartStyling(isDark.value)
    const baseOptions = getBaseOptions(config)
    const label = { show: showLabels.value, fontSize, color: styling.textColor }

    const xLabel = cd.axisLabels?.x || 'X-Axis'
    const yLabel = cd.axisLabels?.y || 'Y-Axis'
    const zLabel = cd.axisLabels?.z || 'Z-Axis'

    // Build one view per available axis
    const views: AxisView[] = []

    if (hasXAxis(chartData)) {
      const rows = cd.series.map((s) => ({ ...s, total: s.values.reduce((a, b) => a + b, 0) }))
      if (sort.value.enabled) rows.sort(sortByTotal(sort.value.order))
      views.push({
        label: xLabel,
        names: rows.map((s) => s.xAxis),
        values: rows.map((s) => Math.max(0, s.total)),
      })
    }

    if (hasYAxis(chartData)) {
      const yTotals = cd.yAxis.map((_, i) =>
        cd.series.reduce((sum, s) => sum + (s.values[i] ?? 0), 0),
      )
      views.push({
        label: yLabel,
        names: cd.yAxis,
        values: yTotals.map((v) => Math.max(0, v)),
      })
    }

    if (hasZAxis(chartData)) {
      const zTotals = new Map<string, number>()
      for (const pt of cd.points ?? []) {
        zTotals.set(pt.zAxis, (zTotals.get(pt.zAxis) ?? 0) + pt.value)
      }
      const zNames = cd.zAxis.filter((z) => z !== '')
      views.push({
        label: zLabel,
        names: zNames,
        values: zNames.map((z) => Math.max(0, zTotals.get(z) ?? 0)),
      })
    }

    if (views.length === 0) return baseOptions as EChartsOption

    // Single-axis: no legend switcher needed
    if (views.length === 1) {
      const view = views[0]!
      return {
        ...baseOptions,
        tooltip: makeTooltip(isDark.value, view.names),
        legend: { show: false },
        radar: radarConfig(makeIndicators(view.names, view.values), styling),
        series: [
          {
            type: 'radar' as const,
            symbol: 'none',
            label,
            data: [{ value: view.values, name: view.label }],
          },
        ],
      }
    }

    // Multi-axis: legend items = axis labels, one active at a time (like pie titles)
    const axisLabels = views.map((v) => v.label)
    const visibleZVal = visibleZ?.value ?? {}
    const activeLabel = axisLabels.find((l) => visibleZVal[l] === true) ?? axisLabels[0]!
    const activeView = views.find((v) => v.label === activeLabel) ?? views[0]!
    const legendSelected = Object.fromEntries(axisLabels.map((l) => [l, l === activeLabel]))

    return {
      ...baseOptions,
      tooltip: makeTooltip(isDark.value, activeView.names),
      legend: {
        data: axisLabels,
        selectedMode: 'single' as const,
        selected: legendSelected,
        bottom: 5,
        textStyle: { color: styling.textColor },
      },
      radar: radarConfig(makeIndicators(activeView.names, activeView.values), styling),
      series: [
        {
          type: 'radar' as const,
          symbol: 'none',
          label,
          data: [{ value: activeView.values, name: activeLabel }],
        },
      ],
    }
  })

  return { options }
}
