import type { EChartsOption } from 'echarts'
import {
  getChartStyling,
  getTooltipTheme,
  createToolboxConfig,
  isLargeXAxis,
  createHeatmapDataZoomConfig,
} from './shared'
import { fontSize } from './shared/common'

// Build a correlation heatmap option from a symmetric K×K matrix. Rows/cols are
// the series labels; the cell value is r ∈ [-1, 1] coloured on a diverging
// red→neutral→green scale (matching the old table's hue intent). NaN cells (a
// constant series / too few pairs) are dropped so they render empty.
export function buildCorrelationOption(
  labels: string[],
  matrix: number[][],
  isDark: boolean,
  title = 'correlation'
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

  // Symmetric K×K matrix — both axes share the same size, so one threshold covers both.
  const large = isLargeXAxis(labels)

  return {
    backgroundColor: styling.backgroundColor,
    toolbox: createToolboxConfig(isDark, title, 2),
    // Large matrix: reserve fixed px for sliders + visualMap. Small: containLabel.
    grid: large
      ? { left: 60, right: '3%', bottom: 110, top: 8, containLabel: false }
      : { left: 8, right: 8, top: 8, bottom: 48, containLabel: true },
    ...(large
      ? { dataZoom: createHeatmapDataZoomConfig(true, true, labels.length, labels.length, styling) }
      : {}),
    tooltip: {
      position: 'top',
      ...getTooltipTheme(isDark),
      formatter: (p: unknown) => {
        const params = p as { data: [number, number, number] }
        const [j, i, v] = params.data
        return `<strong>${labels[i] ?? '—'}</strong> × <strong>${labels[j] ?? '—'}</strong><br/>r = ${v.toFixed(3)}`
      },
    },
    xAxis: {
      type: 'category',
      data: labels,
      splitArea: { show: true },
      // Series names are already on the y axis (the matrix is symmetric), so we
      // drop the x-axis labels to keep the heatmap uncluttered. The tooltip still
      // names both series for any cell.
      axisLabel: { show: false },
      axisTick: { show: false },
      axisLine: { lineStyle: { color: styling.axisColor } },
    },
    yAxis: {
      type: 'category',
      data: labels,
      // Put the first label at the top so the matrix reads like a table.
      inverse: true,
      splitArea: { show: true },
      axisLabel: { color: styling.textColor, fontSize, interval: large ? 'auto' : 0 },
      axisLine: { lineStyle: { color: styling.axisColor } },
    },
    visualMap: {
      min: -1,
      max: 1,
      calculable: true,
      orient: 'horizontal',
      left: 'center',
      bottom: 4,
      precision: 2,
      textStyle: { color: styling.textColor, fontSize },
      inRange: { color: ['#dc2626', neutral, '#16a34a'] },
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
