import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import {
  baseConfig,
  makeGroupedChartData,
  makeMixedChartData,
  makeValueChartData,
  emptyChartData,
  installDevicePixelRatio,
} from '@/test-utils'
import { LARGE_X_THRESHOLD } from './shared/chartConfig'
import { useCategorySeriesChartOptions } from './useCategorySeriesChartOptions'

let restoreDpr: () => void
beforeAll(() => {
  restoreDpr = installDevicePixelRatio()
})
afterAll(() => restoreDpr())

const xOnly = () =>
  emptyChartData({
    title: 'items',
    statType: 'counts',
    yAxis: [],
    series: [
      { xAxis: 'A', values: [10], benchmarkId: '' },
      { xAxis: 'B', values: [20], benchmarkId: '' },
      { xAxis: 'C', values: [null], benchmarkId: '' },
    ],
    axisLabels: { x: 'cat' },
  })

describe('useCategorySeriesChartOptions — line', () => {
  it('emits grouped line series with y groups', () => {
    const { options } = useCategorySeriesChartOptions(
      baseConfig({ chartData: makeGroupedChartData() }),
      'line'
    )
    const series = options.value.series as { type: string; name: string }[]
    expect(series.every((s) => s.type === 'line')).toBe(true)
    expect(series.map((s) => s.name)).toEqual(['Hardware', 'Software'])
    expect(options.value.title).toBeDefined()
  })

  it('stacks lines with area when stack is on', () => {
    const { options } = useCategorySeriesChartOptions(
      baseConfig({ chartData: makeGroupedChartData(), stack: true }),
      'line'
    )
    const series = options.value.series as { stack?: string; areaStyle?: object }[]
    expect(series.every((s) => s.stack === 'total')).toBe(true)
    expect(series.every((s) => s.areaStyle)).toBeTruthy()
  })

  it('enables smooth lines when smooth is on', () => {
    const { options } = useCategorySeriesChartOptions(
      baseConfig({ chartData: makeGroupedChartData(), smooth: true }),
      'line'
    )
    const series = options.value.series as { smooth?: boolean }[]
    expect(series.every((s) => s.smooth === true)).toBe(true)
  })

  it('builds single-series x-only line with pinned tooltip', () => {
    const { options } = useCategorySeriesChartOptions(
      baseConfig({ chartData: xOnly(), showLabels: true }),
      'line'
    )
    const series = options.value.series as {
      type: string
      data: (number | null)[]
      connectNulls?: boolean
    }[]
    expect(series).toHaveLength(1)
    expect(series[0]!.type).toBe('line')
    expect(series[0]!.data).toEqual([10, 20, null])
    expect(series[0]!.connectNulls).toBe(true)
    expect((options.value.legend as { show?: boolean }).show).toBe(false)
    expect((options.value.tooltip as { trigger?: string }).trigger).toBe('axis')
  })

  it('uses large-series symbol none and dataZoom for wide x axes', () => {
    const many = Array.from({ length: LARGE_X_THRESHOLD + 1 }, (_, i) => `x${i}`)
    const chartData = emptyChartData({
      title: 'wide',
      series: many.map((x, i) => ({ xAxis: x, values: [i + 1], benchmarkId: '' })),
      axisLabels: { x: 'x' },
    })
    const { options } = useCategorySeriesChartOptions(baseConfig({ chartData }), 'line')
    const series = options.value.series as { symbol?: string; sampling?: string }[]
    expect(series[0]!.symbol).toBe('none')
    expect(series[0]!.sampling).toBe('lttb')
    expect(options.value.dataZoom).toBeDefined()
  })

  it('omits legend title when y axis label missing', () => {
    const chartData = makeGroupedChartData({ axisLabels: { x: 'region' } })
    const { options } = useCategorySeriesChartOptions(baseConfig({ chartData }), 'line')
    expect(options.value.title).toBeUndefined()
  })
})

describe('useCategorySeriesChartOptions — scatter', () => {
  it('emits scatter series for grouped data', () => {
    const { options } = useCategorySeriesChartOptions(
      baseConfig({ chartData: makeGroupedChartData() }),
      'scatter'
    )
    const series = options.value.series as { type: string }[]
    expect(series.every((s) => s.type === 'scatter')).toBe(true)
  })

  it('applies visualMap for x-only scatter when enabled', () => {
    const { options } = useCategorySeriesChartOptions(
      baseConfig({ chartData: xOnly(), visualMap: true }),
      'scatter'
    )
    expect(options.value.visualMap).toMatchObject({ show: true, dimension: 1 })
    const series = options.value.series as { itemStyle?: unknown; large?: boolean }[]
    expect(series[0]!.itemStyle).toBeUndefined()
    expect(series[0]!.large).toBe(false)
  })

  it('applies visualMap for grouped scatter and collects finite values', () => {
    const chartData = makeGroupedChartData({
      series: [
        { xAxis: 'West', values: [10, null], benchmarkId: '' },
        { xAxis: 'East', values: [Number.NaN, 40], benchmarkId: '' },
      ],
    })
    const { options } = useCategorySeriesChartOptions(
      baseConfig({ chartData, visualMap: true }),
      'scatter'
    )
    expect(options.value.visualMap).toMatchObject({ show: true, max: 40 })
  })

  it('does not stack scatter even when stack flag is set', () => {
    const { options } = useCategorySeriesChartOptions(
      baseConfig({ chartData: makeGroupedChartData(), stack: true }),
      'scatter'
    )
    const series = options.value.series as { stack?: string }[]
    expect(series.every((s) => s.stack === undefined)).toBe(true)
  })
})

describe('useCategorySeriesChartOptions — mixed / value dispatch', () => {
  it('routes mixedTuples to mixed-axis 2D options', () => {
    const { options } = useCategorySeriesChartOptions(
      baseConfig({ chartData: makeMixedChartData() }),
      'line'
    )
    const xAxis = options.value.xAxis as { type: string; data: string[] }
    const series = options.value.series as { type: string; data: [number, number][] }[]
    expect(xAxis.type).toBe('category')
    expect(xAxis.data).toEqual(['West', 'South'])
    expect((options.value.yAxis as { type: string }).type).toBe('value')
    expect(series[0]!.type).toBe('line')
    expect(series[0]!.data).toEqual([
      [0, 1926.35],
      [1, 447.38],
    ])
  })

  it('routes valueTuples to value-axis 2D options', () => {
    const { options } = useCategorySeriesChartOptions(
      baseConfig({ chartData: makeValueChartData() }),
      'scatter'
    )
    expect((options.value.xAxis as { type: string }).type).toBe('value')
    expect((options.value.yAxis as { type: string }).type).toBe('value')
    const series = options.value.series as { type: string; data: [number, number][] }[]
    expect(series[0]!.type).toBe('scatter')
    expect(series[0]!.data).toEqual([
      [100, 12],
      [200, 8],
    ])
  })

  it('prefers mixedTuples when both tuple modes are present', () => {
    const chartData = makeMixedChartData({
      valueTuples: [
        [1, 2],
        [3, 4],
      ],
    })
    const { options } = useCategorySeriesChartOptions(baseConfig({ chartData }), 'line')
    expect((options.value.xAxis as { type: string }).type).toBe('category')
    const series = options.value.series as { data: [number, number][] }[]
    expect(series[0]!.data).toEqual([
      [0, 1926.35],
      [1, 447.38],
    ])
  })

  it('falls through to grouped series when tuple arrays are empty', () => {
    const chartData = makeGroupedChartData({ mixedTuples: [], valueTuples: [] })
    const { options } = useCategorySeriesChartOptions(baseConfig({ chartData }), 'line')
    const series = options.value.series as { type: string; name: string }[]
    expect(series.every((s) => s.type === 'line')).toBe(true)
    expect(series.map((s) => s.name)).toEqual(['Hardware', 'Software'])
  })
})

describe('useCategorySeriesChartOptions — remaining branches', () => {
  it('smooth x-only lines and large grouped dataZoom', () => {
    const { options } = useCategorySeriesChartOptions(
      baseConfig({ chartData: xOnly(), smooth: true }),
      'line'
    )
    expect((options.value.series as { smooth?: boolean }[])[0]!.smooth).toBe(true)

    const many = Array.from({ length: LARGE_X_THRESHOLD + 1 }, (_, i) => `x${i}`)
    const chartData = makeGroupedChartData({
      series: many.map((x) => ({ xAxis: x, values: [1, 2], benchmarkId: '' })),
    })
    const { options: wide } = useCategorySeriesChartOptions(baseConfig({ chartData }), 'line')
    expect(wide.value.dataZoom).toBeDefined()
  })
})
