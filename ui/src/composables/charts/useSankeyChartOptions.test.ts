import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import {
  baseConfig,
  emptyChartData,
  makeGroupedChartData,
  makeSankeyChartData,
  installDevicePixelRatio,
} from '@/test-utils'
import type { ChartData } from '@/types'
import { useSankeyChartOptions } from './useSankeyChartOptions'

let restoreDpr: () => void
beforeAll(() => {
  restoreDpr = installDevicePixelRatio()
})
afterAll(() => restoreDpr())

/** Sankey edge list as chart points (source=x, target=y) with default A→B, B→C. */
const withPoints = (
  points: { xAxis: string; yAxis: string; zAxis?: string; value: number }[] = [
    { xAxis: 'A', yAxis: 'B', value: 10 },
    { xAxis: 'B', yAxis: 'C', value: 5 },
  ]
) =>
  makeSankeyChartData({
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
  })

type SankeySeries = {
  type: string
  data: { name: string }[]
  links: { source: string; target: string; value: number }[]
  label?: { show?: boolean }
  emphasis?: { focus?: string }
  lineStyle?: { curveness?: number }
}

const firstSeries = (options: { series?: unknown }): SankeySeries =>
  (options.series as SankeySeries[])[0]!

describe('useSankeyChartOptions', () => {
  it('emits sankey series type with nodes and links from points', () => {
    const { options } = useSankeyChartOptions(baseConfig({ chartData: withPoints() }))
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
        chartData: withPoints([
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
        chartData: withPoints([
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
        chartData: withPoints([
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
    expect(options.value.legend).toMatchObject({ show: false })
    expect(series.lineStyle).toMatchObject({ curveness: 0.5, color: 'gradient' })
    expect(series.emphasis).toMatchObject({ focus: 'adjacency' })
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
    expect(series.lineStyle?.curveness).toBe(0.5)
  })

  it('emits empty sankey series when both axes exist but points is empty', () => {
    const { options } = useSankeyChartOptions(
      baseConfig({
        chartData: emptyChartData({
          series: [{ xAxis: 'A', values: [1], benchmarkId: '' }],
          yAxis: ['B'],
        }),
      })
    )
    const series = firstSeries(options.value)
    expect(series.data).toEqual([])
    expect(series.links).toEqual([])
  })

  it('treats missing points as an empty list when both axes exist', () => {
    const { options } = useSankeyChartOptions(
      baseConfig({
        chartData: {
          ...emptyChartData({
            series: [{ xAxis: 'A', values: [1], benchmarkId: '' }],
            yAxis: ['B'],
          }),
          points: undefined,
        } as unknown as ChartData,
      })
    )
    const series = firstSeries(options.value)
    expect(series.data).toEqual([])
    expect(series.links).toEqual([])
  })

  it('shows node labels when showLabels is on', () => {
    const { options } = useSankeyChartOptions(
      baseConfig({ chartData: withPoints(), showLabels: true })
    )
    expect(firstSeries(options.value).label?.show).toBe(true)
  })

  it('hides node labels when showLabels is off', () => {
    const { options } = useSankeyChartOptions(
      baseConfig({ chartData: withPoints(), showLabels: false })
    )
    expect(firstSeries(options.value).label?.show).toBe(false)
  })

  it('sorts nodes by total flow when sort is enabled (asc)', () => {
    // A→B:10, C→B:1 → totals: A=10, B=11, C=1
    const { options } = useSankeyChartOptions(
      baseConfig({
        chartData: withPoints([
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
        chartData: withPoints([
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
        chartData: withPoints(points),
        sort: { enabled: false, order: 'asc' },
      })
    )
    // Insertion order from first appearance in points (Z, Y, A, B)
    expect(firstSeries(options.value).data.map((d) => d.name)).toEqual(['Z', 'Y', 'A', 'B'])
  })

  it('hides the legend', () => {
    const { options } = useSankeyChartOptions(baseConfig({ chartData: withPoints() }))
    expect(options.value.legend).toMatchObject({ show: false })
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
        chartData: withPoints([
          { xAxis: '', yAxis: 'B', value: 5 },
          { xAxis: 'A', yAxis: '', value: 5 },
          { xAxis: 'A', yAxis: 'B', value: 3 },
        ]),
      })
    )
    expect(firstSeries(options.value).links).toEqual([{ source: 'A', target: 'B', value: 3 }])
  })

  it('keeps positive link values and assigns each node a color', () => {
    const { options } = useSankeyChartOptions(
      baseConfig({ chartData: withPoints([{ xAxis: 'A', yAxis: 'B', value: 7 }]) })
    )
    const series = firstSeries(options.value)
    expect(series.links).toEqual([{ source: 'A', target: 'B', value: 7 }])
    // Nodes carry a computed color and never leak the internal total field.
    expect(series.data).toEqual([
      { name: 'A', itemStyle: { color: expect.any(String) } },
      { name: 'B', itemStyle: { color: expect.any(String) } },
    ])
  })

  it('breaks sort ties alphabetically when totals are equal', () => {
    const { options } = useSankeyChartOptions(
      baseConfig({
        chartData: withPoints([
          { xAxis: 'C', yAxis: 'D', value: 5 },
          { xAxis: 'A', yAxis: 'B', value: 5 },
        ]),
        sort: { enabled: true, order: 'asc' },
      })
    )
    // Totals: C=5, D=5, A=5, B=5 → alphabetical by name.
    expect(firstSeries(options.value).data.map((d) => d.name)).toEqual(['A', 'B', 'C', 'D'])
  })

  it('applies dark-mode tooltip and label styling', () => {
    const { options } = useSankeyChartOptions(baseConfig({ chartData: withPoints(), isDark: true }))
    const tooltip = options.value.tooltip as { backgroundColor?: string; borderColor?: string }
    expect(tooltip.backgroundColor).toBe('#1f2937')
    expect(tooltip.borderColor).toBe('#4b5563')
    const series = firstSeries(options.value)
    expect(series.label?.show).toBe(false)
  })

  describe('tooltip formatter', () => {
    const formatter = (options: { tooltip?: unknown }) =>
      (options.tooltip as { formatter: (params: any) => string }).formatter

    it('formats edge hover via dataType', () => {
      const { options } = useSankeyChartOptions(baseConfig({ chartData: withPoints() }))
      const html = formatter(options.value)({
        marker: '*',
        dataType: 'edge',
        data: { source: 'A', target: 'B', value: 10 },
        name: 'A → B',
        value: 10,
      })
      expect(html).toContain('<strong>A → B</strong>')
      expect(html).toContain('10')
    })

    it('formats edge hover via data.source fallback', () => {
      const { options } = useSankeyChartOptions(baseConfig({ chartData: withPoints() }))
      const html = formatter(options.value)({
        marker: '*',
        data: { source: 'A', target: 'B', value: 5 },
      })
      expect(html).toContain('<strong>A → B</strong>')
      expect(html).toContain('5')
    })

    it('falls back to params.name for the source when data.source is missing', () => {
      const { options } = useSankeyChartOptions(baseConfig({ chartData: withPoints() }))
      const html = formatter(options.value)({
        marker: '*',
        dataType: 'edge',
        data: { target: 'B', value: 5 },
        name: 'A',
      })
      expect(html).toBe('* <strong>A → B</strong><br/>5')
    })

    it('formats node hover with a value line', () => {
      const { options } = useSankeyChartOptions(baseConfig({ chartData: withPoints() }))
      const html = formatter(options.value)({ marker: '*', name: 'A', value: 10 })
      expect(html).toContain('<strong>A</strong>')
      expect(html).toContain('<br/>')
      expect(html).toContain('10')
    })

    it('omits the value line when node value is null or undefined', () => {
      const { options } = useSankeyChartOptions(baseConfig({ chartData: withPoints() }))
      const withNull = formatter(options.value)({ marker: '*', name: 'A', value: null })
      expect(withNull).toBe('* <strong>A</strong>')
      const withUndefined = formatter(options.value)({ marker: '*', name: 'A', value: undefined })
      expect(withUndefined).toBe('* <strong>A</strong>')
    })

    it('falls back to an empty target string when data.target is missing', () => {
      const { options } = useSankeyChartOptions(baseConfig({ chartData: withPoints() }))
      const html = formatter(options.value)({
        marker: '*',
        dataType: 'edge',
        data: { source: 'A', value: 5 },
      })
      expect(html).toBe('* <strong>A → </strong><br/>5')
    })

    it('falls back to params.value when data.value is missing on an edge', () => {
      const { options } = useSankeyChartOptions(baseConfig({ chartData: withPoints() }))
      const html = formatter(options.value)({
        marker: '*',
        dataType: 'edge',
        data: { source: 'A', target: 'B' },
        value: 9,
      })
      expect(html).toBe('* <strong>A → B</strong><br/>9')
    })

    it('falls back to an empty name when params.name and data.name are missing', () => {
      const { options } = useSankeyChartOptions(baseConfig({ chartData: withPoints() }))
      const html = formatter(options.value)({ marker: '*', value: 3 })
      expect(html).toBe('* <strong></strong><br/>3')
    })

    it('falls back to data.name and data.value for nodes', () => {
      const { options } = useSankeyChartOptions(baseConfig({ chartData: withPoints() }))
      const html = formatter(options.value)({
        marker: '*',
        data: { name: 'B', value: 7 },
      })
      expect(html).toContain('<strong>B</strong>')
      expect(html).toContain('7')
    })
  })
})
