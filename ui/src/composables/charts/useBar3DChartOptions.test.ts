import { describe, it, expect, beforeAll, afterAll, vi } from 'vitest'
import {
  baseConfig,
  emptyChartData,
  continuousRender3D,
  groupedRender3D,
  installDevicePixelRatio,
} from '@/test-utils'
import { useBar3DChartOptions } from './useBar3DChartOptions'

let restoreDpr: () => void
beforeAll(() => {
  restoreDpr = installDevicePixelRatio()
})
afterAll(() => restoreDpr())

const continuousData = () =>
  emptyChartData({
    title: 'x · y · z',
    statType: 'value',
    axisLabels: { x: 'x', y: 'y', z: 'z' },
    render3D: continuousRender3D,
  })

const groupedData = () =>
  emptyChartData({
    title: 'avg',
    statType: 'avg',
    yAxis: ['y1'],
    zAxis: ['zA', 'zB'],
    points: [
      { xAxis: 'x1', yAxis: 'y1', zAxis: 'zA', value: 5 },
      { xAxis: 'x1', yAxis: 'y1', zAxis: 'zB', value: 7 },
    ],
    axisLabels: { x: 'x', y: 'y', z: 'z' },
    render3D: groupedRender3D,
  })

describe('useBar3DChartOptions', () => {
  it('emits bar3D series in continuous mode', () => {
    const { options } = useBar3DChartOptions(
      baseConfig({ chartData: continuousData(), threeD: true })
    )
    const series = options.value.series as { type: string; data: unknown[] }[]
    expect(series[0]!.type).toBe('bar3D')
    expect(series[0]!.data).toHaveLength(2)
    expect((options.value.xAxis3D as { type: string }).type).toBe('value')
  })

  it('emits one bar3D series per z group in grouped mode', () => {
    const { options } = useBar3DChartOptions(baseConfig({ chartData: groupedData(), threeD: true }))
    const series = options.value.series as { type: string; name: string }[]
    expect(series.every((s) => s.type === 'bar3D')).toBe(true)
    expect(series.map((s) => s.name)).toEqual(['zA', 'zB'])
    expect((options.value.xAxis3D as { type: string }).type).toBe('category')
  })

  it('returns empty visualMap when threeDVisualMap is off', () => {
    const { options } = useBar3DChartOptions(
      baseConfig({ chartData: groupedData(), threeD: true, threeDVisualMap: false })
    )
    expect(options.value.visualMap).toEqual([])
  })

  it('returns visualMap object when threeDVisualMap is on', () => {
    const { options } = useBar3DChartOptions(
      baseConfig({ chartData: groupedData(), threeD: true, threeDVisualMap: true })
    )
    expect(options.value.visualMap).toMatchObject({ show: true })
  })
})

const valueModeRender = {
  mode: 'value' as const,
  xValues: ['x1', 'x2'],
  yValues: ['y1'],
  zValues: [],
  barSeries: [{ name: 'sales', data: [{ value: [0, 0, 15] }, { value: [1, 0, 25] }] }],
  lineSeries: [{ name: 'sales', data: [{ value: [0, 0, 15] }, { value: [1, 0, 25] }] }],
  cellTotals: { '0,0': 15, '1,0': 25 },
}

const mixedRender = {
  mode: 'mixed' as const,
  xValues: ['West', 'East'],
  yValues: [],
  zValues: [],
  barSeries: [],
  lineSeries: [{ name: 'pts', data: [{ value: [0, 1.5, 2.5] }, { value: [1, 2.0, 3.0] }] }],
  cellTotals: {},
}

describe('useBar3DChartOptions — remaining branches', () => {
  it('emits bar3D in value mode with statUnit label', () => {
    const { options } = useBar3DChartOptions(
      baseConfig({
        chartData: emptyChartData({
          title: 'sales',
          statType: 'sum',
          statUnit: 'usd',
          axisLabels: { x: 'x', y: 'y' },
          render3D: valueModeRender,
        }),
        threeD: true,
        showLabels: true,
        threeDVisualMap: true,
        threeDRotate: true,
        scale: 'log',
      })
    )
    const series = options.value.series as { type: string; barSize?: number[] }[]
    expect(series[0]!.type).toBe('bar3D')
    expect(series[0]!.barSize).toBeDefined()
    expect((options.value.zAxis3D as { name?: string }).name).toContain('usd')
    expect((options.value.zAxis3D as { type?: string }).type).toBe('log')
    expect(options.value.visualMap).toMatchObject({ show: true })
  })

  it('emits mixed-mode bar3D', () => {
    const { options } = useBar3DChartOptions(
      baseConfig({
        chartData: emptyChartData({
          title: 'mixed',
          statType: 'mixed',
          axisLabels: { x: 'region', y: 'tax', z: 'score' },
          render3D: mixedRender,
        }),
        threeD: true,
      })
    )
    const series = options.value.series as { type: string }[]
    expect(series[0]!.type).toBe('bar3D')
    expect((options.value.xAxis3D as { type: string }).type).toBe('category')
    expect((options.value.yAxis3D as { type: string }).type).toBe('value')
  })

  it('falls back to valuePoints3D when barSeries empty', () => {
    const { options } = useBar3DChartOptions(
      baseConfig({
        chartData: emptyChartData({
          title: 'pts',
          statType: 'value',
          axisLabels: { x: 'x', y: 'y', z: 'z' },
          valuePoints3D: [
            [1, 2, 3],
            [4, 5, 6],
          ],
          render3D: {
            mode: 'continuous',
            xValues: [],
            yValues: [],
            zValues: [],
            barSeries: [],
            lineSeries: [],
            cellTotals: {},
          },
        }),
        threeD: true,
      })
    )
    const series = options.value.series as { type: string; data: unknown[] }[]
    expect(series[0]!.type).toBe('bar3D')
    expect(series[0]!.data).toHaveLength(2)
  })

  it('labels top visible z series when showLabels is on', () => {
    const { options } = useBar3DChartOptions(
      baseConfig({
        chartData: groupedData(),
        threeD: true,
        showLabels: true,
        visibleZ: { zA: true, zB: false },
      })
    )
    const series = options.value.series as { name: string; label: { show?: boolean } }[]
    const zA = series.find((s) => s.name === 'zA')
    expect(zA?.label.show).toBe(true)
  })

  it('uses EMPTY_RENDER when render3D missing', () => {
    const { options } = useBar3DChartOptions(
      baseConfig({
        chartData: emptyChartData({ title: 'empty', statType: 'avg' }),
        threeD: true,
      })
    )
    expect(options.value.series).toEqual([])
  })
})

describe('useBar3DChartOptions — optional defaults', () => {
  it('value mode without optional refs/statUnit/cellTotals/empty series', () => {
    const cfg = baseConfig({
      chartData: emptyChartData({
        title: 'sales',
        statType: 'sum',
        axisLabels: { x: 'x', y: 'y' },
        render3D: {
          mode: 'value',
          xValues: ['x1'],
          yValues: ['y1'],
          zValues: [],
          barSeries: [],
          lineSeries: [],
          cellTotals: undefined as unknown as Record<string, number>,
        },
      }),
      threeD: true,
    })
    delete cfg.threeDRotate
    delete cfg.scale
    delete cfg.threeDVisualMap
    delete cfg.visibleZ
    const { options } = useBar3DChartOptions(cfg)
    expect((options.value.series as unknown[]).length).toBe(0)
    // no barSize when not value with series? empty series still value mode
    expect((options.value.zAxis3D as { name?: string }).name).toBe('sales')
  })

  it('grouped mode without points/visibleZ/cellTotals', () => {
    const cfg = baseConfig({
      chartData: emptyChartData({
        title: 'avg',
        statType: 'avg',
        render3D: {
          ...groupedRender3D,
          cellTotals: undefined as unknown as Record<string, number>,
        },
      }),
      threeD: true,
    })
    delete cfg.visibleZ
    cfg.chartData.value.points = []
    const { options } = useBar3DChartOptions(cfg)
    expect((options.value.series as { type: string }[]).every((s) => s.type === 'bar3D')).toBe(true)
  })
})

describe('useBar3DChartOptions — more optional branches', () => {
  it('grouped continuousCtx defaults when rotate/scale omitted', () => {
    const cfg = baseConfig({
      chartData: groupedData(),
      threeD: true,
    })
    delete cfg.threeDRotate
    delete cfg.scale
    cfg.chartData.value.points = []
    const { options } = useBar3DChartOptions(cfg)
    expect((options.value.series as unknown[]).length).toBe(2)
  })

  it('treats nullish points as empty in grouped mode', async () => {
    const utils = await import('@/lib/utils')
    const spy = vi.spyOn(utils, 'is3D').mockReturnValue(false)
    const cfg = baseConfig({ chartData: groupedData(), threeD: true })
    ;(cfg.chartData.value as { points?: unknown }).points = undefined
    const { options } = useBar3DChartOptions(cfg)
    expect((options.value.series as unknown[]).length).toBe(2)
    spy.mockRestore()
  })

  it('value mode with visualMap omits per-series itemStyle', () => {
    const { options } = useBar3DChartOptions(
      baseConfig({
        chartData: emptyChartData({
          title: 'sales',
          statType: 'sum',
          render3D: valueModeRender,
        }),
        threeD: true,
        threeDVisualMap: true,
      })
    )
    const series = options.value.series as { itemStyle?: unknown; barSize?: unknown }[]
    expect(series[0]!.itemStyle).toBeUndefined()
    expect(series[0]!.barSize).toBeDefined()
  })
})

describe('useBar3DChartOptions — value mode itemStyle', () => {
  it('value mode with visualMap off keeps itemStyle', () => {
    const { options } = useBar3DChartOptions(
      baseConfig({
        chartData: emptyChartData({
          title: 'sales',
          statType: 'sum',
          render3D: valueModeRender,
        }),
        threeD: true,
        threeDVisualMap: false,
      })
    )
    const series = options.value.series as { itemStyle?: { color?: string }; barSize?: unknown }[]
    expect(series[0]!.itemStyle?.color).toBeDefined()
    expect(series[0]!.barSize).toBeDefined()
  })
})
