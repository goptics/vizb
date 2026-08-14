import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions } from './baseChartOptions'
import { hasXAxis, hasYAxis } from '@/lib/utils'
import { getChartStyling } from './shared/chartConfig'
import { fontSize } from './shared/common'
import { formatTooltipValue } from './shared/chartConfig'
import { emptyEdgeChartOption, prepareEdgeChart } from './shared/edgeChart'

/** ECharts defaults right: '20%' for label room; we keep a tighter inset. */
const sankeyLayout = {
  left: '4%',
  right: '8%',
  top: '4%',
  bottom: '4%',
} as const

const sankeySeriesDefaults = {
  type: 'sankey' as const,
  ...sankeyLayout,
  emphasis: { focus: 'adjacency' as const },
  lineStyle: { curveness: 0.5, color: 'gradient' as const },
}

export function useSankeyChartOptions(config: BaseChartConfig) {
  const { chartData, showLabels, isDark } = config

  const options = computed<EChartsOption>(() => {
    // Sankey needs source (x) and target (y). Without both, emit an empty series
    // so the canvas stays blank rather than mis-drawing from partial axes.
    if (!hasXAxis(chartData) || !hasYAxis(chartData)) {
      return emptyEdgeChartOption(config, sankeySeriesDefaults)
    }

    const styling = getChartStyling(isDark.value)
    const { nodes, links, tooltip } = prepareEdgeChart(config)
    // Drop the internal total field before handing nodes to ECharts.
    const data = nodes.map(({ name, itemStyle }) => ({ name, itemStyle }))

    return {
      ...getBaseOptions(config),
      legend: { show: false },
      tooltip,
      series: [
        {
          ...sankeySeriesDefaults,
          data,
          links,
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
