import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import {
  baseConfig,
  makeRadarChartData,
  emptyChartData,
  installDevicePixelRatio,
} from '@/test-utils'
import { useRadarChartOptions } from './useRadarChartOptions'

let restoreDpr: () => void
beforeAll(() => {
  restoreDpr = installDevicePixelRatio()
})
afterAll(() => restoreDpr())

const radarXYZ = () =>
  emptyChartData({
    title: 'profile3d',
    statType: 'score',
    yAxis: ['speed', 'memory', 'allocs'],
    zAxis: ['pool-a', 'pool-b', ''],
    series: [
      { xAxis: 'algo-a', values: [10, 20, 5], benchmarkId: '' },
      { xAxis: 'algo-b', values: [15, 12, 8], benchmarkId: '' },
    ],
    points: [
      { xAxis: 'algo-a', yAxis: 'speed', zAxis: 'pool-a', value: 10 },
      { xAxis: 'algo-a', yAxis: 'memory', zAxis: 'pool-a', value: 20 },
      { xAxis: 'algo-a', yAxis: 'allocs', zAxis: 'pool-a', value: 5 },
      { xAxis: 'algo-b', yAxis: 'speed', zAxis: 'pool-a', value: 4 },
      { xAxis: 'algo-a', yAxis: 'speed', zAxis: 'pool-b', value: 30 },
      { xAxis: 'algo-a', yAxis: 'memory', zAxis: 'pool-b', value: 8 },
      { xAxis: 'algo-a', yAxis: 'unknown', zAxis: 'pool-b', value: 99 },
    ],
    axisLabels: { x: 'algo', y: 'metric', z: 'pool' },
  })

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

  it('sorts x+y series by total when sort is enabled', () => {
    const { options } = useRadarChartOptions(
      baseConfig({
        chartData: makeRadarChartData(),
        sort: { enabled: true, order: 'asc' },
      })
    )
    const series = options.value.series as { name: string }[]
    // algo-a total 35, algo-b total 35 — stable enough; use unequal data
    expect(series.length).toBe(2)

    const unequal = makeRadarChartData({
      series: [
        { xAxis: 'small', values: [1, 1, 1], benchmarkId: '' },
        { xAxis: 'big', values: [10, 10, 10], benchmarkId: '' },
      ],
    })
    const { options: asc } = useRadarChartOptions(
      baseConfig({ chartData: unequal, sort: { enabled: true, order: 'asc' } })
    )
    const { options: desc } = useRadarChartOptions(
      baseConfig({ chartData: unequal, sort: { enabled: true, order: 'desc' } })
    )
    expect((asc.value.series as { name: string }[]).map((s) => s.name)).toEqual(['small', 'big'])
    expect((desc.value.series as { name: string }[]).map((s) => s.name)).toEqual(['big', 'small'])
  })

  it('handles null values in x+y spokes', () => {
    const chartData = makeRadarChartData({
      series: [{ xAxis: 'a', values: [null, 5, null], benchmarkId: '' }],
    })
    const { options } = useRadarChartOptions(baseConfig({ chartData }))
    const series = options.value.series as { data: { value: number[] }[] }[]
    expect(series[0]!.data[0]!.value).toEqual([0, 5, 0])
  })

  it('builds x-only radar with x values as spokes', () => {
    const chartData = emptyChartData({
      title: 'xonly',
      statType: 'score',
      yAxis: [],
      series: [
        { xAxis: 'A', values: [3], benchmarkId: '' },
        { xAxis: 'B', values: [9], benchmarkId: '' },
        { xAxis: 'C', values: [-2], benchmarkId: '' },
      ],
    })
    const { options } = useRadarChartOptions(
      baseConfig({ chartData, sort: { enabled: true, order: 'desc' } })
    )
    const radar = options.value.radar as { indicator?: { name: string; max: number }[] }
    expect(radar.indicator?.map((i) => i.name)).toEqual(['B', 'A', 'C'])
    const series = options.value.series as { data: { value: number[]; name: string }[] }[]
    expect(series).toHaveLength(1)
    expect(series[0]!.data[0]!.name).toBe('score')
    expect(series[0]!.data[0]!.value).toEqual([9, 3, 0])
    expect((options.value.legend as { show?: boolean }).show).toBe(false)
  })

  it('builds y-only radar with column totals as a single polygon', () => {
    const chartData = emptyChartData({
      title: 'yonly',
      statType: 'score',
      yAxis: ['s', 'm'],
      series: [
        { xAxis: '', values: [1, 4], benchmarkId: '' },
        { xAxis: '', values: [2, 6], benchmarkId: '' },
      ],
    })
    const { options } = useRadarChartOptions(baseConfig({ chartData }))
    const radar = options.value.radar as { indicator?: { name: string }[] }
    expect(radar.indicator?.map((i) => i.name)).toEqual(['s', 'm'])
    const series = options.value.series as { data: { value: number[] }[] }[]
    expect(series).toHaveLength(1)
    expect(series[0]!.data[0]!.value).toEqual([3, 10])
  })

  it('builds z-legend radar for x+y+z points', () => {
    const { options } = useRadarChartOptions(baseConfig({ chartData: radarXYZ() }))
    const legend = options.value.legend as { data?: string[] }
    expect(legend.data).toEqual(['pool-a', 'pool-b'])
    const series = options.value.series as {
      name: string
      type: string
      data: { name: string; value: number[] }[]
    }[]
    expect(series.every((s) => s.type === 'radar')).toBe(true)
    expect(series.map((s) => s.name).sort()).toEqual(['pool-a', 'pool-b'])
    // render order is largest-total first (pool-a total 39 vs pool-b 38)
    expect(series[0]!.name).toBe('pool-a')
  })

  it('sorts z legend by totals asc/desc', () => {
    const data = radarXYZ()
    const { options: asc } = useRadarChartOptions(
      baseConfig({ chartData: data, sort: { enabled: true, order: 'asc' } })
    )
    const { options: desc } = useRadarChartOptions(
      baseConfig({ chartData: data, sort: { enabled: true, order: 'desc' } })
    )
    const ascLegend = (asc.value.legend as { data: string[] }).data
    const descLegend = (desc.value.legend as { data: string[] }).data
    expect(ascLegend).toEqual(['pool-b', 'pool-a'])
    expect(descLegend).toEqual(['pool-a', 'pool-b'])
  })

  it('invokes radar tooltip formatter', () => {
    const { options } = useRadarChartOptions(baseConfig({ chartData: makeRadarChartData() }))
    const tooltip = options.value.tooltip as {
      formatter: (p: { data: { name: string; value: number[] } }) => string
    }
    const html = tooltip.formatter({ data: { name: 'algo-a', value: [10, 20, 5] } })
    expect(html).toContain('algo-a')
    expect(html).toContain('speed')
  })

  it('clamps indicator max to at least 1 for empty spokes', () => {
    const chartData = makeRadarChartData({
      series: [{ xAxis: 'a', values: [0, 0, 0], benchmarkId: '' }],
    })
    const { options } = useRadarChartOptions(baseConfig({ chartData }))
    const radar = options.value.radar as { indicator?: { max: number }[] }
    expect(radar.indicator?.every((i) => i.max >= 1)).toBe(true)
  })
})

describe('useRadarChartOptions — branch edges', () => {
  it('covers x-only without sort and empty first values', () => {
    const chartData = emptyChartData({
      title: 'x',
      statType: 's',
      series: [
        { xAxis: 'A', values: [], benchmarkId: '' },
        { xAxis: 'B', values: [2], benchmarkId: '' },
      ],
    })
    const { options } = useRadarChartOptions(baseConfig({ chartData }))
    const series = options.value.series as { data: { value: number[] }[] }[]
    expect(series[0]!.data[0]!.value).toEqual([0, 2])
  })

  it('covers y-only nulls and x+y null spoke max updates', () => {
    const yOnly = emptyChartData({
      title: 'y',
      statType: 's',
      yAxis: ['a', 'b'],
      series: [{ xAxis: '', values: [null, 3], benchmarkId: '' }],
    })
    const { options } = useRadarChartOptions(baseConfig({ chartData: yOnly }))
    expect((options.value.series as { data: { value: number[] }[] }[])[0]!.data[0]!.value).toEqual([
      0, 3,
    ])

    const xy = makeRadarChartData({
      series: [{ xAxis: 'a', values: [null, 4, 1], benchmarkId: '' }],
    })
    const { options: xyOpts } = useRadarChartOptions(baseConfig({ chartData: xy }))
    expect((xyOpts.value.series as { data: { value: number[] }[] }[])[0]!.data[0]!.value).toEqual([
      0, 4, 1,
    ])
  })

  it('covers z mode without points and missing z totals during sort', () => {
    const data = emptyChartData({
      title: 'z',
      statType: 's',
      yAxis: ['m'],
      zAxis: ['only'],
      series: [{ xAxis: 'x', values: [1], benchmarkId: '' }],
      // no points
    })
    const { options } = useRadarChartOptions(
      baseConfig({ chartData: data, sort: { enabled: true, order: 'asc' } })
    )
    expect((options.value.legend as { data: string[] }).data).toEqual(['only'])
    expect((options.value.series as unknown[]).length).toBe(1)
  })
})

describe('useRadarChartOptions — nullish indicator max', () => {
  it('handles shorter value arrays than y spokes', () => {
    const chartData = makeRadarChartData({
      yAxis: ['a', 'b', 'c', 'd'],
      series: [{ xAxis: 'x', values: [5], benchmarkId: '' }],
    })
    const { options } = useRadarChartOptions(baseConfig({ chartData }))
    const radar = options.value.radar as { indicator: { max: number }[] }
    expect(radar.indicator).toHaveLength(4)
    expect(radar.indicator.every((i) => i.max >= 1)).toBe(true)
  })

  it('z mode with undefined points uses empty loop', () => {
    const data = radarXYZ()
    ;(data as { points?: unknown }).points = undefined
    const { options } = useRadarChartOptions(baseConfig({ chartData: data }))
    expect((options.value.series as unknown[]).length).toBe(2)
  })
})

describe('useRadarChartOptions — z total missing keys', () => {
  it('sorts z values that have no points using zero totals', () => {
    const data = emptyChartData({
      title: 'z',
      statType: 's',
      yAxis: ['m'],
      zAxis: ['present', 'absent'],
      series: [{ xAxis: 'x', values: [1], benchmarkId: '' }],
      points: [{ xAxis: 'x', yAxis: 'm', zAxis: 'present', value: 5 }],
    })
    const { options: asc } = useRadarChartOptions(
      baseConfig({ chartData: data, sort: { enabled: true, order: 'asc' } })
    )
    expect((asc.value.legend as { data: string[] }).data).toEqual(['absent', 'present'])
  })
})
