import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { hasXAxis, hasYAxis } from '@/lib/utils'
import { getChartStyling, getTooltipTheme, formatTooltipValue } from './shared'
import { fontSize } from './shared/common'
import { buildEdgeGraph, sortEdgeGraphNodes } from './shared/edgeGraph'

/** ECharts-style color circle for tooltip rows (matches other chart formatters). */
function tooltipColorDot(color: string): string {
  return `<span style="display:inline-block;width:10px;height:10px;border-radius:50%;background:${color};margin-right:6px"></span>`
}

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
        formatter: (params: any) => {
          // Link hover: colored source → colored target + weight
          if (params.dataType === 'edge' || params.data?.source != null) {
            const src = String(params.data?.source ?? params.name ?? '')
            const tgt = String(params.data?.target ?? '')
            const val = params.data?.value ?? params.value
            return (
              `${tooltipColorDot(colorFor(src))}<strong>${src}</strong>` +
              ` → ` +
              `${tooltipColorDot(colorFor(tgt))}<strong>${tgt}</strong>` +
              `<br/>${formatTooltipValue(val)}`
            )
          }
          // Node hover: name + total flow through the node
          const name = params.name ?? params.data?.name ?? ''
          const val = params.value ?? params.data?.value
          const valueLine =
            val === undefined || val === null ? '' : `<br/>${formatTooltipValue(val)}`
          const marker = params.marker || tooltipColorDot(colorFor(String(name)))
          return `${marker} <strong>${name}</strong>${valueLine}`
        },
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
