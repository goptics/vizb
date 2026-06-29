import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig } from './baseChartOptions'
import { buildValueAxes2DOptions } from './shared/valueMode'
import { buildMixedAxes2DOptions } from './shared/mixedMode'
import { useCategorySeriesChartOptions } from './useCategorySeriesChartOptions'

export function useLineChartOptions(config: BaseChartConfig) {
  const { chartData } = config
  const grouped = useCategorySeriesChartOptions(config, 'line')

  const options = computed<EChartsOption>(() => {
    if (chartData.value.mixedTuples?.length) {
      return buildMixedAxes2DOptions(config, 'line')
    }
    if (chartData.value.valueTuples?.length) {
      return buildValueAxes2DOptions(config, 'line')
    }
    return grouped.options.value
  })

  return { options }
}
