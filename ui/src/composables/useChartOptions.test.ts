import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import { ref } from 'vue'
import type { ChartData, ChartType, ScaleType, Axis } from '@/types'
import {
  baseConfig,
  makeGroupedChartData,
  makePieChartData,
  makeHeatmapChartData,
  makeRadarChartData,
  groupedRender3D,
  installDevicePixelRatio,
} from '@/test-utils'
import { useChartOptions } from './useChartOptions'

let restoreDpr: () => void
beforeAll(() => {
  restoreDpr = installDevicePixelRatio()
})
afterAll(() => restoreDpr())

function dispatch(
  chartType: ChartType,
  chartData: ChartData,
  opts: {
    threeD?: boolean
    arrangementTarget?: string
    chartAxes?: Axis[]
  } = {}
) {
  // baseConfig is the shared shape; useChartOptions takes loose refs in chart-card order.
  const cfg = baseConfig({
    chartData,
    threeD: opts.threeD ?? false,
    arrangementTarget: opts.arrangementTarget ?? 'xy',
    chartType,
  })
  return useChartOptions(
    cfg.chartData,
    cfg.sort,
    cfg.showLabels,
    cfg.isDark,
    ref(chartType),
    cfg.scale ?? ref<ScaleType>('linear'),
    cfg.stack ?? ref(false),
    cfg.threeDRotate ?? ref(false),
    cfg.visibleZ ?? ref({}),
    cfg.threeD ?? ref(false),
    cfg.threeDVisualMap ?? ref(false),
    cfg.visualMap ?? ref(false),
    cfg.arrangementTarget ?? ref('xy'),
    ref(opts.chartAxes),
    ref(undefined),
    ref(undefined),
    cfg.smooth ?? ref(false),
    cfg.horizontal ?? ref(false)
  )
}

function firstSeriesType(options: { series?: unknown }): string | undefined {
  const series = options.series as { type?: string }[] | undefined
  return series?.[0]?.type
}

const grouped3DData = (): ChartData =>
  makeGroupedChartData({
    zAxis: ['zA', 'zB'],
    points: [
      { xAxis: 'West', yAxis: 'Hardware', zAxis: 'zA', value: 5 },
      { xAxis: 'West', yAxis: 'Hardware', zAxis: 'zB', value: 7 },
    ],
    axisLabels: { x: 'region', y: 'category', z: 'group' },
    render3D: groupedRender3D,
  })

describe('useChartOptions dispatch', () => {
  it.each([
    {
      chartType: 'bar' as const,
      threeD: false,
      data: () => makeGroupedChartData(),
      expected: 'bar',
    },
    {
      chartType: 'line' as const,
      threeD: false,
      data: () => makeGroupedChartData(),
      expected: 'line',
    },
    {
      chartType: 'scatter' as const,
      threeD: false,
      data: () => makeGroupedChartData(),
      expected: 'scatter',
    },
    { chartType: 'pie' as const, threeD: false, data: () => makePieChartData(), expected: 'pie' },
    {
      chartType: 'heatmap' as const,
      threeD: false,
      data: () => makeHeatmapChartData(),
      expected: 'heatmap',
    },
    {
      chartType: 'radar' as const,
      threeD: false,
      data: () => makeRadarChartData(),
      expected: 'radar',
    },
    {
      chartType: 'sankey' as const,
      threeD: false,
      data: () =>
        makeGroupedChartData({
          points: [
            { xAxis: 'West', yAxis: 'Hardware', zAxis: '', value: 10 },
            { xAxis: 'East', yAxis: 'Software', zAxis: '', value: 40 },
          ],
        }),
      expected: 'sankey',
    },
    { chartType: 'bar' as const, threeD: true, data: grouped3DData, expected: 'bar3D' },
    { chartType: 'line' as const, threeD: true, data: grouped3DData, expected: 'line3D' },
    {
      chartType: 'scatter' as const,
      threeD: true,
      data: grouped3DData,
      expected: 'scatter3D',
    },
  ])(
    'chartType=$chartType threeD=$threeD → series type $expected',
    ({ chartType, threeD, data, expected }) => {
      const { options } = dispatch(chartType, data(), { threeD })
      expect(firstSeriesType(options.value)).toBe(expected)
    }
  )

  it('pie stays 2D pie even when chart data is 3D-shaped', () => {
    const { options } = dispatch('pie', grouped3DData(), { threeD: true })
    expect(firstSeriesType(options.value)).toBe('pie')
  })

  it('sankey stays 2D sankey even when chart data is 3D-shaped', () => {
    const { options } = dispatch('sankey', grouped3DData(), { threeD: true })
    expect(firstSeriesType(options.value)).toBe('sankey')
  })

  it('default branch falls back to bar options for unknown chart types', () => {
    const { options } = dispatch('unknown' as ChartType, makeGroupedChartData(), { threeD: false })
    expect(firstSeriesType(options.value)).toBe('bar')
  })

  it('default branch falls back to bar3D when use3D is true', () => {
    const { options } = dispatch('unknown' as ChartType, grouped3DData(), { threeD: true })
    expect(firstSeriesType(options.value)).toBe('bar3D')
  })
})
