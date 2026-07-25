import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import { baseConfig, makePieChartData, installDevicePixelRatio } from '@/test-utils'
import { usePieChartOptions } from './usePieChartOptions'

let restoreDpr: () => void
beforeAll(() => {
  restoreDpr = installDevicePixelRatio()
})
afterAll(() => restoreDpr())

describe('usePieChartOptions', () => {
  it('emits pie series for x-only multi-slice data', () => {
    const { options } = usePieChartOptions(baseConfig({ chartData: makePieChartData() }))
    const series = options.value.series as { type: string; data: { name: string }[] }[]
    expect(series).toHaveLength(1)
    expect(series[0]!.type).toBe('pie')
    expect(series[0]!.data.map((d) => d.name)).toEqual(['A', 'B', 'C'])
  })

  it('shows slice labels when showLabels is on', () => {
    const { options } = usePieChartOptions(
      baseConfig({ chartData: makePieChartData(), showLabels: true })
    )
    const series = options.value.series as { label?: { show?: boolean } }[]
    expect(series[0]!.label?.show).toBe(true)
  })

  it('hides slice labels when showLabels is off', () => {
    const { options } = usePieChartOptions(
      baseConfig({ chartData: makePieChartData(), showLabels: false })
    )
    const series = options.value.series as { label?: { show?: boolean } }[]
    expect(series[0]!.label?.show).toBe(false)
  })
})
