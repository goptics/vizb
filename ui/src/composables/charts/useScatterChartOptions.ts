import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig } from './baseChartOptions'
import { buildScatterAxes2DValueOptions } from './shared/valueMode'
import { useCategorySeriesChartOptions } from './useCategorySeriesChartOptions'

export function useScatterChartOptions(config: BaseChartConfig) {
  const { chartData } = config
  const grouped = useCategorySeriesChartOptions(config, 'scatter')

  const options = computed<EChartsOption>(() => {
    if (chartData.value.valueTuples?.length) {
      return buildScatterAxes2DValueOptions(config)
    }
    return grouped.options.value
  })

  return { options }
}
