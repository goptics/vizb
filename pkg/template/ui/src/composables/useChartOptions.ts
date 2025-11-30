import { computed, type Ref } from 'vue'
import type { ChartData } from '../types/benchmark'
import type { Sort } from '../types/benchmark'
import type { ChartType } from '../types/benchmark'
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
  chartType: Ref<ChartType>
) {
  const config: BaseChartConfig = {
    chartData,
    sort,
    showLabels,
    isDark,
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
