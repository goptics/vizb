import { describe, it, expect, beforeAll, afterAll, vi } from 'vitest'
import {
  baseConfig,
  emptyChartData,
  continuousRender3D,
  groupedRender3D,
  installDevicePixelRatio,
} from '@/test-utils'
import { useLine3DChartOptions } from './useLine3DChartOptions'

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

describe('useLine3DChartOptions', () => {
  it('emits line3D series in continuous mode', () => {
    const { options } = useLine3DChartOptions(
      baseConfig({ chartData: continuousData(), threeD: true })
    )
    const series = options.value.series as { type: string; data: unknown[] }[]
    expect(series[0]!.type).toBe('line3D')
    expect(series[0]!.data).toHaveLength(2)
    expect((options.value.xAxis3D as { type: string }).type).toBe('value')
  })

  it('emits line3D series per z group in grouped mode', () => {
    const { options } = useLine3DChartOptions(
      baseConfig({ chartData: groupedData(), threeD: true })
    )
    const series = options.value.series as { type: string; name: string }[]
    const lineSeries = series.filter((s) => s.type === 'line3D')
    expect(lineSeries).toHaveLength(2)
    expect(lineSeries.map((s) => s.name)).toEqual(['zA', 'zB'])
    expect((options.value.xAxis3D as { type: string }).type).toBe('category')
  })

  it('returns empty visualMap when threeDVisualMap is off', () => {
    const { options } = useLine3DChartOptions(
      baseConfig({ chartData: groupedData(), threeD: true, threeDVisualMap: false })
    )
    expect(options.value.visualMap).toEqual([])
  })

  it('returns visualMap object when threeDVisualMap is on', () => {
    const { options } = useLine3DChartOptions(
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
  barSeries: [{ name: 'sales', data: [{ value: [0, 0, 15] }] }],
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

describe('useLine3DChartOptions — remaining branches', () => {
  it('emits line3D + scatter labels in value mode', () => {
    const { options } = useLine3DChartOptions(
      baseConfig({
        chartData: emptyChartData({
          title: 'sales',
          statType: 'sum',
          statUnit: 'ms',
          axisLabels: { x: 'x', y: 'y' },
          render3D: valueModeRender,
        }),
        threeD: true,
        showLabels: true,
        threeDVisualMap: true,
        scale: 'log',
        threeDRotate: true,
      })
    )
    const series = options.value.series as { type: string }[]
    expect(series.some((s) => s.type === 'line3D')).toBe(true)
    expect(series.some((s) => s.type === 'scatter3D')).toBe(true)
    expect((options.value.zAxis3D as { name?: string }).name).toContain('ms')
  })

  it('emits mixed-mode line3D', () => {
    const { options } = useLine3DChartOptions(
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
    expect((options.value.series as { type: string }[])[0]!.type).toBe('line3D')
  })

  it('falls back to valuePoints3D when lineSeries empty', () => {
    const { options } = useLine3DChartOptions(
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
    expect(series[0]!.type).toBe('line3D')
    expect(series[0]!.data).toHaveLength(2)
  })

  it('labels top visible z scatter overlay', () => {
    const { options } = useLine3DChartOptions(
      baseConfig({
        chartData: groupedData(),
        threeD: true,
        showLabels: true,
        visibleZ: { zA: false, zB: true },
      })
    )
    const scatter = (
      options.value.series as { type: string; name: string; label: { show?: boolean } }[]
    ).filter((s) => s.type === 'scatter3D')
    expect(scatter.find((s) => s.name === 'zB')?.label.show).toBe(true)
  })
})

describe('useLine3DChartOptions — optional defaults', () => {
  it('value mode without optionals and empty series data', () => {
    const cfg = baseConfig({
      chartData: emptyChartData({
        title: 'sales',
        statType: 'sum',
        render3D: {
          mode: 'value',
          xValues: ['x1'],
          yValues: ['y1'],
          zValues: [],
          barSeries: [],
          lineSeries: [],
          cellTotals: {},
        },
      }),
      threeD: true,
    })
    delete cfg.threeDRotate
    delete cfg.scale
    delete cfg.visibleZ
    const { options } = useLine3DChartOptions(cfg)
    expect(options.value.series).toEqual([])
  })

  it('grouped without points/visibleZ and EMPTY_RENDER fallback', () => {
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
    const { options } = useLine3DChartOptions(cfg)
    expect((options.value.series as { type: string }[]).some((s) => s.type === 'line3D')).toBe(true)

    const empty = useLine3DChartOptions(
      baseConfig({ chartData: emptyChartData({ title: 'e', statType: 'avg' }), threeD: true })
    )
    expect(empty.options.value.series).toEqual([])
  })
})

describe('useLine3DChartOptions — more optional branches', () => {
  it('grouped defaults without rotate/scale/points', async () => {
    const utils = await import('@/lib/utils')
    const spy = vi.spyOn(utils, 'is3D').mockReturnValue(false)
    const cfg = baseConfig({ chartData: groupedData(), threeD: true })
    delete cfg.threeDRotate
    delete cfg.scale
    ;(cfg.chartData.value as { points?: unknown }).points = undefined
    const { options } = useLine3DChartOptions(cfg)
    expect((options.value.series as { type: string }[]).some((s) => s.type === 'line3D')).toBe(true)
    spy.mockRestore()
  })
})
