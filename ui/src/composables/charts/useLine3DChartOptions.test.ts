import { describe, it, expect, beforeAll, afterAll } from 'vitest'
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
