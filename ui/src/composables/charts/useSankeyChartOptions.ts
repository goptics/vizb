import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { getNextColorFor, hasXAxis, hasYAxis } from '@/lib/utils'
import { getChartStyling, getTooltipTheme, formatTooltipValue } from './shared'
import { fontSize } from './shared/common'
import type { Point3D, SortOrder } from '@/types'

type SankeyLink = { source: string; target: string; value: number }
type SankeyNode = { name: string; itemStyle?: { color?: string }; total: number }

/** Aggregate raw points into unique nodes + summed (source, target) links. z is ignored. */
function buildSankeyGraph(points: Point3D[]): {
  nodes: SankeyNode[]
  links: SankeyLink[]
} {
  const linkTotals = new Map<string, { source: string; target: string; value: number }>()
  const nodeTotals = new Map<string, number>()

  for (const p of points) {
    const source = p.xAxis
    const target = p.yAxis
    if (!source || !target) continue

    const key = `${source}\0${target}`
    const prev = linkTotals.get(key)
    const add = p.value
    if (prev) {
      prev.value += add
    } else {
      linkTotals.set(key, { source, target, value: add })
    }

    // Total flow through a node = sum of link weights touching it (count each
    // link once per endpoint). Self-loops count twice so degree is consistent.
    nodeTotals.set(source, (nodeTotals.get(source) ?? 0) + add)
    nodeTotals.set(target, (nodeTotals.get(target) ?? 0) + add)
  }

  const nodes: SankeyNode[] = Array.from(nodeTotals.entries()).map(([name, total]) => ({
    name,
    total,
    itemStyle: { color: getNextColorFor(name) },
  }))

  const links: SankeyLink[] = Array.from(linkTotals.values()).map((l) => ({
    source: l.source,
    target: l.target,
    value: Math.max(0, l.value),
  }))

  return { nodes, links }
}

function sortNodes(nodes: SankeyNode[], order: SortOrder): SankeyNode[] {
  const mult = order === 'asc' ? 1 : -1
  return [...nodes].sort((a, b) => {
    if (a.total !== b.total) return mult * (a.total - b.total)
    return a.name.localeCompare(b.name)
  })
}

function minimalSankeyOption(config: BaseChartConfig): EChartsOption {
  const base = getBaseOptions(config)
  return {
    ...base,
    legend: { show: false },
    series: [
      {
        type: 'sankey',
        data: [],
        links: [],
        emphasis: { focus: 'adjacency' },
        lineStyle: { curveness: 0.5, color: 'gradient' },
      },
    ],
  } as EChartsOption
}

export function useSankeyChartOptions(config: BaseChartConfig) {
  const { chartData, sort, showLabels, isDark } = config

  const options = computed<EChartsOption>(() => {
    // Sankey needs source (x) and target (y). Without both, emit an empty series
    // so the canvas stays blank rather than mis-drawing from partial axes.
    if (!hasXAxis(chartData) || !hasYAxis(chartData)) {
      return minimalSankeyOption(config)
    }

    const styling = getChartStyling(isDark.value)
    const base = getBaseOptions(config)
    let { nodes, links } = buildSankeyGraph(chartData.value.points ?? [])

    if (sort.value.enabled) {
      nodes = sortNodes(nodes, sort.value.order)
    }

    // Drop the internal total field before handing nodes to ECharts.
    const data = nodes.map(({ name, itemStyle }) => ({ name, itemStyle }))

    return {
      ...base,
      legend: { show: false },
      tooltip: {
        trigger: 'item',
        ...getTooltipTheme(isDark.value),
        formatter: (params: any) => {
          // Link hover: source → target + weight
          if (params.dataType === 'edge' || params.data?.source != null) {
            const src = params.data?.source ?? params.name
            const tgt = params.data?.target ?? ''
            const val = params.data?.value ?? params.value
            return `${params.marker} <strong>${src} → ${tgt}</strong><br/>${formatTooltipValue(val)}`
          }
          // Node hover: name + total flow through the node
          const name = params.name ?? params.data?.name ?? ''
          const val = params.value ?? params.data?.value
          const valueLine =
            val === undefined || val === null ? '' : `<br/>${formatTooltipValue(val)}`
          return `${params.marker} <strong>${name}</strong>${valueLine}`
        },
      } as EChartsOption['tooltip'],
      series: [
        {
          type: 'sankey',
          data,
          links,
          emphasis: { focus: 'adjacency' },
          lineStyle: { curveness: 0.5, color: 'gradient' },
          label: {
            show: showLabels.value,
            color: styling.textColor,
            fontSize,
          },
        },
      ],
    } as EChartsOption
  })

  return { options }
}
