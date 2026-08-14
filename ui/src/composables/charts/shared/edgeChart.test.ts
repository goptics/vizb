import { describe, expect, it, beforeAll, afterAll } from 'vitest'
import {
  baseConfig,
  emptyChartData,
  installDevicePixelRatio,
  makeSankeyChartData,
} from '@/test-utils'
import type { ChartData } from '@/types'
import { emptyEdgeChartOption, prepareEdgeChart } from './edgeChart'

let restoreDpr: () => void
beforeAll(() => {
  restoreDpr = installDevicePixelRatio()
})
afterAll(() => restoreDpr())

const chartData = () =>
  makeSankeyChartData({
    yAxis: ['B', 'C'],
    series: [
      { xAxis: 'A', values: [0], benchmarkId: '' },
      { xAxis: 'C', values: [0], benchmarkId: '' },
    ],
    points: [
      { xAxis: 'A', yAxis: 'B', zAxis: '', value: 10 },
      { xAxis: 'C', yAxis: 'B', zAxis: '', value: 1 },
    ],
  })

describe('prepareEdgeChart', () => {
  it('builds links and sorts nodes when sort is enabled', () => {
    const { nodes, links } = prepareEdgeChart(
      baseConfig({
        chartData: chartData(),
        sort: { enabled: true, order: 'asc' },
      })
    )
    expect(links).toEqual([
      { source: 'A', target: 'B', value: 10 },
      { source: 'C', target: 'B', value: 1 },
    ])
    expect(nodes.map((n) => n.name)).toEqual(['C', 'A', 'B'])
  })

  it('keeps insertion order when sort is disabled', () => {
    const { nodes } = prepareEdgeChart(
      baseConfig({
        chartData: chartData(),
        sort: { enabled: false, order: 'asc' },
      })
    )
    expect(nodes.map((n) => n.name)).toEqual(['A', 'B', 'C'])
  })

  it('treats missing points as an empty list', () => {
    const { nodes, links } = prepareEdgeChart(
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
    expect(nodes).toEqual([])
    expect(links).toEqual([])
  })

  it('builds an item tooltip that formats edges via the node color map', () => {
    const { nodes, tooltip } = prepareEdgeChart(baseConfig({ chartData: chartData() }))
    expect(tooltip).toMatchObject({ trigger: 'item' })
    const colorByName = new Map(nodes.map((n) => [n.name, n.itemStyle?.color ?? '']))
    const html = (
      tooltip as { formatter: (params: { dataType: string; data: object }) => string }
    ).formatter({
      dataType: 'edge',
      data: { source: 'A', target: 'B', value: 10 },
    })
    expect(html).toContain('<strong>A</strong>')
    expect(html).toContain('<strong>B</strong>')
    expect(html).toContain(`background:${colorByName.get('A')}`)
    expect(html).toContain(`background:${colorByName.get('B')}`)
  })

  it('falls back to #888 for unknown tooltip nodes', () => {
    const { tooltip } = prepareEdgeChart(baseConfig({ chartData: chartData() }))
    const html = (
      tooltip as { formatter: (params: { name: string; value: number }) => string }
    ).formatter({
      name: 'Unknown',
      value: 1,
    })
    expect(html).toContain('<strong>Unknown</strong>')
    expect(html).toContain('background:#888')
  })
})

describe('emptyEdgeChartOption', () => {
  it('hides the legend and forces empty data and links onto the series', () => {
    const options = emptyEdgeChartOption(baseConfig(), {
      type: 'sankey',
      data: [{ name: 'stale' }],
      links: [{ source: 'A', target: 'B', value: 1 }],
    })
    expect(options.legend).toMatchObject({ show: false })
    expect((options.series as Array<Record<string, unknown>>)[0]).toMatchObject({
      type: 'sankey',
      data: [],
      links: [],
    })
  })
})
