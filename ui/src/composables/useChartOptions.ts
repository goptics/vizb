import { computed, type Ref } from 'vue'
import type { ChartData, Sort, ChartType, ScaleType } from '../types'
import type { EChartsOption } from 'echarts'
import { useBarChartOptions } from './charts/useBarChartOptions'
import { useLineChartOptions } from './charts/useLineChartOptions'
import { usePieChartOptions } from './charts/usePieChartOptions'
import type { BaseChartConfig } from './charts/baseChartOptions'

export function useChartOptions(
  chartData: Ref<ChartData>,
  sort: Ref<Sort>,
  showLabels: Ref<boolean>,
  isDark: Ref<boolean>,
  chartType: Ref<ChartType>,
  scale: Ref<ScaleType>
) {
  const config: BaseChartConfig = {
    chartData,
    sort,
    showLabels,
    isDark,
    scale,
  }

  const barOptions = useBarChartOptions(config)
  const lineOptions = useLineChartOptions(config)
  const pieOptions = usePieChartOptions(config)

  const options = computed<EChartsOption>(() => {
    switch (chartType.value) {
      case 'bar':
        return barOptions.options.value
      case 'line':
        return lineOptions.options.value
      case 'pie':
        return pieOptions.options.value
      default:
        return barOptions.options.value
    }
  })

  return { options }
}
