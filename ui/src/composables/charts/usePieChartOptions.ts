import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import type { TitleOption } from 'echarts/types/dist/shared'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { getNextColorFor, hasXAxis, hasYAxis, hasZAxis } from '@/lib/utils'
import {
  getChartStyling,
  createPieSeriesConfig,
  getTooltipTheme,
  formatTooltipValue,
} from './shared'
import { fontSize, sortByTotal, sortByValue } from './shared/common'
import type { Point3D } from '@/types'

type SeriesWithTotal = { xAxis: string; values: number[]; total: number }

const computeYAxisTotals = (yAxis: string[], series: SeriesWithTotal[]): Map<string, number> => {
  const totals = new Map<string, number>()
  yAxis.forEach((y, i) => {
    totals.set(
      y,
      series.reduce((sum, s) => sum + (s.values[i] || 0), 0)
    )
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

const makePieTitle = (
  text: string,
  left: string,
  styling: ReturnType<typeof getChartStyling>
): TitleOption => ({
  text,
  left,
  top: '5%',
  textAlign: 'center',
  textStyle: { color: styling.textColor, fontSize, fontWeight: 'bold' },
})

export function usePieChartOptions(config: BaseChartConfig) {
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

    const formatter = (params: any) => {
      const percent = Number(params.percent).toFixed(2)
      return `${params.name} (${percent}%)`
    }

    // Pie hover tooltip: the generic config shows name + value only. Slices are
    // shares of a whole, so surface the percent here too (not just on labels).
    baseOptions.tooltip = {
      trigger: 'item',
      ...getTooltipTheme(isDark.value),
      formatter: (params: any) =>
        `${params.marker} <strong>${params.name}</strong><br/>${formatTooltipValue(params.value)} (${Number(params.percent).toFixed(2)}%)`,
    } as EChartsOption['tooltip']

    const xAxisPieData = sorted.series.map((s) => ({
      name: s.xAxis,
      value: Math.max(0, s.total),
      itemStyle: { color: getNextColorFor(s.xAxis) },
    }))

    const options: EChartsOption = {
      ...baseOptions,
      legend: { show: false },
      series: [
        createPieSeriesConfig(
          chartData.value.statType,
          xAxisPieData,
          showLabels.value,
          styling,
          formatter
        ),
      ],
    }

    if (!hasYAxis(chartData)) return options

    const yAxisTotals = computeYAxisTotals(chartData.value.yAxis, sorted.series)
    const yAxisPieData = chartData.value.yAxis.map((y) => ({
      name: y,
      value: Math.max(0, yAxisTotals.get(y) ?? 0),
      itemStyle: { color: getNextColorFor(y) },
    }))

    if (sort.value.enabled) {
      yAxisPieData.sort(sortByValue(sort.value.order))
    }

    if (!hasXAxis(chartData)) {
      options.series = [
        createPieSeriesConfig(
          chartData.value.statType,
          yAxisPieData,
          showLabels.value,
          styling,
          formatter
        ),
      ]
      return options
    }

    if (hasZAxis(chartData)) {
      const zAxisTotals = computeZAxisTotals(chartData.value.points ?? [])
      const zAxisPieData = chartData.value.zAxis
        .filter((z) => z !== '')
        .map((z) => ({
          name: z,
          value: Math.max(0, zAxisTotals.get(z) ?? 0),
          itemStyle: { color: getNextColorFor(z) },
        }))

      if (sort.value.enabled) {
        zAxisPieData.sort(sortByValue(sort.value.order))
      }

      const labels = chartData.value.axisLabels
      options.title = [
        makePieTitle(labels?.x || 'X-Axis', '16.66%', styling),
        makePieTitle(labels?.y || 'Y-Axis', '50%', styling),
        makePieTitle(labels?.z || 'Z-Axis', '83.33%', styling),
      ]

      const specs3D = [
        {
          name: 'By X-Axis',
          data: xAxisPieData,
          radius: ['25%', '50%'] as [string, string],
          center: ['16.66%', '50%'] as [string, string],
        },
        {
          name: 'By Y-Axis',
          data: yAxisPieData,
          radius: ['25%', '50%'] as [string, string],
          center: ['50%', '50%'] as [string, string],
        },
        {
          name: 'By Z-Axis',
          data: zAxisPieData,
          radius: ['25%', '50%'] as [string, string],
          center: ['83.33%', '50%'] as [string, string],
        },
      ]
      options.series = specs3D.map(({ name, data, radius, center }) =>
        createPieSeriesConfig(name, data, showLabels.value, styling, formatter, radius, center)
      )

      return options
    }

    options.title = [
      makePieTitle(chartData.value.axisLabels?.x || 'X-Axis', '25%', styling),
      makePieTitle(chartData.value.axisLabels?.y || 'Y-Axis', '75%', styling),
    ]

    const specs2D = [
      {
        name: 'By X-Axis',
        data: xAxisPieData,
        radius: ['30%', '60%'] as [string, string],
        center: ['25%', '50%'] as [string, string],
      },
      {
        name: 'By Y-Axis',
        data: yAxisPieData,
        radius: ['30%', '60%'] as [string, string],
        center: ['75%', '50%'] as [string, string],
      },
    ]
    options.series = specs2D.map(({ name, data, radius, center }) =>
      createPieSeriesConfig(name, data, showLabels.value, styling, formatter, radius, center)
    )

    return options
  })

  return { options }
}
