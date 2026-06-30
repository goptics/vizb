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
})
