import { describe, expect, it } from 'vitest'
import { buildEdgeGraph, sortEdgeGraphNodes } from './edgeGraph'

const point = (xAxis: string, yAxis: string, value: number, zAxis = '') => ({
  xAxis,
  yAxis,
  zAxis,
  value,
})

describe('edgeGraph', () => {
  it('sums duplicate directional links and keeps reverse links separate', () => {
    const graph = buildEdgeGraph([
      point('A', 'B', 3, 'first'),
      point('A', 'B', 7, 'second'),
      point('B', 'A', 5),
    ])

    expect(graph.links).toEqual([
      { source: 'A', target: 'B', value: 10 },
      { source: 'B', target: 'A', value: 5 },
    ])
  })

  it('assigns stable colors and ignores incomplete edges', () => {
    const first = buildEdgeGraph([point('A', 'B', 4), point('', 'C', 9), point('D', '', 9)])
    const second = buildEdgeGraph([point('A', 'B', 4)])

    expect(first.nodes).toEqual([
      { name: 'A', total: 4, itemStyle: { color: expect.any(String) } },
      { name: 'B', total: 4, itemStyle: { color: expect.any(String) } },
    ])
    expect(first.nodes.map((node) => node.itemStyle?.color)).toEqual(
      second.nodes.map((node) => node.itemStyle?.color)
    )
  })

  it('clamps aggregated link weights to non-negative display values', () => {
    const graph = buildEdgeGraph([point('A', 'B', -8), point('A', 'C', 2)])

    expect(graph.links).toEqual([
      { source: 'A', target: 'B', value: 0 },
      { source: 'A', target: 'C', value: 2 },
    ])
  })

  it('sorts nodes by total and breaks ties by name', () => {
    const { nodes } = buildEdgeGraph([point('C', 'D', 5), point('A', 'B', 5), point('A', 'C', 2)])

    expect(sortEdgeGraphNodes(nodes, 'asc').map((node) => node.name)).toEqual(['B', 'D', 'A', 'C'])
    expect(sortEdgeGraphNodes(nodes, 'desc').map((node) => node.name)).toEqual(['A', 'C', 'B', 'D'])
  })
})
