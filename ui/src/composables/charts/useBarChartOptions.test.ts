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
