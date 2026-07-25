import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import { ref, type Ref } from 'vue'
import type { ChartData, Sort } from '@/types'
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

describe('getBaseOptions pixelRatio branch', () => {
  it('falls back to dpr 1 when devicePixelRatio is falsy', () => {
    const g = globalThis as unknown as { window: { devicePixelRatio: number } }
    const prev = g.window.devicePixelRatio
    try {
      g.window.devicePixelRatio = 0
      const cfg = makeMinimalConfig()
      // Continuous value-mode 3D so is3D() is true and saveAsImage uses dpr.
      cfg.chartData.value = {
        ...cfg.chartData.value,
        statType: 'value',
        valuePoints3D: [[1, 2, 3]],
        render3D: {
          mode: 'continuous',
          xValues: [],
          yValues: [],
          zValues: [],
          barSeries: [],
          lineSeries: [{ name: 'pts', data: [{ value: [1, 2, 3] }] }],
          cellTotals: {},
        },
      }
      cfg.chartType = { value: 'scatter' } as BaseChartConfig['chartType']
      cfg.arrangementTarget = { value: 'xyz' } as BaseChartConfig['arrangementTarget']
      cfg.chartAxes = {
        value: [
          { key: 'x', label: 'x', type: 'value' },
          { key: 'y', label: 'y', type: 'value' },
          { key: 'z', label: 'z', type: 'value' },
        ],
      } as BaseChartConfig['chartAxes']
      const opts = getBaseOptions(cfg)
      const toolbox = opts.toolbox as { feature?: { saveAsImage?: { pixelRatio?: number } } }
      expect(toolbox.feature?.saveAsImage?.pixelRatio).toBe(1)
    } finally {
      g.window.devicePixelRatio = prev ?? 1
    }
  })
})
