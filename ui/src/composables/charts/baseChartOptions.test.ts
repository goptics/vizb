import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import { ref, type Ref } from 'vue'
import type { ChartData, Sort } from '../../types'
import { getBaseOptions, type BaseChartConfig } from './baseChartOptions'

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
