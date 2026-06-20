import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import { ref, type Ref } from 'vue'
import type { ChartData, Sort } from '@/types'
import { getBaseOptions, type BaseChartConfig } from './baseChartOptions'
import { useLineChartOptions } from './useLineChartOptions'
import { useBarChartOptions } from './useBarChartOptions'

// vitest runs in node — stub window.devicePixelRatio so getBaseOptions's
// is3D-pixelRatio branch has something to read.
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

// Minimal ChartData that satisfies the bits getBaseOptions / is3D touch.
// `yAxis` and `zAxis` are non-empty so is3D returns false (a 2D chart).
const makeChartData = (): ChartData => ({
  title: 't',
  statType: 'avg',
  yAxis: ['y'],
  zAxis: ['z'],
  series: [],
  points: [],
})

// Build a BaseChartConfig WITHOUT the (now optional) scale / threeDRotate fields.
// TypeScript will reject this if those fields are still marked required, so the
// test acts as a compile-time guard for the relaxation.
const makeMinimalConfig = (): BaseChartConfig => {
  const chartData: Ref<ChartData> = ref(makeChartData())
  const sort: Ref<Sort> = ref({ enabled: false, order: 'asc' })
  const showLabels = ref(false)
  const isDark = ref(false)
  return { chartData, sort, showLabels, isDark }
}

describe('BaseChartConfig (relaxed scale/threeDRotate)', () => {
  it('getBaseOptions works without scale/threeDRotate in the config', () => {
    const opts = getBaseOptions(makeMinimalConfig())
    expect(opts.tooltip).toBeDefined()
    expect(opts.toolbox).toBeDefined()
    expect(opts.legend).toBeDefined()
    expect(opts.emphasis).toEqual({ focus: 'series' })
  })

  it('getBaseOptions still works when scale/threeDRotate are provided', () => {
    const chartData: Ref<ChartData> = ref(makeChartData())
    const sort: Ref<Sort> = ref({ enabled: false, order: 'asc' })
    const showLabels = ref(false)
    const isDark = ref(false)
    const scale = ref<'linear' | 'log'>('log')
    const threeDRotate = ref(true)
    const cfg: BaseChartConfig = {
      chartData,
      sort,
      showLabels,
      isDark,
      scale,
      threeDRotate,
    }
    const opts = getBaseOptions(cfg)
    expect(opts.tooltip).toBeDefined()
    expect(opts.toolbox).toBeDefined()
    expect(opts.legend).toBeDefined()
  })
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

describe('useLineChartOptions — value mode', () => {
  it('emits xAxis.type=value when valueTuples present', () => {
    const { options } = useLineChartOptions(makeValueConfig())
    expect((options.value.xAxis as { type: string }).type).toBe('value')
  })

  it('emits yAxis.type=value when valueTuples present', () => {
    const { options } = useLineChartOptions(makeValueConfig())
    expect((options.value.yAxis as { type: string }).type).toBe('value')
  })

  it('emits line series with valueTuples as data', () => {
    const { options } = useLineChartOptions(makeValueConfig())
    const s = (options.value.series as { type: string; data: [number, number][] }[])[0]!
    expect(s.type).toBe('line')
    expect(s.data).toEqual([
      [100, 12],
      [200, 8],
    ])
  })

  it('includes dataZoom for both axes', () => {
    const { options } = useLineChartOptions(makeValueConfig())
    const dz = options.value.dataZoom as { xAxisIndex?: number; yAxisIndex?: number }[]
    expect(dz).toBeDefined()
    expect(dz.some((z) => z.xAxisIndex === 0)).toBe(true)
    expect(dz.some((z) => z.yAxisIndex === 0)).toBe(true)
  })
})

describe('useBarChartOptions — value mode', () => {
  it('emits xAxis.type=value when valueTuples present', () => {
    const { options } = useBarChartOptions(makeValueConfig())
    expect((options.value.xAxis as { type: string }).type).toBe('value')
  })

  it('emits bar series with valueTuples as data', () => {
    const { options } = useBarChartOptions(makeValueConfig())
    const s = (options.value.series as { type: string; data: [number, number][] }[])[0]!
    expect(s.type).toBe('bar')
    expect(s.data).toEqual([
      [100, 12],
      [200, 8],
    ])
  })
})
