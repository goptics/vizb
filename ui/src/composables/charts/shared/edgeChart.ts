import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from '../baseChartOptions'
import { getTooltipTheme } from './chartConfig'
import {
  buildEdgeGraph,
  formatEdgeTooltip,
  sortEdgeGraphNodes,
  type EdgeGraphLink,
  type EdgeGraphNode,
} from './edgeGraph'

export type PreparedEdgeChart = {
  nodes: EdgeGraphNode[]
  links: EdgeGraphLink[]
  tooltip: EChartsOption['tooltip']
}

/** Graph + item tooltip shared by Sankey and Chord. Series and legend stay with the caller. */
export function prepareEdgeChart(config: BaseChartConfig): PreparedEdgeChart {
  const { nodes, links } = buildEdgeGraph(config.chartData.value.points ?? [])
  const sorted = config.sort.value.enabled
    ? sortEdgeGraphNodes(nodes, config.sort.value.order)
    : nodes
  // buildEdgeGraph always assigns a color, so lookups below never miss.
  const nodeColor = new Map<string, string>(
    sorted.map((n) => [n.name, n.itemStyle!.color!] as const)
  )
  const colorFor = (name: string): string => nodeColor.get(name) ?? '#888'
  return {
    nodes: sorted,
    links,
    tooltip: {
      trigger: 'item',
      ...getTooltipTheme(config.isDark.value),
      formatter: (params: any) => formatEdgeTooltip(params, colorFor),
    } as EChartsOption['tooltip'],
  }
}

/** Blank graph series when x or y is missing. Caller supplies type-specific defaults. */
export function emptyEdgeChartOption(
  config: BaseChartConfig,
  series: Record<string, unknown>
): EChartsOption {
  return {
    ...getBaseOptions(config),
    legend: { show: false },
    series: [{ ...series, data: [], links: [] }],
  } as EChartsOption
}
