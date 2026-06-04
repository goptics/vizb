import { computed, type Ref } from 'vue'
import type { ChartData, Sort, ChartType, ScaleType } from '../types'
import type { EChartsOption } from 'echarts'
import { useBarChartOptions } from './charts/useBarChartOptions'
import { useLineChartOptions } from './charts/useLineChartOptions'
import { usePieChartOptions } from './charts/usePieChartOptions'
import { use3DChartOptions } from './charts/use3DChartOptions'
import type { BaseChartConfig } from './charts/baseChartOptions'
import { is3D } from '../lib/utils'

export function useChartOptions(
  chartData: Ref<ChartData>,
  sort: Ref<Sort>,
  showLabels: Ref<boolean>,
  isDark: Ref<boolean>,
  chartType: Ref<ChartType>,
  scale: Ref<ScaleType>,
  autoRotate: Ref<boolean>
) {
  const config: BaseChartConfig = {
    chartData,
    sort,
    showLabels,
    isDark,
    scale,
    autoRotate,
  }

  const barOptions = useBarChartOptions(config)
  const lineOptions = useLineChartOptions(config)
  const pieOptions = usePieChartOptions(config)
  const bar3DOptions = use3DChartOptions(config, 'bar3D')
  const line3DOptions = use3DChartOptions(config, 'line3D')

  const options = computed<EChartsOption>(() => {
    // When x, y AND z are all present, bar/line render as 3D charts.
    // Pie has no 3D equivalent, so it falls through to the 3-up pie layout.
    const threeD = is3D(chartData)

    switch (chartType.value) {
      case 'bar':
        return threeD ? bar3DOptions.options.value : barOptions.options.value
      case 'line':
        return threeD ? line3DOptions.options.value : lineOptions.options.value
      case 'pie':
        return pieOptions.options.value
      default:
        return threeD ? bar3DOptions.options.value : barOptions.options.value
    }
  })

  return { options }
}
