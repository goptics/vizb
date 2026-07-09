import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import { ref } from 'vue'
import type { ChartData } from '@/types'
import type { BaseChartConfig } from './baseChartOptions'
import { useLineChartOptions } from './useLineChartOptions'

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
  title: 'region vs latency',
  statType: 'mixed',
  yAxis: [],
  zAxis: [],
  series: [],
  points: [],
  axisLabels: { x: 'region', y: 'latency' },
  xCategories: ['Asia', 'EU'],
  mixedTuples: [
    [0, 12],
    [1, 11],
  ],
})

const makeMixedConfig = (): BaseChartConfig => ({
  chartData: ref(makeMixedChartData()),
  sort: ref({ enabled: false, order: 'asc' }),
  showLabels: ref(false),
  isDark: ref(false),
})

const makeSmoothMixedConfig = (): BaseChartConfig => ({
  ...makeMixedConfig(),
  smooth: ref(true),
})

const makeValueChartData = (): ChartData => ({
  title: 'price vs latency',
  statType: 'value',
  yAxis: [],
  zAxis: [],
  series: [],
  points: [],
  axisLabels: { x: 'price', y: 'latency' },
  valueTuples: [
    [100, 12],
    [200, 8],
  ],
})

const makeValueConfig = (): BaseChartConfig => ({
  chartData: ref(makeValueChartData()),
  sort: ref({ enabled: false, order: 'asc' }),
  showLabels: ref(false),
  isDark: ref(false),
})

const makeSmoothValueConfig = (): BaseChartConfig => ({
  ...makeValueConfig(),
  smooth: ref(true),
})

const makeGroupedChartData = (): ChartData => ({
  title: 'sales by month',
  statType: 'grouped',
  yAxis: ['North', 'South'],
  zAxis: [],
  series: [
    { xAxis: 'Jan', values: [10, 8], benchmarkId: 'jan' },
    { xAxis: 'Feb', values: [12, 9], benchmarkId: 'feb' },
  ],
  points: [],
  axisLabels: { x: 'month', y: 'region' },
})

const makeGroupedConfig = (opts: { smooth?: boolean; stack?: boolean } = {}): BaseChartConfig => ({
  chartData: ref(makeGroupedChartData()),
  sort: ref({ enabled: false, order: 'asc' }),
  showLabels: ref(false),
  isDark: ref(false),
  scale: ref<'linear' | 'log'>('linear'),
  smooth: ref(opts.smooth ?? false),
  stack: ref(opts.stack ?? false),
})

describe('useLineChartOptions — grouped mode', () => {
  it('emits stacked area line series when stack is enabled', () => {
    const { options } = useLineChartOptions(makeGroupedConfig({ stack: true }))
    const series = options.value.series as { stack?: string; areaStyle?: Record<string, never> }[]
    expect(series).toHaveLength(2)
    expect(series.every((s) => s.stack === 'total')).toBe(true)
    expect(series.every((s) => s.areaStyle !== undefined)).toBe(true)
  })

  it('does not stack grouped lines by default', () => {
    const { options } = useLineChartOptions(makeGroupedConfig())
    const series = options.value.series as { stack?: string; areaStyle?: unknown }[]
    expect(series.every((s) => s.stack === undefined)).toBe(true)
    expect(series.every((s) => s.areaStyle === undefined)).toBe(true)
  })
})

describe('useLineChartOptions — mixed mode', () => {
  it('emits line series with mixedTuples as data', () => {
    const { options } = useLineChartOptions(makeMixedConfig())
    const s = (options.value.series as { type: string; data: [number, number][] }[])[0]!
    expect(s.type).toBe('line')
    expect(s.data).toEqual([
      [0, 12],
      [1, 11],
    ])
  })

  it('emits smooth line series when enabled', () => {
    const { options } = useLineChartOptions(makeSmoothMixedConfig())
    const s = (options.value.series as { smooth?: boolean }[])[0]!
    expect(s.smooth).toBe(true)
  })

  it('uses axis trigger with themed snap line pointer for category x', () => {
    const { options } = useLineChartOptions(makeMixedConfig())
    const tooltip = options.value.tooltip as {
      trigger?: string
      axisPointer?: { type?: string; snap?: boolean; lineStyle?: { color?: string } }
    }
    expect(tooltip.trigger).toBe('axis')
    expect(tooltip.axisPointer?.type).toBe('line')
    expect(tooltip.axisPointer?.snap).toBe(true)
    expect(tooltip.axisPointer?.lineStyle?.color).toBe('#d1d5db')
  })
})

describe('useLineChartOptions — axes value mode', () => {
  it('uses cross axisPointer when both axes are value type', () => {
    const { options } = useLineChartOptions(makeValueConfig())
    expect((options.value.tooltip as { axisPointer?: { type?: string } }).axisPointer?.type).toBe(
      'cross'
    )
  })

  it('emits smooth line series when enabled', () => {
    const { options } = useLineChartOptions(makeSmoothValueConfig())
    const s = (options.value.series as { smooth?: boolean }[])[0]!
    expect(s.smooth).toBe(true)
  })
})

describe('useLineChartOptions — grouped mode', () => {
  it('preserves straight lines by default', () => {
    const { options } = useLineChartOptions(makeGroupedConfig())
    const series = options.value.series as { smooth?: boolean }[]
    expect(series.every((s) => s.smooth === undefined)).toBe(true)
  })

  it('emits smooth on every grouped line series when enabled', () => {
    const { options } = useLineChartOptions(makeGroupedConfig({ smooth: true }))
    const series = options.value.series as { smooth?: boolean }[]
    expect(series).toHaveLength(2)
    expect(series.every((s) => s.smooth === true)).toBe(true)
  })
})
