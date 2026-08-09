import { getNextColorFor } from '@/lib/utils'
import type { Point3D, SortOrder } from '@/types'

export type EdgeGraphLink = { source: string; target: string; value: number }
export type EdgeGraphNode = { name: string; itemStyle?: { color?: string }; total: number }

/** Aggregate edge points into unique nodes and directional links. z is ignored. */
export function buildEdgeGraph(points: Point3D[]): {
  nodes: EdgeGraphNode[]
  links: EdgeGraphLink[]
} {
  const linkTotals = new Map<string, EdgeGraphLink>()
  const nodeTotals = new Map<string, number>()

  for (const point of points) {
    const source = point.xAxis
    const target = point.yAxis
    if (!source || !target) continue

    const key = `${source}\0${target}`
    const link = linkTotals.get(key)
    if (link) {
      link.value += point.value
    } else {
      linkTotals.set(key, { source, target, value: point.value })
    }

    // Total flow through a node = sum of link weights touching it (count each
    // link once per endpoint). Self-loops count twice so degree is consistent.
    nodeTotals.set(source, (nodeTotals.get(source) ?? 0) + point.value)
    nodeTotals.set(target, (nodeTotals.get(target) ?? 0) + point.value)
  }

  const nodes = Array.from(nodeTotals.entries()).map(([name, total]) => ({
    name,
    total,
    itemStyle: { color: getNextColorFor(name) },
  }))
  const links = Array.from(linkTotals.values()).map((link) => ({
    ...link,
    value: Math.max(0, link.value),
  }))

  return { nodes, links }
}

export function sortEdgeGraphNodes(nodes: EdgeGraphNode[], order: SortOrder): EdgeGraphNode[] {
  const multiplier = order === 'asc' ? 1 : -1
  return [...nodes].sort((a, b) => {
    if (a.total !== b.total) return multiplier * (a.total - b.total)
    return a.name.localeCompare(b.name)
  })
}
