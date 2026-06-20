import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { COLOR_PALETTE, getNextColorFor } from '@/lib/utils'
import { getChartStyling, getTooltipTheme } from './shared'
import {
  EMPTY_RENDER,
  makeAxis3DCommon,
  axis3DName,
  create3DTooltipFormatter,
  createZLegendConfig,
  create3DGridConfig,
  create3DCellLabel,
  resolve3DVisualMap,
  createValue3DTooltipFormatter,
  createContinuous3DAxes,
  createContinuous3DTooltipFormatter,
  continuous3DGridCounts,
  round2,
} from './shared'
import type { Series3DData } from '@/types'

function valuePoints3DToSeries(points: [number, number, number][], title: string): Series3DData[] {
  return [
    {
      name: title,
      data: points.map(([x, y, z]) => ({ value: [x, y, z] })),
    },
  ]
}

export function useScatter3DChartOptions(config: BaseChartConfig) {
  const { chartData, isDark, threeDRotate, visibleZ, showLabels, scale, threeDVisualMap } = config

  const options = computed<EChartsOption>(() => {
    const styling = getChartStyling(isDark.value)
    const base = getBaseOptions(config)
    const render = chartData.value.render3D ?? EMPTY_RENDER
    const { xValues, yValues, zValues } = render
    const useVisualMap = threeDVisualMap?.value === true
    const defaultColor = COLOR_PALETTE[0]!
    const axisCommon = makeAxis3DCommon(styling)
    const zAxis3DBase = {
      ...(scale?.value === 'log'
        ? { type: 'log' as const, logBase: 10 }
        : { type: 'value' as const }),
      ...axisCommon,
    }
    const grid3D = create3DGridConfig({
      styling,
      autoRotate: threeDRotate?.value ?? false,
      orthographic: true,
      xCount: xValues.length,
      yCount: yValues.length,
    })

    const valuePoints3D = chartData.value.valuePoints3D
    if (!render.lineSeries.length && valuePoints3D?.length) {
      const seriesData = valuePoints3DToSeries(valuePoints3D, chartData.value.title)
      const pointCount = seriesData[0]?.data.length ?? 0
      const { xCount, yCount } = continuous3DGridCounts(pointCount)
      const axes3D = createContinuous3DAxes(
        styling,
        chartData.value.axisLabels?.x,
        chartData.value.axisLabels?.y,
        chartData.value.axisLabels?.z,
        scale?.value ?? 'linear'
      )
      const fallbackGrid3D = create3DGridConfig({
        styling,
        autoRotate: threeDRotate?.value ?? false,
        orthographic: true,
        xCount,
        yCount,
      })

      return {
        ...base,
        legend: { show: false },
        visualMap: resolve3DVisualMap(useVisualMap, seriesData, styling),
        tooltip: {
          ...base.tooltip,
          ...getTooltipTheme(isDark.value),
          formatter: createContinuous3DTooltipFormatter(isDark.value, {
            x: chartData.value.axisLabels?.x,
            y: chartData.value.axisLabels?.y,
            z: chartData.value.axisLabels?.z,
          }),
        },
        ...axes3D,
        grid3D: fallbackGrid3D,
        series: seriesData.map((s: Series3DData) => ({
          name: s.name,
          type: 'scatter3D' as const,
          data: s.data,
          symbolSize: 8,
          ...(useVisualMap ? {} : { itemStyle: { color: defaultColor } }),
          label: {
            show: showLabels.value,
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

    if (render.mode === 'continuous') {
      const seriesData = render.lineSeries
      const pointCount = seriesData[0]?.data.length ?? 0
      const { xCount, yCount } = continuous3DGridCounts(pointCount)
      const axes3D = createContinuous3DAxes(
        styling,
        chartData.value.axisLabels?.x,
        chartData.value.axisLabels?.y,
        chartData.value.axisLabels?.z,
        scale?.value ?? 'linear'
      )
      const continuousGrid3D = create3DGridConfig({
        styling,
        autoRotate: threeDRotate?.value ?? false,
        orthographic: true,
        xCount,
        yCount,
      })

      return {
        ...base,
        legend: { show: false },
        visualMap: resolve3DVisualMap(useVisualMap, seriesData, styling),
        tooltip: {
          ...base.tooltip,
          ...getTooltipTheme(isDark.value),
          formatter: createContinuous3DTooltipFormatter(isDark.value, {
            x: chartData.value.axisLabels?.x,
            y: chartData.value.axisLabels?.y,
            z: chartData.value.axisLabels?.z,
          }),
        },
        ...axes3D,
        grid3D: continuousGrid3D,
        series: seriesData.map((s: Series3DData) => ({
          name: s.name,
          type: 'scatter3D' as const,
          data: s.data,
          symbolSize: 8,
          ...(useVisualMap ? {} : { itemStyle: { color: defaultColor } }),
          label: {
            show: showLabels.value,
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

    if (render.mode === 'value') {
      const seriesData = render.lineSeries
      const valueLabel = chartData.value.statUnit
        ? `${chartData.value.title} (${chartData.value.statUnit})`
        : chartData.value.title
      const cellTotals = render.cellTotals ?? {}

      return {
        ...base,
        legend: { show: false },
        visualMap: resolve3DVisualMap(useVisualMap, seriesData, styling),
        tooltip: {
          ...base.tooltip,
          ...getTooltipTheme(isDark.value),
          formatter: createValue3DTooltipFormatter({
            xValues,
            yValues,
            seriesData: seriesData[0]?.data ?? [],
            isDark: isDark.value,
            xAxisLabel: chartData.value.axisLabels?.x,
            yAxisLabel: chartData.value.axisLabels?.y,
            valueLabel,
            seriesColor: defaultColor,
          }),
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
          ...zAxis3DBase,
          ...axis3DName(valueLabel, styling),
        },
        grid3D,
        series: seriesData.map((s: Series3DData) => ({
          name: s.name,
          type: 'scatter3D' as const,
          data: s.data,
          symbolSize: 10,
          ...(useVisualMap ? {} : { itemStyle: { color: defaultColor } }),
          label: create3DCellLabel(showLabels.value, cellTotals, styling.textColor),
          emphasis: { label: { show: false } },
        })),
      } as unknown as EChartsOption
    }

    const points = chartData.value.points ?? []
    const seriesData = render.lineSeries
    const sel = visibleZ?.value ?? {}
    const aggPoints = points.filter((p) => sel[p.zAxis] !== false)
    const cellTotals = render.cellTotals ?? {}
    const lastVisibleZName =
      [...zValues].reverse().find((z) => sel[z] !== false) ?? zValues[zValues.length - 1]

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
      visualMap: resolve3DVisualMap(useVisualMap, seriesData, styling),
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
        ...zAxis3DBase,
        ...axis3DName(chartData.value.axisLabels?.z, styling),
      },
      grid3D,
      series: seriesData.map((s: Series3DData) => {
        const color = getNextColorFor(s.name)
        return {
          name: s.name,
          type: 'scatter3D' as const,
          data: s.data,
          symbolSize: 10,
          itemStyle: { color },
          label: create3DCellLabel(
            showLabels.value && s.name === lastVisibleZName,
            cellTotals,
            styling.textColor
          ),
          emphasis: { label: { show: false } },
        }
      }),
    } as unknown as EChartsOption
  })

  return { options }
}
