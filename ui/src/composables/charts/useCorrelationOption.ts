import type { EChartsOption } from 'echarts'
import type { CorrelationMethod } from '../../lib/stats'
import {
  getChartStyling,
  getTooltipTheme,
  createToolboxConfig,
  isLargeXAxis,
  createHeatmapDataZoomConfig,
  createHeatmapLayoutConfig,
  hasRotatedXLabels,
} from './shared'
import { fontSize } from './shared/common'

const PREFIX: Record<CorrelationMethod, string> = {
  pearson: 'r',
  spearman: 'ρ',
  kendall: 'τ',
  dcor: 'dCor',
}

export function buildCorrelationOption(
  labels: string[],
  matrix: number[][],
  isDark: boolean,
  title = 'correlation',
  method: CorrelationMethod = 'pearson'
): EChartsOption {
  const styling = getChartStyling(isDark)
  const neutral = isDark ? '#374151' : '#f3f4f6'

  const data: [number, number, number][] = []
  for (let i = 0; i < matrix.length; i++) {
    const row = matrix[i]!
    for (let j = 0; j < row.length; j++) {
      const v = row[j]!
      if (Number.isFinite(v)) data.push([j, i, v])
    }
  }

  const large = isLargeXAxis(labels)
  const layout = createHeatmapLayoutConfig({
    hasXDataZoom: large,
    hasYDataZoom: large,
    compact: true,
  })

  const isDiverging = method !== 'dcor'
  const vmMin = isDiverging ? -1 : 0
  const vmColors = isDiverging
    ? ['#dc2626', neutral, '#16a34a']
    : [isDark ? '#1e293b' : '#f9fafb', '#1d4ed8']

  const prefix = PREFIX[method]

  return {
    backgroundColor: styling.backgroundColor,
    toolbox: createToolboxConfig(isDark, title, 2),
    grid: layout.grid,
    ...(large
      ? { dataZoom: createHeatmapDataZoomConfig(true, true, labels.length, labels.length, styling) }
      : {}),
    tooltip: {
      position: 'top',
      ...getTooltipTheme(isDark),
      formatter: (p: unknown) => {
        const params = p as { data: [number, number, number] }
        const [j, i, v] = params.data
        return `<strong>${labels[i] ?? '—'}</strong> × <strong>${labels[j] ?? '—'}</strong><br/>${prefix} = ${v.toFixed(3)}`
      },
    },
    xAxis: {
      type: 'category',
      data: labels,
      splitArea: { show: true },
      axisLabel: {
        color: styling.textColor,
        fontSize,
        interval: large ? 'auto' : 0,
        rotate: hasRotatedXLabels(labels, large) ? 30 : 0,
      },
      axisTick: { show: false },
      axisLine: { lineStyle: { color: styling.axisColor } },
    },
    yAxis: {
      type: 'category',
      data: labels,
      inverse: true,
      splitArea: { show: true },
      axisLabel: { color: styling.textColor, fontSize, interval: large ? 'auto' : 0 },
      axisLine: { lineStyle: { color: styling.axisColor } },
    },
    visualMap: {
      min: vmMin,
      max: 1,
      calculable: true,
      orient: 'horizontal',
      left: 'center',
      bottom: layout.visualMapBottom,
      precision: 2,
      textStyle: { color: styling.textColor, fontSize },
      inRange: { color: vmColors },
    },
    series: [
      {
        type: 'heatmap',
        data,
        label: {
          show: true,
          color: styling.textColor,
          fontSize,
          formatter: (p: unknown) => {
            const v = (p as { data: [number, number, number] }).data[2]
            return Number(v.toFixed(2)).toString()
          },
        },
        emphasis: { itemStyle: { shadowBlur: 6, shadowColor: 'rgba(0,0,0,0.35)' } },
      },
    ],
  }
}
