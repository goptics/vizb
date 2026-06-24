import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig } from './baseChartOptions'
import { buildValueAxes2DOptions } from './shared/valueMode'
import { useCategorySeriesChartOptions } from './useCategorySeriesChartOptions'

export function useScatterChartOptions(config: BaseChartConfig) {
  const { chartData } = config
  const grouped = useCategorySeriesChartOptions(config, 'scatter')

  const options = computed<EChartsOption>(() => {
    if (chartData.value.valueTuples?.length) {
      return buildValueAxes2DOptions(config, 'scatter')
    }
    return grouped.options.value
  })

  return { options }
}
