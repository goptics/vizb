import { describe, it, expect, beforeAll, afterAll } from 'vitest'
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
