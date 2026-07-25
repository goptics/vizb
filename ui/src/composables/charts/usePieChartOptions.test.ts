import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import { baseConfig, makePieChartData, emptyChartData, installDevicePixelRatio } from '@/test-utils'
import { usePieChartOptions } from './usePieChartOptions'

let restoreDpr: () => void
beforeAll(() => {
  restoreDpr = installDevicePixelRatio()
})
afterAll(() => restoreDpr())

const xyPie = () =>
  emptyChartData({
    title: 'share',
    statType: 'count',
    yAxis: ['North', 'South'],
    series: [
      { xAxis: 'A', values: [10, 5], benchmarkId: '' },
      { xAxis: 'B', values: [20, 15], benchmarkId: '' },
    ],
    axisLabels: { x: 'group', y: 'region' },
  })

const xyzPie = () =>
  emptyChartData({
    title: 'share3d',
    statType: 'count',
    yAxis: ['North', 'South'],
    zAxis: ['zA', 'zB', ''],
    series: [
      { xAxis: 'A', values: [10, 5], benchmarkId: '' },
      { xAxis: 'B', values: [20, 15], benchmarkId: '' },
    ],
    points: [
      { xAxis: 'A', yAxis: 'North', zAxis: 'zA', value: 10 },
      { xAxis: 'A', yAxis: 'South', zAxis: 'zB', value: 5 },
      { xAxis: 'B', yAxis: 'North', zAxis: 'zA', value: 20 },
      { xAxis: 'B', yAxis: 'South', zAxis: 'zB', value: 15 },
    ],
    axisLabels: { x: 'group', y: 'region', z: 'pool' },
  })

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

  it('sorts x-only slices by total', () => {
    const { options } = usePieChartOptions(
      baseConfig({
        chartData: makePieChartData(),
        sort: { enabled: true, order: 'asc' },
      })
    )
    const series = options.value.series as { data: { name: string; value: number }[] }[]
    expect(series[0]!.data.map((d) => d.name)).toEqual(['A', 'B', 'C'])
    const desc = usePieChartOptions(
      baseConfig({
        chartData: makePieChartData(),
        sort: { enabled: true, order: 'desc' },
      })
    )
    expect(
      (desc.options.value.series as { data: { name: string }[] }[])[0]!.data.map((d) => d.name)
    ).toEqual(['C', 'B', 'A'])
  })

  it('builds dual pies for x+y data', () => {
    const { options } = usePieChartOptions(baseConfig({ chartData: xyPie() }))
    const series = options.value.series as { name?: string; data: { name: string }[] }[]
    expect(series).toHaveLength(2)
    expect(series[0]!.data.map((d) => d.name)).toEqual(['A', 'B'])
    expect(series[1]!.data.map((d) => d.name)).toEqual(['North', 'South'])
    const titles = options.value.title as { text: string }[]
    expect(titles.map((t) => t.text)).toEqual(['group', 'region'])
  })

  it('sorts y-axis dual-pie slices', () => {
    const { options } = usePieChartOptions(
      baseConfig({
        chartData: xyPie(),
        sort: { enabled: true, order: 'asc' },
      })
    )
    const series = options.value.series as { data: { name: string; value: number }[] }[]
    // North=30, South=20
    expect(series[1]!.data.map((d) => d.name)).toEqual(['South', 'North'])
  })

  it('uses default dual titles when axisLabels missing', () => {
    const data = xyPie()
    data.axisLabels = {}
    const { options } = usePieChartOptions(baseConfig({ chartData: data }))
    const titles = options.value.title as { text: string }[]
    expect(titles.map((t) => t.text)).toEqual(['X-Axis', 'Y-Axis'])
  })

  it('builds y-only pie when x axis is empty', () => {
    const chartData = emptyChartData({
      title: 'yonly',
      statType: 'count',
      yAxis: ['North', 'South'],
      series: [
        { xAxis: '', values: [10, 5], benchmarkId: '' },
        { xAxis: '', values: [20, 15], benchmarkId: '' },
      ],
    })
    const { options } = usePieChartOptions(
      baseConfig({ chartData, sort: { enabled: true, order: 'desc' } })
    )
    const series = options.value.series as { data: { name: string; value: number }[] }[]
    expect(series).toHaveLength(1)
    expect(series[0]!.data.map((d) => d.name)).toEqual(['North', 'South'])
    expect(series[0]!.data.map((d) => d.value)).toEqual([30, 20])
  })

  it('builds triple pies for x+y+z data', () => {
    const { options } = usePieChartOptions(baseConfig({ chartData: xyzPie() }))
    const series = options.value.series as { data: { name: string }[] }[]
    expect(series).toHaveLength(3)
    expect(series[2]!.data.map((d) => d.name)).toEqual(['zA', 'zB'])
    const titles = options.value.title as { text: string }[]
    expect(titles.map((t) => t.text)).toEqual(['group', 'region', 'pool'])
  })

  it('sorts z-axis slices and uses default 3D titles', () => {
    const data = xyzPie()
    data.axisLabels = {}
    const { options } = usePieChartOptions(
      baseConfig({ chartData: data, sort: { enabled: true, order: 'asc' } })
    )
    const series = options.value.series as { data: { name: string; value: number }[] }[]
    // zA=30, zB=20
    expect(series[2]!.data.map((d) => d.name)).toEqual(['zB', 'zA'])
    const titles = options.value.title as { text: string }[]
    expect(titles.map((t) => t.text)).toEqual(['X-Axis', 'Y-Axis', 'Z-Axis'])
  })

  it('formats slice labels and item tooltips with percent', () => {
    const { options } = usePieChartOptions(
      baseConfig({ chartData: makePieChartData(), showLabels: true })
    )
    const series = options.value.series as {
      label?: { formatter?: (p: { name: string; percent: number }) => string }
    }[]
    const labelFmt = series[0]!.label?.formatter
    expect(labelFmt?.({ name: 'A', percent: 16.666 })).toBe('A (16.67%)')

    const tooltip = options.value.tooltip as {
      formatter: (p: { marker: string; name: string; value: number; percent: number }) => string
    }
    const html = tooltip.formatter({
      marker: '*',
      name: 'A',
      value: 10,
      percent: 16.666,
    })
    expect(html).toContain('<strong>A</strong>')
    expect(html).toContain('(16.67%)')
  })

  it('handles null values when totaling y-axis series', () => {
    const chartData = emptyChartData({
      title: 'nulls',
      statType: 'count',
      yAxis: ['N'],
      series: [{ xAxis: 'A', values: [null], benchmarkId: '' }],
    })
    const { options } = usePieChartOptions(baseConfig({ chartData }))
    const series = options.value.series as { data: { value: number }[] }[]
    expect(series[1]!.data[0]!.value).toBe(0)
  })
})

describe('usePieChartOptions — branch edges', () => {
  it('defaults empty x-only totals and missing z totals without points', () => {
    const xOnly = makePieChartData({
      series: [{ xAxis: 'A', values: [], benchmarkId: '' }],
    })
    const { options } = usePieChartOptions(baseConfig({ chartData: xOnly }))
    expect((options.value.series as { data: { value: number }[] }[])[0]!.data[0]!.value).toBe(0)

    const xyz = emptyChartData({
      title: 'z',
      statType: 'count',
      yAxis: ['ghost'],
      zAxis: ['missingZ'],
      series: [{ xAxis: 'A', values: [1], benchmarkId: '' }],
    })
    delete (xyz as { points?: unknown }).points
    const { options: o3 } = usePieChartOptions(baseConfig({ chartData: xyz }))
    const series = o3.value.series as { data: { value: number }[] }[]
    expect(series).toHaveLength(3)
    expect(series[2]!.data[0]!.value).toBe(0)
  })
})

describe('usePieChartOptions — y map fallback', () => {
  it('still emits y pie when y totals map is empty-like', () => {
    const chartData = emptyChartData({
      title: 't',
      statType: 'count',
      yAxis: ['ghost'],
      series: [{ xAxis: 'A', values: [], benchmarkId: '' }],
    })
    const { options } = usePieChartOptions(baseConfig({ chartData }))
    const series = options.value.series as { data: { value: number }[] }[]
    expect(series[1]!.data[0]!.value).toBe(0)
  })
})
