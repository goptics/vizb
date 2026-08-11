import { describe, expect, it, beforeAll, afterAll } from 'vitest'
import {
  baseConfig,
  emptyChartData,
  installDevicePixelRatio,
  makeSankeyChartData,
} from '@/test-utils'
import { useChordChartOptions } from './useChordChartOptions'

let restoreDpr: () => void
beforeAll(() => {
  restoreDpr = installDevicePixelRatio()
})
afterAll(() => restoreDpr())

const firstSeries = (options: { series?: unknown }) =>
  (options.series as Array<Record<string, any>>)[0]!

const chartData = (
  points = [
    { xAxis: 'A', yAxis: 'B', value: 3 },
    { xAxis: 'A', yAxis: 'B', value: 7 },
    { xAxis: 'B', yAxis: 'A', value: 5 },
  ]
) =>
  makeSankeyChartData({
    yAxis: ['A', 'B'],
    series: [
      { xAxis: 'A', values: [0], benchmarkId: '' },
      { xAxis: 'B', values: [0], benchmarkId: '' },
    ],
    points: points.map((point) => ({ ...point, zAxis: '' })),
  })

describe('useChordChartOptions', () => {
  it('emits a chord series with summed directional links and stable node colors', () => {
    const { options } = useChordChartOptions(baseConfig({ chartData: chartData() }))
    const series = firstSeries(options.value)
    expect(series.type).toBe('chord')
    expect(series.links).toEqual([
      { source: 'A', target: 'B', value: 10 },
      { source: 'B', target: 'A', value: 5 },
    ])
    expect(series.data).toEqual([
      { name: 'A', value: 15, itemStyle: { color: expect.any(String) } },
      { name: 'B', value: 15, itemStyle: { color: expect.any(String) } },
    ])
    expect(series.emphasis).toEqual({ focus: 'adjacency' })
    expect(series.lineStyle).toMatchObject({ color: 'gradient' })
  })

  it('supports labels, sorting, and non-negative display weights', () => {
    const { options } = useChordChartOptions(
      baseConfig({
        chartData: chartData([
          { xAxis: 'A', yAxis: 'B', value: -5 },
          { xAxis: 'C', yAxis: 'B', value: 3 },
        ]),
        showLabels: true,
        sort: { enabled: true, order: 'asc' },
      })
    )
    const series = firstSeries(options.value)
    expect(series.label?.show).toBe(true)
    expect(series.links).toEqual([
      { source: 'A', target: 'B', value: 0 },
      { source: 'C', target: 'B', value: 3 },
    ])
    expect(series.data.map((node: { name: string }) => node.name)).toEqual(['A', 'B', 'C'])
  })

  it('returns an empty chord series when an axis is missing', () => {
    const { options } = useChordChartOptions(
      baseConfig({
        chartData: emptyChartData({
          series: [{ xAxis: 'A', values: [1], benchmarkId: '' }],
          yAxis: [],
          points: [{ xAxis: 'A', yAxis: 'B', zAxis: '', value: 1 }],
        }),
      })
    )
    const series = firstSeries(options.value)
    expect(series.type).toBe('chord')
    expect(series.data).toEqual([])
    expect(series.links).toEqual([])
  })
})
