import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import type { TitleOption } from 'echarts/types/dist/shared'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { getNextColorFor, hasXAxis, hasYAxis, hasZAxis } from '../../lib/utils'
import { getChartStyling, getTooltipTheme, formatTooltipValue } from './shared'
import { fontSize, sortByTotal, sortByValue } from './shared/common'
import type { Point3D } from '../../types'

type SeriesWithTotal = { xAxis: string; values: number[]; total: number }

type RadarDataItem = { name: string; value: number; itemStyle: { color: string | undefined } }

const computeYAxisTotals = (yAxis: string[], series: SeriesWithTotal[]): Map<string, number> => {
  const totals = new Map<string, number>()
  yAxis.forEach((y, i) => {
    totals.set(y, series.reduce((sum, s) => sum + (s.values[i] || 0), 0))
  })
  return totals
}

const computeZAxisTotals = (points: Point3D[]): Map<string, number> => {
  const totals = new Map<string, number>()
  for (const point of points) {
    totals.set(point.zAxis, (totals.get(point.zAxis) ?? 0) + point.value)
  }
  return totals
}

const makeRadarTitle = (text: string, left: string, styling: ReturnType<typeof getChartStyling>): TitleOption => ({
  text,
  left,
  top: '5%',
  textAlign: 'center',
  textStyle: { color: styling.textColor, fontSize, fontWeight: 'bold' },
})

const makeIndicators = (data: RadarDataItem[]) =>
  data.map((d) => ({ name: d.name, max: Math.max(d.value * 1.1, 1) }))

const makeRadarSeriesEntry = (
  name: string,
  data: RadarDataItem[],
  showLabels: boolean,
  styling: ReturnType<typeof getChartStyling>,
  radarIndex?: number,
): any => ({
  name,
  type: 'radar' as const,
  ...(radarIndex !== undefined ? { radarIndex } : {}),
  label: { show: showLabels, fontSize, color: styling.textColor },
  data: [
    {
      name,
      value: data.map((d) => d.value),
      itemStyle: { color: data[0]?.itemStyle?.color },
      lineStyle: { color: data[0]?.itemStyle?.color },
      areaStyle: { color: data[0]?.itemStyle?.color ?? '#5470C6', opacity: 0.15 },
    },
  ],
})

const makeRadarConfig = (
  data: RadarDataItem[],
  radius: string,
  center: [string, string],
  styling: ReturnType<typeof getChartStyling>,
): any => ({
  radius,
  center,
  indicator: makeIndicators(data),
  axisName: { color: styling.textColor },
  splitLine: { lineStyle: { color: styling.axisColor } },
  splitArea: { areaStyle: { opacity: 0.05 } },
})

// Tooltip that maps each polygon vertex back to its spoke name.
// ECharts radar params carry `value` (the full value array) and `seriesIndex`;
// `params.indicator` is not part of the callback — spoke names come from the
// data arrays that built the radar, captured via `dataArrays`.
const makeRadarTooltip = (
  isDark: boolean,
  dataArrays: RadarDataItem[][],
): EChartsOption['tooltip'] => ({
  trigger: 'item',
  ...getTooltipTheme(isDark),
  formatter: (params: any) => {
    if (!params) return ''
    const vals: number[] = Array.isArray(params.value) ? params.value : []
    const dataArr = dataArrays[params.seriesIndex ?? 0] ?? dataArrays[0] ?? []
    const rows = dataArr.map((d, i) => `${d.name}: <b>${formatTooltipValue(vals[i])}</b>`).join('<br/>')
    return `<strong>${params.name}</strong><br/>${rows}`
  },
} as EChartsOption['tooltip'])

export function useRadarChartOptions(config: BaseChartConfig) {
  const { chartData, sort, showLabels, isDark } = config

  const sortedData = computed(() => {
    const seriesWithTotals = chartData.value.series.map((series) => ({
      ...series,
      total: hasYAxis(chartData)
        ? series.values.reduce((sum, val) => sum + val, 0)
        : series.values[0] || 0,
    }))

    if (sort.value.enabled) {
      seriesWithTotals.sort(sortByTotal(sort.value.order))
    }

    return { series: seriesWithTotals }
  })

  const options = computed<EChartsOption>(() => {
    const sorted = sortedData.value
    const styling = getChartStyling(isDark.value)
    const baseOptions = getBaseOptions(config)

    const xAxisRadarData: RadarDataItem[] = sorted.series.map((s) => ({
      name: s.xAxis,
      value: Math.max(0, s.total),
      itemStyle: { color: getNextColorFor(s.xAxis) },
    }))

    // X-only (no Y): single radar with xAxis items as spokes
    if (!hasYAxis(chartData)) {
      return {
        ...baseOptions,
        tooltip: makeRadarTooltip(isDark.value, [xAxisRadarData]),
        legend: { show: false },
        radar: makeRadarConfig(xAxisRadarData, '70%', ['50%', '55%'], styling),
        series: [makeRadarSeriesEntry(chartData.value.statType, xAxisRadarData, showLabels.value, styling)],
      }
    }

    const yAxisTotals = computeYAxisTotals(chartData.value.yAxis, sorted.series)
    const yAxisRadarData: RadarDataItem[] = chartData.value.yAxis.map((y) => ({
      name: y,
      value: Math.max(0, yAxisTotals.get(y) ?? 0),
      itemStyle: { color: getNextColorFor(y) },
    }))

    if (sort.value.enabled) {
      yAxisRadarData.sort(sortByValue(sort.value.order))
    }

    // Y-only (no X): single radar with yAxis items as spokes
    if (!hasXAxis(chartData)) {
      return {
        ...baseOptions,
        tooltip: makeRadarTooltip(isDark.value, [yAxisRadarData]),
        legend: { show: false },
        radar: makeRadarConfig(yAxisRadarData, '70%', ['50%', '55%'], styling),
        series: [makeRadarSeriesEntry(chartData.value.statType, yAxisRadarData, showLabels.value, styling)],
      }
    }

    // X + Y + Z: three side-by-side radars
    if (hasZAxis(chartData)) {
      const zAxisTotals = computeZAxisTotals(chartData.value.points ?? [])
      const zAxisRadarData: RadarDataItem[] = chartData.value.zAxis
        .filter((z) => z !== '')
        .map((z) => ({
          name: z,
          value: Math.max(0, zAxisTotals.get(z) ?? 0),
          itemStyle: { color: getNextColorFor(z) },
        }))

      if (sort.value.enabled) {
        zAxisRadarData.sort(sortByValue(sort.value.order))
      }

      const labels = chartData.value.axisLabels
      return {
        ...baseOptions,
        tooltip: makeRadarTooltip(isDark.value, [xAxisRadarData, yAxisRadarData, zAxisRadarData]),
        legend: { show: false },
        title: [
          makeRadarTitle(labels?.x || 'X-Axis', '16.66%', styling),
          makeRadarTitle(labels?.y || 'Y-Axis', '50%', styling),
          makeRadarTitle(labels?.z || 'Z-Axis', '83.33%', styling),
        ],
        radar: [
          makeRadarConfig(xAxisRadarData, '25%', ['16.66%', '55%'], styling),
          makeRadarConfig(yAxisRadarData, '25%', ['50%', '55%'], styling),
          makeRadarConfig(zAxisRadarData, '25%', ['83.33%', '55%'], styling),
        ],
        series: [
          makeRadarSeriesEntry('By X-Axis', xAxisRadarData, showLabels.value, styling, 0),
          makeRadarSeriesEntry('By Y-Axis', yAxisRadarData, showLabels.value, styling, 1),
          makeRadarSeriesEntry('By Z-Axis', zAxisRadarData, showLabels.value, styling, 2),
        ],
      }
    }

    // X + Y: two side-by-side radars
    return {
      ...baseOptions,
      tooltip: makeRadarTooltip(isDark.value, [xAxisRadarData, yAxisRadarData]),
      legend: { show: false },
      title: [
        makeRadarTitle(chartData.value.axisLabels?.x || 'X-Axis', '25%', styling),
        makeRadarTitle(chartData.value.axisLabels?.y || 'Y-Axis', '75%', styling),
      ],
      radar: [
        makeRadarConfig(xAxisRadarData, '35%', ['25%', '55%'], styling),
        makeRadarConfig(yAxisRadarData, '35%', ['75%', '55%'], styling),
      ],
      series: [
        makeRadarSeriesEntry('By X-Axis', xAxisRadarData, showLabels.value, styling, 0),
        makeRadarSeriesEntry('By Y-Axis', yAxisRadarData, showLabels.value, styling, 1),
      ],
    }
  })

  return { options }
}
