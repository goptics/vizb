import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import { ref } from 'vue'
import type { ChartData } from '@/types'
import type { BaseChartConfig } from './baseChartOptions'
import { useScatterChartOptions } from './useScatterChartOptions'

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
  visualMap: ref(false),
})

const makeGroupedChartData = (): ChartData => ({
  title: 'avg',
  statType: 'avg',
  yAxis: ['A', 'B'],
  zAxis: [],
  series: [
    { xAxis: 'x1', values: [10, 20], benchmarkId: '' },
    { xAxis: 'x2', values: [15, 25], benchmarkId: '' },
  ],
  points: [],
  axisLabels: { x: 'category', y: 'group' },
})

const makeGroupedConfig = (): BaseChartConfig => ({
  chartData: ref(makeGroupedChartData()),
  sort: ref({ enabled: false, order: 'asc' }),
  showLabels: ref(false),
  isDark: ref(false),
  visualMap: ref(false),
})

describe('useScatterChartOptions — axes value mode', () => {
  it('omits axis names when axisLabels are absent', () => {
    const cfg = makeValueConfig()
    cfg.chartData.value = { ...makeValueChartData(), axisLabels: {} }
    const { options } = useScatterChartOptions(cfg)
    const xAxis = options.value.xAxis as { name?: string }
    const yAxis = options.value.yAxis as { name?: string }
    expect(xAxis.name).toBeUndefined()
    expect(yAxis.name).toBeUndefined()
  })

  it('emits xAxis.type=value when valueTuples present', () => {
    const { options } = useScatterChartOptions(makeValueConfig())
    expect((options.value.xAxis as { type: string }).type).toBe('value')
  })

  it('emits yAxis.type=value when valueTuples present', () => {
    const { options } = useScatterChartOptions(makeValueConfig())
    expect((options.value.yAxis as { type: string }).type).toBe('value')
  })

  it('emits scatter series with valueTuples as data', () => {
    const { options } = useScatterChartOptions(makeValueConfig())
    const s = (options.value.series as { type: string; data: [number, number][] }[])[0]!
    expect(s.type).toBe('scatter')
    expect(s.data).toEqual([
      [100, 12],
      [200, 8],
    ])
  })

  it('uses a taller plot grid (minimal top) without a legend band', () => {
    const { options } = useScatterChartOptions(makeValueConfig())
    expect((options.value.grid as { top?: number }).top).toBe(8)
  })

  it('includes dataZoom for both axes', () => {
    const { options } = useScatterChartOptions(makeValueConfig())
    const dz = options.value.dataZoom as { xAxisIndex?: number; yAxisIndex?: number }[]
    expect(dz).toBeDefined()
    expect(dz.some((z) => z.xAxisIndex === 0)).toBe(true)
    expect(dz.some((z) => z.yAxisIndex === 0)).toBe(true)
  })

  it('inherits grouping axis title styling (bold 16px names)', () => {
    const { options } = useScatterChartOptions(makeValueConfig())
    const xAxis = options.value.xAxis as {
      nameTextStyle?: { fontWeight?: string; fontSize?: number }
    }
    const yAxis = options.value.yAxis as {
      nameTextStyle?: { fontWeight?: string; fontSize?: number }
    }
    expect(xAxis.nameTextStyle?.fontWeight).toBe('bold')
    expect(xAxis.nameTextStyle?.fontSize).toBe(16)
    expect(yAxis.nameTextStyle?.fontWeight).toBe('bold')
    expect(yAxis.nameTextStyle?.fontSize).toBe(16)
  })

  it('inherits toolbox and emphasis from getBaseOptions', () => {
    const { options } = useScatterChartOptions(makeValueConfig())
    expect(options.value.toolbox).toBeDefined()
    expect(options.value.emphasis).toEqual({ focus: 'series' })
  })

  it('respects showLabels like grouping charts', () => {
    const cfg = makeValueConfig()
    cfg.showLabels.value = true
    const { options } = useScatterChartOptions(cfg)
    const s = (options.value.series as { label?: { show?: boolean } }[])[0]!
    expect(s.label?.show).toBe(true)
  })

  it('emits visualMap when enabled in value mode', () => {
    const cfg = makeValueConfig()
    cfg.visualMap!.value = true
    const { options } = useScatterChartOptions(cfg)
    expect(options.value.visualMap).toMatchObject({ show: true, dimension: 1 })
    const s = (options.value.series as { itemStyle?: { color?: string } }[])[0]!
    expect(s.itemStyle?.color).toBeUndefined()
  })

  it('omits visualMap when disabled', () => {
    const { options } = useScatterChartOptions(makeValueConfig())
    expect(options.value.visualMap).toEqual([])
  })
})

describe('useScatterChartOptions — group mode', () => {
  it('omits x-axis name when axisLabels.x is absent', () => {
    const cfg = makeGroupedConfig()
    cfg.chartData.value = { ...makeGroupedChartData(), axisLabels: { y: 'group' } }
    const { options } = useScatterChartOptions(cfg)
    const xAxis = options.value.xAxis as { name?: string }
    expect(xAxis.name).toBeUndefined()
  })

  it('emits category axes for grouped data', () => {
    const { options } = useScatterChartOptions(makeGroupedConfig())
    expect((options.value.xAxis as { type: string }).type).toBe('category')
    expect((options.value.yAxis as { type: string }).type).toBe('value')
  })

  it('transposes y groups into scatter series', () => {
    const { options } = useScatterChartOptions(makeGroupedConfig())
    const series = options.value.series as { type: string; name: string; data: number[] }[]
    expect(series).toHaveLength(2)
    expect(series[0]!.type).toBe('scatter')
    expect(series[0]!.name).toBe('A')
    expect(series[0]!.data).toEqual([10, 15])
    expect(series[1]!.name).toBe('B')
    expect(series[1]!.data).toEqual([20, 25])
  })

  it('shows legend for multi-series grouped scatter', () => {
    const { options } = useScatterChartOptions(makeGroupedConfig())
    expect((options.value.legend as { show?: boolean }).show).toBe(true)
  })

  it('emits visualMap for grouped scatter when enabled', () => {
    const cfg = makeGroupedConfig()
    cfg.visualMap!.value = true
    const { options } = useScatterChartOptions(cfg)
    expect(options.value.visualMap).toMatchObject({ show: true })
  })
})
