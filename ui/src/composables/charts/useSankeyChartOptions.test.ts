import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import {
  baseConfig,
  emptyChartData,
  makeGroupedChartData,
  installDevicePixelRatio,
} from '@/test-utils'
import { useSankeyChartOptions } from './useSankeyChartOptions'

let restoreDpr: () => void
beforeAll(() => {
  restoreDpr = installDevicePixelRatio()
})
afterAll(() => restoreDpr())

/** Minimal x+y sankey edge list as chart points (source=x, target=y). */
const makeSankeyChartData = (
  points: { xAxis: string; yAxis: string; zAxis?: string; value: number }[] = [
    { xAxis: 'A', yAxis: 'B', value: 10 },
    { xAxis: 'B', yAxis: 'C', value: 5 },
  ]
) =>
  emptyChartData({
    title: 'flow',
    statType: 'sum',
    yAxis: [...new Set(points.map((p) => p.yAxis))],
    series: [...new Set(points.map((p) => p.xAxis))].map((x) => ({
      xAxis: x,
      values: [0],
      benchmarkId: '',
    })),
    points: points.map((p) => ({
      xAxis: p.xAxis,
      yAxis: p.yAxis,
      zAxis: p.zAxis ?? '',
      value: p.value,
    })),
    axisLabels: { x: 'source', y: 'target' },
  })

type SankeySeries = {
  type: string
  data: { name: string }[]
  links: { source: string; target: string; value: number }[]
  label?: { show?: boolean }
  emphasis?: { focus?: string }
  lineStyle?: { curveness?: number }
  left?: string
  right?: string
  top?: string
  bottom?: string
}

const firstSeries = (options: { series?: unknown }): SankeySeries =>
  (options.series as SankeySeries[])[0]!

describe('useSankeyChartOptions', () => {
  it('emits sankey series type with nodes and links from points', () => {
    const { options } = useSankeyChartOptions(baseConfig({ chartData: makeSankeyChartData() }))
    const series = firstSeries(options.value)
    expect(series.type).toBe('sankey')
    expect(series.data.map((d) => d.name).sort()).toEqual(['A', 'B', 'C'])
    expect(series.links).toEqual(
      expect.arrayContaining([
        { source: 'A', target: 'B', value: 10 },
        { source: 'B', target: 'C', value: 5 },
      ])
    )
    expect(series.links).toHaveLength(2)
    expect(series.emphasis?.focus).toBe('adjacency')
    expect(series.lineStyle?.curveness).toBe(0.5)
  })

  it('sums duplicate (source, target) pairs', () => {
    const { options } = useSankeyChartOptions(
      baseConfig({
        chartData: makeSankeyChartData([
          { xAxis: 'A', yAxis: 'B', value: 3 },
          { xAxis: 'A', yAxis: 'B', value: 7 },
          { xAxis: 'A', yAxis: 'C', value: 2 },
        ]),
      })
    )
    const links = firstSeries(options.value).links
    expect(links).toEqual(
      expect.arrayContaining([
        { source: 'A', target: 'B', value: 10 },
        { source: 'A', target: 'C', value: 2 },
      ])
    )
    expect(links).toHaveLength(2)
  })

  it('ignores z when summing the same (source, target) pair', () => {
    const { options } = useSankeyChartOptions(
      baseConfig({
        chartData: makeSankeyChartData([
          { xAxis: 'A', yAxis: 'B', zAxis: 'z1', value: 4 },
          { xAxis: 'A', yAxis: 'B', zAxis: 'z2', value: 6 },
        ]),
      })
    )
    const series = firstSeries(options.value)
    expect(series.links).toEqual([{ source: 'A', target: 'B', value: 10 }])
    // Only source/target nodes — z never becomes a node
    expect(series.data.map((d) => d.name).sort()).toEqual(['A', 'B'])
  })

  it('clamps negative link values to 0 for display', () => {
    const { options } = useSankeyChartOptions(
      baseConfig({
        chartData: makeSankeyChartData([
          { xAxis: 'A', yAxis: 'B', value: -5 },
          { xAxis: 'A', yAxis: 'C', value: 3 },
        ]),
      })
    )
    const links = firstSeries(options.value).links
    expect(links.find((l) => l.target === 'B')?.value).toBe(0)
    expect(links.find((l) => l.target === 'C')?.value).toBe(3)
  })

  it('returns empty sankey series when y axis is missing', () => {
    const { options } = useSankeyChartOptions(
      baseConfig({
        chartData: emptyChartData({
          series: [{ xAxis: 'A', values: [1], benchmarkId: '' }],
          yAxis: [],
          points: [{ xAxis: 'A', yAxis: 'B', zAxis: '', value: 1 }],
        }),
      })
    )
    const series = firstSeries(options.value)
    expect(series.type).toBe('sankey')
    expect(series.data).toEqual([])
    expect(series.links).toEqual([])
  })

  it('returns empty sankey series when x axis is missing', () => {
    const { options } = useSankeyChartOptions(
      baseConfig({
        chartData: emptyChartData({
          series: [{ xAxis: '', values: [1], benchmarkId: '' }],
          yAxis: ['B'],
          points: [{ xAxis: 'A', yAxis: 'B', zAxis: '', value: 1 }],
        }),
      })
    )
    const series = firstSeries(options.value)
    expect(series.data).toEqual([])
    expect(series.links).toEqual([])
  })

  it('shows node labels when showLabels is on', () => {
    const { options } = useSankeyChartOptions(
      baseConfig({ chartData: makeSankeyChartData(), showLabels: true })
    )
    expect(firstSeries(options.value).label?.show).toBe(true)
  })

  it('hides node labels when showLabels is off', () => {
    const { options } = useSankeyChartOptions(
      baseConfig({ chartData: makeSankeyChartData(), showLabels: false })
    )
    expect(firstSeries(options.value).label?.show).toBe(false)
  })

  it('sorts nodes by total flow when sort is enabled (asc)', () => {
    // A→B:10, C→B:1 → totals: A=10, B=11, C=1
    const { options } = useSankeyChartOptions(
      baseConfig({
        chartData: makeSankeyChartData([
          { xAxis: 'A', yAxis: 'B', value: 10 },
          { xAxis: 'C', yAxis: 'B', value: 1 },
        ]),
        sort: { enabled: true, order: 'asc' },
      })
    )
    expect(firstSeries(options.value).data.map((d) => d.name)).toEqual(['C', 'A', 'B'])
  })

  it('sorts nodes by total flow when sort is enabled (desc)', () => {
    const { options } = useSankeyChartOptions(
      baseConfig({
        chartData: makeSankeyChartData([
          { xAxis: 'A', yAxis: 'B', value: 10 },
          { xAxis: 'C', yAxis: 'B', value: 1 },
        ]),
        sort: { enabled: true, order: 'desc' },
      })
    )
    expect(firstSeries(options.value).data.map((d) => d.name)).toEqual(['B', 'A', 'C'])
  })

  it('leaves node order unsorted when sort is disabled', () => {
    const points = [
      { xAxis: 'Z', yAxis: 'Y', value: 1 },
      { xAxis: 'A', yAxis: 'B', value: 100 },
    ]
    const { options } = useSankeyChartOptions(
      baseConfig({
        chartData: makeSankeyChartData(points),
        sort: { enabled: false, order: 'asc' },
      })
    )
    // Insertion order from first appearance in points (Z, Y, A, B)
    expect(firstSeries(options.value).data.map((d) => d.name)).toEqual(['Z', 'Y', 'A', 'B'])
  })

  it('hides the legend', () => {
    const { options } = useSankeyChartOptions(baseConfig({ chartData: makeSankeyChartData() }))
    expect(options.value.legend).toMatchObject({ show: false })
  })

  it('uses a tighter series inset than ECharts sankey defaults (right 20%)', () => {
    const { options } = useSankeyChartOptions(baseConfig({ chartData: makeSankeyChartData() }))
    const series = firstSeries(options.value)
    expect(series).toMatchObject({ left: '4%', right: '8%', top: '4%', bottom: '4%' })
  })

  it('works with standard grouped chart fixtures that carry points', () => {
    const chartData = makeGroupedChartData({
      points: [
        { xAxis: 'West', yAxis: 'Hardware', zAxis: '', value: 10 },
        { xAxis: 'East', yAxis: 'Software', zAxis: '', value: 40 },
      ],
    })
    const { options } = useSankeyChartOptions(baseConfig({ chartData }))
    const series = firstSeries(options.value)
    expect(series.type).toBe('sankey')
    expect(series.links).toHaveLength(2)
  })

  it('skips points with empty source or target', () => {
    const { options } = useSankeyChartOptions(
      baseConfig({
        chartData: makeSankeyChartData([
          { xAxis: '', yAxis: 'B', value: 5 },
          { xAxis: 'A', yAxis: '', value: 5 },
          { xAxis: 'A', yAxis: 'B', value: 3 },
        ]),
      })
    )
    expect(firstSeries(options.value).links).toEqual([{ source: 'A', target: 'B', value: 3 }])
  })

  it('edge tooltip shows color circles for both source and target', () => {
    const { options } = useSankeyChartOptions(baseConfig({ chartData: makeSankeyChartData() }))
    const series = firstSeries(options.value)
    const colorByName = new Map(
      series.data.map((d) => {
        const node = d as { name: string; itemStyle?: { color?: string } }
        return [node.name, node.itemStyle?.color ?? ''] as const
      })
    )
    const formatter = (
      options.value.tooltip as {
        formatter: (p: {
          dataType?: string
          data?: { source?: string; target?: string; value?: number }
          value?: number
        }) => string
      }
    ).formatter

    const html = formatter({
      dataType: 'edge',
      data: { source: 'A', target: 'B', value: 10 },
    })

    expect(html).toContain('<strong>A</strong>')
    expect(html).toContain('<strong>B</strong>')
    expect(html).toContain('→')
    // One color circle per endpoint, using the same node colors as the series
    expect(html).toContain(`background:${colorByName.get('A')}`)
    expect(html).toContain(`background:${colorByName.get('B')}`)
    expect(html.match(/border-radius:50%/g)?.length).toBe(2)
  })
})
