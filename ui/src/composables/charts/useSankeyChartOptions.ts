import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { hasXAxis, hasYAxis } from '@/lib/utils'
import { getChartStyling, getTooltipTheme } from './shared/chartConfig'
import { fontSize } from './shared/common'
import { formatTooltipValue } from './shared/chartConfig'
import { buildEdgeGraph, formatEdgeTooltip, sortEdgeGraphNodes } from './shared/edgeGraph'

/** ECharts defaults right: '20%' for label room; we keep a tighter inset. */
const sankeyLayout = {
  left: '4%',
  right: '8%',
  top: '4%',
  bottom: '4%',
} as const

function minimalSankeyOption(config: BaseChartConfig): EChartsOption {
  const base = getBaseOptions(config)
  return {
    ...base,
    legend: { show: false },
    series: [
      {
        type: 'sankey',
        ...sankeyLayout,
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
    let { nodes, links } = buildEdgeGraph(chartData.value.points ?? [])

    if (sort.value.enabled) {
      nodes = sortEdgeGraphNodes(nodes, sort.value.order)
    }

    // Drop the internal total field before handing nodes to ECharts.
    const data = nodes.map(({ name, itemStyle }) => ({ name, itemStyle }))
    // buildSankeyGraph always assigns a color, so lookups below never miss.
    const nodeColor = new Map<string, string>(
      nodes.map((n) => [n.name, n.itemStyle!.color!] as const)
    )
    const colorFor = (name: string): string => nodeColor.get(name) ?? '#888'

    return {
      ...base,
      legend: { show: false },
      tooltip: {
        trigger: 'item',
        ...getTooltipTheme(isDark.value),
        formatter: (params: any) => formatEdgeTooltip(params, colorFor),
      } as EChartsOption['tooltip'],
      series: [
        {
          type: 'sankey',
          ...sankeyLayout,
          data,
          links,
          emphasis: { focus: 'adjacency' },
          lineStyle: { curveness: 0.5, color: 'gradient' },
          label: {
            show: true,
            color: styling.textColor,
            fontSize,
          },
          edgeLabel: {
            show: showLabels.value,
            formatter: (params: { data?: { value?: number }; value?: number }) =>
              formatTooltipValue(params.data?.value ?? params.value),
            color: styling.textColor,
            fontSize,
          },
        },
      ],
    } as EChartsOption
  })

  return { options }
}
