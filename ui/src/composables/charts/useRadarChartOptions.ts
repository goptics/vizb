import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { getNextColorFor, hasXAxis, hasYAxis, hasZAxis } from '../../lib/utils'
import { getChartStyling, getTooltipTheme, formatRadarItemTooltip } from './shared'
import { fontSize, sortByTotal } from './shared/common'

const makeIndicators = (names: string[], perSpokeMax: number[]) =>
  names.map((name, i) => ({ name, max: Math.max((perSpokeMax[i] ?? 0) * 1.1, 1) }))

const makeTooltip = (
  isDark: boolean,
  indicatorNames: string[],
): EChartsOption['tooltip'] =>
  ({
    trigger: 'item',
    ...getTooltipTheme(isDark),
    formatter: (params: any) =>
      formatRadarItemTooltip(params, indicatorNames, isDark, getNextColorFor),
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
  const { chartData, sort, showLabels, isDark } = config

  const options = computed<EChartsOption>(() => {
    const cd = chartData.value
    const styling = getChartStyling(isDark.value)
    const baseOptions = getBaseOptions(config)
    const label = { show: showLabels.value, fontSize, color: styling.textColor }

    // X only: xAxis values as spokes, single polygon with totals
    if (!hasYAxis(chartData)) {
      const rows = cd.series.map((s) => ({ ...s, total: s.values[0] ?? 0 }))
      if (sort.value.enabled) rows.sort(sortByTotal(sort.value.order))
      const names = rows.map((s) => s.xAxis)
      const values = rows.map((s) => Math.max(0, s.total))
      return {
        ...baseOptions,
        tooltip: makeTooltip(isDark.value, names),
        legend: { show: false },
        radar: radarConfig(makeIndicators(names, values), styling),
        series: [
          {
            type: 'radar' as const,
            symbol: 'none',
            label,
            data: [{ value: values, name: cd.statType }],
          },
        ],
      }
    }

    const yAxis = cd.yAxis

    // Y only: yAxis values as spokes, single polygon with column totals
    if (!hasXAxis(chartData)) {
      const spokeTotals = yAxis.map((_, i) =>
        cd.series.reduce((sum, s) => sum + (s.values[i] ?? 0), 0),
      )
      return {
        ...baseOptions,
        tooltip: makeTooltip(isDark.value, yAxis),
        legend: { show: false },
        radar: radarConfig(makeIndicators(yAxis, spokeTotals), styling),
        series: [
          {
            type: 'radar' as const,
            symbol: 'none',
            label,
            data: [{ value: spokeTotals.map((v) => Math.max(0, v)), name: cd.statType }],
          },
        ],
      }
    }

    // X + Y + Z: Z = series in legend, X = multiple data points per Z series, Y = spokes
    if (hasZAxis(chartData)) {
      const yIndex = new Map(yAxis.map((y, i) => [y, i]))
      const grouped = new Map<string, Map<string, number[]>>()
      for (const pt of cd.points ?? []) {
        if (!grouped.has(pt.zAxis)) grouped.set(pt.zAxis, new Map())
        const zMap = grouped.get(pt.zAxis)!
        if (!zMap.has(pt.xAxis)) zMap.set(pt.xAxis, new Array(yAxis.length).fill(0))
        const idx = yIndex.get(pt.yAxis)
        if (idx !== undefined) zMap.get(pt.xAxis)![idx] = pt.value
      }

      const perSpokeMax = new Array<number>(yAxis.length).fill(0)
      for (const xMap of grouped.values()) {
        for (const vals of xMap.values()) {
          vals.forEach((v, i) => {
            if (v > (perSpokeMax[i] ?? 0)) perSpokeMax[i] = v
          })
        }
      }

      let zValues = cd.zAxis.filter((z) => z !== '')
      if (sort.value.enabled) {
        const zTotals = new Map<string, number>()
        for (const [z, xMap] of grouped) {
          let total = 0
          for (const vals of xMap.values()) total += vals.reduce((a, b) => a + b, 0)
          zTotals.set(z, total)
        }
        zValues = [...zValues].sort((a, b) =>
          sort.value.order === 'asc'
            ? (zTotals.get(a) ?? 0) - (zTotals.get(b) ?? 0)
            : (zTotals.get(b) ?? 0) - (zTotals.get(a) ?? 0),
        )
      }

      // Render largest Z series first so smaller ones stay on top and are hoverable.
      const zTotalsForRender = new Map<string, number>()
      for (const [z, xMap] of grouped) {
        let t = 0
        for (const vals of xMap.values()) t += vals.reduce((a, b) => a + b, 0)
        zTotalsForRender.set(z, t)
      }
      const renderZValues = [...zValues].sort(
        (a, b) => (zTotalsForRender.get(b) ?? 0) - (zTotalsForRender.get(a) ?? 0),
      )

      return {
        ...baseOptions,
        tooltip: makeTooltip(isDark.value, yAxis),
        legend: { data: zValues, bottom: 5, textStyle: { color: styling.textColor } },
        radar: radarConfig(makeIndicators(yAxis, perSpokeMax), styling),
        series: renderZValues.map((z) => ({
          name: z,
          type: 'radar' as const,
          symbol: 'circle',
          symbolSize: 10,
          label,
          itemStyle: { color: getNextColorFor(z) },
          lineStyle: { width: 1.5, opacity: 0.7 },
          areaStyle: { opacity: 0.1 },
          data: [...(grouped.get(z) ?? new Map<string, number[]>()).entries()].map(([x, vals]) => ({
            value: vals.map((v) => Math.max(0, v)),
            name: x,
          })),
        })),
      }
    }

    // X + Y: one polygon per X value, Y = spokes, legend = X values
    const rows = cd.series.map((s) => ({
      ...s,
      total: s.values.reduce((sum, v) => sum + v, 0),
    }))
    if (sort.value.enabled) rows.sort(sortByTotal(sort.value.order))

    const perSpokeMax = new Array<number>(yAxis.length).fill(0)
    for (const s of rows) {
      s.values.forEach((v, i) => {
        if (v > (perSpokeMax[i] ?? 0)) perSpokeMax[i] = v
      })
    }

    return {
      ...baseOptions,
      tooltip: makeTooltip(isDark.value, yAxis),
      legend: {
        data: rows.map((s) => s.xAxis),
        bottom: 5,
        textStyle: { color: styling.textColor },
      },
      radar: radarConfig(makeIndicators(yAxis, perSpokeMax), styling),
      series: rows.map((s) => ({
        name: s.xAxis,
        type: 'radar' as const,
        symbol: 'circle',
        symbolSize: 10,
        label,
        itemStyle: { color: getNextColorFor(s.xAxis) },
        lineStyle: { width: 1.5, opacity: 0.7 },
        areaStyle: { opacity: 0.1 },
        data: [{ value: s.values.map((v) => Math.max(0, v)), name: s.xAxis }],
      })),
    }
  })

  return { options }
}
