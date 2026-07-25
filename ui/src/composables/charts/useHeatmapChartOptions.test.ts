import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import { baseConfig, makeHeatmapChartData, installDevicePixelRatio } from '@/test-utils'
import { LARGE_X_THRESHOLD } from './shared'
import { useHeatmapChartOptions } from './useHeatmapChartOptions'

let restoreDpr: () => void
beforeAll(() => {
  restoreDpr = installDevicePixelRatio()
})
afterAll(() => restoreDpr())

describe('useHeatmapChartOptions', () => {
  it('emits heatmap series type', () => {
    const { options } = useHeatmapChartOptions(baseConfig({ chartData: makeHeatmapChartData() }))
    const series = options.value.series as { type: string }[]
    expect(series[0]!.type).toBe('heatmap')
  })

  it('omits dataZoom for small axes', () => {
    const { options } = useHeatmapChartOptions(baseConfig({ chartData: makeHeatmapChartData() }))
    expect(options.value.dataZoom).toBeUndefined()
  })

  it('attaches dataZoom when x categories exceed LARGE_X_THRESHOLD', () => {
    const manyX = Array.from({ length: LARGE_X_THRESHOLD + 1 }, (_, i) => `x${i}`)
    const chartData = makeHeatmapChartData({
      series: manyX.map((x) => ({ xAxis: x, values: [1, 2], benchmarkId: '' })),
    })
    const { options } = useHeatmapChartOptions(baseConfig({ chartData }))
    expect(options.value.dataZoom).toBeDefined()
    expect(
      Array.isArray(options.value.dataZoom) ? options.value.dataZoom.length : 1
    ).toBeGreaterThan(0)
  })

  it('attaches dataZoom when y categories exceed LARGE_X_THRESHOLD', () => {
    const manyY = Array.from({ length: LARGE_X_THRESHOLD + 1 }, (_, i) => `y${i}`)
    const chartData = makeHeatmapChartData({
      yAxis: manyY,
      series: [
        {
          xAxis: 'x1',
          values: manyY.map((_, i) => i),
          benchmarkId: '',
        },
      ],
    })
    const { options } = useHeatmapChartOptions(baseConfig({ chartData }))
    expect(options.value.dataZoom).toBeDefined()
  })
})
