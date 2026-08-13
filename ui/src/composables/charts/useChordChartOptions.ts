import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { hasXAxis, hasYAxis } from '@/lib/utils'
import { getChartStyling, getTooltipTheme } from './shared'
import { fontSize } from './shared/common'
import { buildEdgeGraph, formatEdgeTooltip, sortEdgeGraphNodes } from './shared/edgeGraph'

// Thin outer ring (demo geometry). top = legend band + gap; bottom = same gap.
const chordSeriesDefaults = {
  padAngle: 1,
  top: 52,
  bottom: 24,
  left: '8%',
  right: '8%',
  center: ['50%', '50%'] as [string, string],
  radius: ['70%', '80%'] as [string, string],
  // Override ECharts' non-zero default; optional rounding is a future flag.
  itemStyle: { borderRadius: 0 },
  lineStyle: {
    opacity: 0.3,
    color: 'gradient' as const,
  },
  emphasis: { focus: 'self' as const },
} as const

function minimalChordOption(config: BaseChartConfig): EChartsOption {
  return {
    ...getBaseOptions(config),
    legend: { show: false },
    series: [
      {
        type: 'chord',
        ...chordSeriesDefaults,
        data: [],
        links: [],
      },
    ],
  } as unknown as EChartsOption
}

/** Build the ECharts 6 Chord option from Vizb's shared x→y edge points. */
export function useChordChartOptions(config: BaseChartConfig) {
  const { chartData, sort, showLabels, isDark, visibleZ } = config

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
    // ChartCard writes legend toggles into visibleZ; persist them so ChartChord's
    // notMerge remount on resize does not re-show hidden nodes.
    const selected = visibleZ?.value ?? {}

    return {
      ...getBaseOptions(config),
      legend: {
        show: nodes.length > 0,
        type: 'scroll',
        left: 'center',
        top: 8,
        itemWidth: 10,
        itemHeight: 10,
        data: nodes.map((n) => n.name),
        selected,
        textStyle: { fontSize, color: styling.textColor },
      },
      tooltip: {
        trigger: 'item',
        ...getTooltipTheme(isDark.value),
        formatter: (params: any) => formatEdgeTooltip(params, colorFor),
      } as EChartsOption['tooltip'],
      series: [
        {
          type: 'chord',
          ...chordSeriesDefaults,
          data,
          links,
          label: {
            show: showLabels.value,
            position: 'inside',
            color: '#fff',
            fontWeight: 'bold',
            fontSize,
          },
        },
      ],
    } as unknown as EChartsOption
  })

  return { options }
}
