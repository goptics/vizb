import { describe, it, expect, beforeAll, afterAll, vi } from 'vitest'
import { ref } from 'vue'
import type { ChartData } from '@/types'
import {
  baseConfig,
  emptyChartData,
  continuousRender3D,
  groupedRender3D,
  installDevicePixelRatio,
} from '@/test-utils'
import { useScatter3DChartOptions } from './useScatter3DChartOptions'

let restoreDpr: () => void
beforeAll(() => {
  restoreDpr = installDevicePixelRatio()
})
afterAll(() => restoreDpr())

const makeConfig = (chartData: ChartData) =>
  baseConfig({
    chartData,
    threeD: true,
    threeDVisualMap: false,
  })

describe('useScatter3DChartOptions — continuous mode', () => {
  it('emits scatter3D series on value axes', () => {
    const { options } = useScatter3DChartOptions(
      makeConfig(
        emptyChartData({
          title: 'x · y · z',
          statType: 'value',
          axisLabels: { x: 'x', y: 'y', z: 'z' },
          render3D: continuousRender3D,
        })
      )
    )
    const series = options.value.series as { type: string; data: { value: number[] }[] }[]
    expect(series).toHaveLength(1)
    expect(series[0]!.type).toBe('scatter3D')
    expect(series[0]!.data).toHaveLength(2)
    expect((options.value.xAxis3D as { type: string }).type).toBe('value')
    expect((options.value.yAxis3D as { type: string }).type).toBe('value')
    expect((options.value.zAxis3D as { type: string }).type).toBe('value')
  })
})

describe('useScatter3DChartOptions — grouped mode', () => {
  it('emits one scatter3D series per z group with category axes', () => {
    const { options } = useScatter3DChartOptions(
      makeConfig(
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
      )
    )
    const series = options.value.series as { type: string; name: string }[]
    expect(series).toHaveLength(2)
    expect(series.every((s) => s.type === 'scatter3D')).toBe(true)
    expect(series.map((s) => s.name)).toEqual(['zA', 'zB'])
    expect((options.value.xAxis3D as { type: string }).type).toBe('category')
    expect((options.value.legend as { show?: boolean }).show).toBe(true)
  })

  it('honors 2D visualMap flag on 3D scatter when threeDVisualMap is off', () => {
    const config = makeConfig(
      emptyChartData({
        title: 'avg',
        statType: 'avg',
        yAxis: ['y1'],
        zAxis: ['zA'],
        points: [{ xAxis: 'x1', yAxis: 'y1', zAxis: 'zA', value: 5 }],
        axisLabels: { x: 'x', y: 'y', z: 'z' },
        render3D: groupedRender3D,
      })
    )
    config.threeDVisualMap = ref(false)
    config.visualMap = ref(true)
    const { options } = useScatter3DChartOptions(config)
    expect(options.value.visualMap).toMatchObject({ show: true })
  })

  it('applies category visualMap (dimension 2) on grouped series when enabled', () => {
    const config = makeConfig(
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
    )
    config.threeDVisualMap = ref(true)
    const { options } = useScatter3DChartOptions(config)
    expect(options.value.visualMap).toMatchObject({ show: true, dimension: 2 })
  })
})

describe('useScatter3DChartOptions — valuePoints3D fallback', () => {
  it('renders continuous scatter3D when valuePoints3D present without render3D', () => {
    const { options } = useScatter3DChartOptions(
      makeConfig(
        emptyChartData({
          title: 'x · y · z',
          statType: 'value',
          axisLabels: { x: 'x', y: 'y', z: 'z' },
          valuePoints3D: [
            [1, 2, 3],
            [4, 5, 6],
          ],
        })
      )
    )
    const series = options.value.series as { type: string; data: { value: number[] }[] }[]
    expect(series).toHaveLength(1)
    expect(series[0]!.type).toBe('scatter3D')
    expect(series[0]!.data).toHaveLength(2)
    expect((options.value.xAxis3D as { type: string }).type).toBe('value')
  })
})

const valueModeRender = {
  mode: 'value' as const,
  xValues: ['x1', 'x2'],
  yValues: ['y1'],
  zValues: [],
  barSeries: [],
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

describe('useScatter3DChartOptions — remaining branches', () => {
  it('emits scatter3D in value mode with statUnit and log scale', () => {
    const { options } = useScatter3DChartOptions(
      baseConfig({
        chartData: emptyChartData({
          title: 'sales',
          statType: 'sum',
          statUnit: 'kb',
          axisLabels: { x: 'x', y: 'y' },
          render3D: valueModeRender,
        }),
        threeD: true,
        showLabels: true,
        threeDVisualMap: false,
        scale: 'log',
        threeDRotate: true,
      })
    )
    const series = options.value.series as { type: string; itemStyle?: { color?: string } }[]
    expect(series[0]!.type).toBe('scatter3D')
    expect(series[0]!.itemStyle?.color).toBeDefined()
    expect((options.value.zAxis3D as { type?: string }).type).toBe('log')
    expect((options.value.zAxis3D as { name?: string }).name).toContain('kb')
  })

  it('uses visualMap colors in value mode when enabled', () => {
    const { options } = useScatter3DChartOptions(
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
    const series = options.value.series as { itemStyle?: unknown }[]
    expect(series[0]!.itemStyle).toBeUndefined()
    expect(options.value.visualMap).toMatchObject({ show: true })
  })

  it('emits mixed-mode scatter3D', () => {
    const { options } = useScatter3DChartOptions(
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
    expect((options.value.series as { type: string }[])[0]!.type).toBe('scatter3D')
  })

  it('labels only the top visible z series', () => {
    const { options } = useScatter3DChartOptions(
      baseConfig({
        chartData: emptyChartData({
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
        }),
        threeD: true,
        showLabels: true,
        visibleZ: { zA: true, zB: false },
      })
    )
    const series = options.value.series as { name: string; label: { show?: boolean } }[]
    expect(series.find((s) => s.name === 'zA')?.label.show).toBe(true)
    expect(series.find((s) => s.name === 'zB')?.label.show).toBe(false)
  })
})

describe('useScatter3DChartOptions — optional defaults', () => {
  it('value mode without optionals and empty series', () => {
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
    const { options } = useScatter3DChartOptions(cfg)
    expect(options.value.series).toEqual([])
  })

  it('grouped without points/visibleZ/cellTotals', () => {
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
    const { options } = useScatter3DChartOptions(cfg)
    expect((options.value.series as { type: string }[]).every((s) => s.type === 'scatter3D')).toBe(
      true
    )
  })
})

describe('useScatter3DChartOptions — more optional branches', () => {
  it('grouped with undefined points and all-z-hidden label fallback', async () => {
    const utils = await import('@/lib/utils')
    const spy = vi.spyOn(utils, 'is3D').mockReturnValue(false)
    const cfg = baseConfig({
      chartData: emptyChartData({
        title: 'avg',
        statType: 'avg',
        yAxis: ['y1'],
        zAxis: ['zA', 'zB'],
        axisLabels: { x: 'x', y: 'y', z: 'z' },
        render3D: groupedRender3D,
      }),
      threeD: true,
      showLabels: true,
      visibleZ: { zA: false, zB: false },
    })
    ;(cfg.chartData.value as { points?: unknown }).points = undefined
    const { options } = useScatter3DChartOptions(cfg)
    const series = options.value.series as { name: string; label: { show?: boolean } }[]
    // last z name used when none visible
    expect(series.find((s) => s.name === 'zB')?.label.show).toBe(true)
    spy.mockRestore()
  })
})
