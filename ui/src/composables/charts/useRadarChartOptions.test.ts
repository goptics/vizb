import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import { baseConfig, makeRadarChartData, installDevicePixelRatio } from '@/test-utils'
import { useRadarChartOptions } from './useRadarChartOptions'

let restoreDpr: () => void
beforeAll(() => {
  restoreDpr = installDevicePixelRatio()
})
afterAll(() => restoreDpr())

describe('useRadarChartOptions', () => {
  it('emits radar series for x+y data', () => {
    const { options } = useRadarChartOptions(baseConfig({ chartData: makeRadarChartData() }))
    const series = options.value.series as { type: string }[]
    expect(series.length).toBeGreaterThan(0)
    expect(series.every((s) => s.type === 'radar')).toBe(true)
  })

  it('builds one indicator per y-axis spoke', () => {
    const chartData = makeRadarChartData()
    const { options } = useRadarChartOptions(baseConfig({ chartData }))
    const radar = options.value.radar as { indicator?: { name: string }[] }
    expect(radar.indicator).toHaveLength(chartData.yAxis.length)
    expect(radar.indicator?.map((i) => i.name)).toEqual(['speed', 'memory', 'allocs'])
  })

  it('emits one radar series per x value', () => {
    const chartData = makeRadarChartData()
    const { options } = useRadarChartOptions(baseConfig({ chartData }))
    const series = options.value.series as { name: string; type: string }[]
    expect(series.map((s) => s.name)).toEqual(chartData.series.map((s) => s.xAxis))
  })
})
