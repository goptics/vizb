import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import { ref } from 'vue'
import type { ChartData, Render3D } from '@/types'
import type { BaseChartConfig } from './baseChartOptions'
import { useScatter3DChartOptions } from './useScatter3DChartOptions'

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

const makeConfig = (chartData: ChartData): BaseChartConfig => ({
  chartData: ref(chartData),
  sort: ref({ enabled: false, order: 'asc' }),
  showLabels: ref(false),
  isDark: ref(false),
  scale: ref('linear'),
  threeDRotate: ref(false),
  visibleZ: ref({}),
  threeD: ref(true),
  threeDVisualMap: ref(false),
})

const continuousRender: Render3D = {
  mode: 'continuous',
  xValues: [],
  yValues: [],
  zValues: [],
  barSeries: [{ name: 'pts', data: [{ value: [1, 2, 3] }, { value: [4, 5, 6] }] }],
  lineSeries: [{ name: 'pts', data: [{ value: [1, 2, 3] }, { value: [4, 5, 6] }] }],
  cellTotals: {},
}

const groupedRender: Render3D = {
  mode: 'grouped',
  xValues: ['x1', 'x2'],
  yValues: ['y1'],
  zValues: ['zA', 'zB'],
  barSeries: [
    { name: 'zA', data: [{ value: [0, 0, 5] }] },
    { name: 'zB', data: [{ value: [0, 0, 7] }] },
  ],
  lineSeries: [
    { name: 'zA', data: [{ value: [0, 0, 5] }] },
    { name: 'zB', data: [{ value: [0, 0, 7] }] },
  ],
  cellTotals: { '0,0': 12 },
}

describe('useScatter3DChartOptions — continuous mode', () => {
  it('emits scatter3D series on value axes', () => {
    const { options } = useScatter3DChartOptions(
      makeConfig({
        title: 'x · y · z',
        statType: 'value',
        yAxis: [],
        zAxis: [],
        series: [],
        points: [],
        axisLabels: { x: 'x', y: 'y', z: 'z' },
        render3D: continuousRender,
      })
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
      makeConfig({
        title: 'avg',
        statType: 'avg',
        yAxis: ['y1'],
        zAxis: ['zA', 'zB'],
        series: [],
        points: [
          { xAxis: 'x1', yAxis: 'y1', zAxis: 'zA', value: 5 },
          { xAxis: 'x1', yAxis: 'y1', zAxis: 'zB', value: 7 },
        ],
        axisLabels: { x: 'x', y: 'y', z: 'z' },
        render3D: groupedRender,
      })
    )
    const series = options.value.series as { type: string; name: string }[]
    expect(series).toHaveLength(2)
    expect(series.every((s) => s.type === 'scatter3D')).toBe(true)
    expect(series.map((s) => s.name)).toEqual(['zA', 'zB'])
    expect((options.value.xAxis3D as { type: string }).type).toBe('category')
    expect((options.value.legend as { show?: boolean }).show).toBe(true)
  })

  it('honors 2D visualMap flag on 3D scatter when threeDVisualMap is off', () => {
    const config = makeConfig({
      title: 'avg',
      statType: 'avg',
      yAxis: ['y1'],
      zAxis: ['zA'],
      series: [],
      points: [{ xAxis: 'x1', yAxis: 'y1', zAxis: 'zA', value: 5 }],
      axisLabels: { x: 'x', y: 'y', z: 'z' },
      render3D: groupedRender,
    })
    config.threeDVisualMap = ref(false)
    config.visualMap = ref(true)
    const { options } = useScatter3DChartOptions(config)
    expect(options.value.visualMap).toMatchObject({ show: true })
  })

  it('applies category visualMap (dimension 2) on grouped series when enabled', () => {
    const config = makeConfig({
      title: 'avg',
      statType: 'avg',
      yAxis: ['y1'],
      zAxis: ['zA', 'zB'],
      series: [],
      points: [
        { xAxis: 'x1', yAxis: 'y1', zAxis: 'zA', value: 5 },
        { xAxis: 'x1', yAxis: 'y1', zAxis: 'zB', value: 7 },
      ],
      axisLabels: { x: 'x', y: 'y', z: 'z' },
      render3D: groupedRender,
    })
    config.threeDVisualMap = ref(true)
    const { options } = useScatter3DChartOptions(config)
    expect(options.value.visualMap).toMatchObject({ show: true, dimension: 2 })
  })
})

describe('useScatter3DChartOptions — valuePoints3D fallback', () => {
  it('renders continuous scatter3D when valuePoints3D present without render3D', () => {
    const { options } = useScatter3DChartOptions(
      makeConfig({
        title: 'x · y · z',
        statType: 'value',
        yAxis: [],
        zAxis: [],
        series: [],
        points: [],
        axisLabels: { x: 'x', y: 'y', z: 'z' },
        valuePoints3D: [
          [1, 2, 3],
          [4, 5, 6],
        ],
      })
    )
    const series = options.value.series as { type: string; data: { value: number[] }[] }[]
    expect(series).toHaveLength(1)
    expect(series[0]!.type).toBe('scatter3D')
    expect(series[0]!.data).toHaveLength(2)
    expect((options.value.xAxis3D as { type: string }).type).toBe('value')
  })
})
