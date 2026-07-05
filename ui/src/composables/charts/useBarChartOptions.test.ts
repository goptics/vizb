import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import { ref } from 'vue'
import type { ChartData } from '@/types'
import type { BaseChartConfig } from './baseChartOptions'
import { useBarChartOptions } from './useBarChartOptions'

const originalDPR = (globalThis as { window?: { devicePixelRatio: number } }).window
  ?.devicePixelRatio
beforeAll(() => {
  ;(globalThis as unknown as { window: { devicePixelRatio: number } }).window = {
    devicePixelRatio: 1,
  }
})
afterAll(() => {
  if (originalDPR === undefined) {
    delete (globalThis as { window?: unknown }).window
  } else {
    ;(globalThis as unknown as { window: { devicePixelRatio: number } }).window = {
      devicePixelRatio: originalDPR,
    }
  }
})

const makeMixedChartData = (): ChartData => ({
  title: 'region vs tax',
  statType: 'mixed',
  yAxis: [],
  zAxis: [],
  series: [],
  points: [],
  axisLabels: { x: 'region', y: 'tax' },
  xCategories: ['West', 'South'],
  mixedTuples: [
    [0, 1926.35],
    [1, 447.38],
  ],
})

const makeMixedConfig = (): BaseChartConfig => ({
  chartData: ref(makeMixedChartData()),
  sort: ref({ enabled: false, order: 'asc' }),
  showLabels: ref(false),
  isDark: ref(false),
})

describe('useBarChartOptions — mixed mode', () => {
  it('emits bar series with mixedTuples as data', () => {
    const { options } = useBarChartOptions(makeMixedConfig())
    const s = (options.value.series as { type: string; data: [number, number][] }[])[0]!
    expect(s.type).toBe('bar')
    expect(s.data).toEqual([
      [0, 1926.35],
      [1, 447.38],
    ])
  })

  it('emits category xAxis', () => {
    const { options } = useBarChartOptions(makeMixedConfig())
    expect((options.value.xAxis as { type: string; data: string[] }).type).toBe('category')
    expect((options.value.xAxis as { data: string[] }).data).toEqual(['West', 'South'])
  })

  it('uses axis trigger with themed shadow pointer for category x', () => {
    const { options } = useBarChartOptions(makeMixedConfig())
    const tooltip = options.value.tooltip as {
      trigger?: string
      axisPointer?: {
        type?: string
        snap?: boolean
        shadowStyle?: { color?: string; opacity?: number }
      }
    }
    expect(tooltip.trigger).toBe('axis')
    expect(tooltip.axisPointer?.type).toBe('shadow')
    expect(tooltip.axisPointer?.snap).toBeUndefined()
    expect(tooltip.axisPointer?.shadowStyle?.color).toBe('#d1d5db')
    expect(tooltip.axisPointer?.shadowStyle?.opacity).toBe(0.4)
  })
})

const makeSimpleChartData = (): ChartData => ({
  title: 'items',
  statType: 'counts',
  yAxis: [],
  zAxis: [],
  series: [
    { xAxis: 'A', values: [10], benchmarkId: '' },
    { xAxis: 'B', values: [20], benchmarkId: '' },
    { xAxis: 'C', values: [30], benchmarkId: '' },
  ],
  points: [],
  axisLabels: { x: 'category', y: 'value' },
})

const makeSimpleConfig = (horizontal: boolean): BaseChartConfig => ({
  chartData: ref(makeSimpleChartData()),
  sort: ref({ enabled: false, order: 'asc' }),
  showLabels: ref(false),
  isDark: ref(false),
  horizontal: ref(horizontal),
})

const makeGroupedChartData = (): ChartData => ({
  title: 'regions',
  statType: 'counts',
  yAxis: ['North', 'South'],
  zAxis: [],
  series: [
    { xAxis: 'A', values: [10, 20], benchmarkId: '' },
    { xAxis: 'B', values: [15, 25], benchmarkId: '' },
  ],
  points: [],
  axisLabels: { x: 'category', y: 'region' },
})

const makeGroupedConfig = (horizontal: boolean): BaseChartConfig => ({
  chartData: ref(makeGroupedChartData()),
  sort: ref({ enabled: false, order: 'asc' }),
  showLabels: ref(false),
  isDark: ref(false),
  horizontal: ref(horizontal),
})

describe('useBarChartOptions — horizontal mode', () => {
  it('renders horizontal 1D bars with value xAxis and category yAxis', () => {
    const { options } = useBarChartOptions(makeSimpleConfig(true))
    const opt = options.value
    expect((opt.xAxis as { type: string }).type).toBe('value')
    expect((opt.yAxis as { type: string }).type).toBe('category')
    expect((opt.yAxis as { data: string[] }).data).toEqual(['A', 'B', 'C'])
    expect((opt.yAxis as { name?: string }).name).toBe('category')
    expect((opt.xAxis as { name?: string }).name).toBeUndefined()
  })

  it('renders horizontal grouped bars with correct series', () => {
    const { options } = useBarChartOptions(makeGroupedConfig(true))
    const opt = options.value
    expect((opt.xAxis as { type: string }).type).toBe('value')
    expect((opt.yAxis as { type: string }).type).toBe('category')
    expect((opt.yAxis as { name?: string }).name).toBe('category')
    expect((opt.xAxis as { name?: string }).name).toBeUndefined()
    const series = opt.series as { type: string; name: string; data: number[] }[]
    expect(series.length).toBe(2)
    expect(series[0]!.name).toBe('North')
    expect(series[1]!.name).toBe('South')
    expect(series[0]!.data).toEqual([10, 15])
  })

  it('places grouped horizontal legend at bottom center', () => {
    const { options } = useBarChartOptions(makeGroupedConfig(true))
    const legend = options.value.legend as { left?: string; bottom?: number; top?: number }
    expect(legend.left).toBe('center')
    expect(legend.bottom).toBe(0)
    expect(legend.top).toBeUndefined()
    expect(options.value.title).toBeUndefined()
  })

  it('renders vertical bars by default (horizontal not set)', () => {
    const { options } = useBarChartOptions(makeSimpleConfig(false))
    const opt = options.value
    expect((opt.xAxis as { type: string }).type).toBe('category')
    expect((opt.yAxis as { type: string }).type).toBe('value')
  })
})
