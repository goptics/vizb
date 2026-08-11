import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { hasXAxis, hasYAxis } from '@/lib/utils'
import { getChartStyling, getTooltipTheme } from './shared'
import { fontSize } from './shared/common'
import { buildEdgeGraph, formatEdgeTooltip, sortEdgeGraphNodes } from './shared/edgeGraph'

const chordLayout = {
  left: '4%',
  right: '4%',
  top: '8%',
  bottom: '8%',
  center: ['50%', '52%'],
  radius: ['58%', '72%'],
} as const

function minimalChordOption(config: BaseChartConfig): EChartsOption {
  return {
    ...getBaseOptions(config),
    legend: { show: false },
    series: [
      {
        type: 'chord',
        ...chordLayout,
        data: [],
        links: [],
        emphasis: { focus: 'adjacency' },
        lineStyle: { color: 'gradient', opacity: 0.35 },
      },
    ],
  } as unknown as EChartsOption
}

/** Build the ECharts 6 Chord option from Vizb's shared x→y edge points. */
export function useChordChartOptions(config: BaseChartConfig) {
  const { chartData, sort, showLabels, isDark } = config

  const options = computed<EChartsOption>(() => {
    if (!hasXAxis(chartData) || !hasYAxis(chartData)) {
      return minimalChordOption(config)
    }

    const styling = getChartStyling(isDark.value)
    const { nodes: rawNodes, links } = buildEdgeGraph(chartData.value.points ?? [])
    const nodes = sort.value.enabled ? sortEdgeGraphNodes(rawNodes, sort.value.order) : rawNodes
    const data = nodes.map(({ name, total, itemStyle }) => ({
      name,
      value: Math.max(0, total),
      itemStyle,
    }))
    // buildEdgeGraph always assigns a color, so lookups below never miss the map key's value.
    const nodeColor = new Map<string, string>(
      nodes.map((n) => [n.name, n.itemStyle!.color!] as const)
    )
    const colorFor = (name: string): string => nodeColor.get(name) ?? '#888'

    return {
      ...getBaseOptions(config),
      legend: { show: false },
      tooltip: {
        trigger: 'item',
        ...getTooltipTheme(isDark.value),
        formatter: (params: any) => formatEdgeTooltip(params, colorFor),
      } as EChartsOption['tooltip'],
      series: [
        {
          type: 'chord',
          ...chordLayout,
          data,
          links,
          emphasis: { focus: 'adjacency' },
          lineStyle: { color: 'gradient', opacity: 0.35, curveness: 0.15 },
          label: {
            show: showLabels.value,
            color: styling.textColor,
            fontSize,
          },
        },
      ],
    } as unknown as EChartsOption
  })

  return { options }
}
