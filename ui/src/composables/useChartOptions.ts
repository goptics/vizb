import { computed, type Ref } from 'vue'
import type { Axis, ChartData, Sort, ChartType, ScaleType } from '../types'
import type { EChartsOption } from 'echarts'
import { useBarChartOptions } from './charts/useBarChartOptions'
import { useLineChartOptions } from './charts/useLineChartOptions'
import { usePieChartOptions } from './charts/usePieChartOptions'
import { useHeatmapChartOptions } from './charts/useHeatmapChartOptions'
import { useRadarChartOptions } from './charts/useRadarChartOptions'
import { useBar3DChartOptions } from './charts/useBar3DChartOptions'
import { useLine3DChartOptions } from './charts/useLine3DChartOptions'
import { useScatterChartOptions } from './charts/useScatterChartOptions'
import { useScatter3DChartOptions } from './charts/useScatter3DChartOptions'
import type { BaseChartConfig } from './charts/baseChartOptions'
import { is3D } from '../lib/utils'

export function useChartOptions(
  chartData: Ref<ChartData>,
  sort: Ref<Sort>,
  showLabels: Ref<boolean>,
  isDark: Ref<boolean>,
  chartType: Ref<ChartType>,
  scale: Ref<ScaleType>,
  threeDRotate: Ref<boolean>,
  visibleZ: Ref<Record<string, boolean>>,
  threeD: Ref<boolean>,
  threeDVisualMap: Ref<boolean>,
  arrangementTarget: Ref<string>,
  chartAxes: Ref<Axis[] | undefined>
) {
  const config: BaseChartConfig = {
    chartData,
    sort,
    showLabels,
    isDark,
    scale,
    threeDRotate,
    visibleZ,
    threeD,
    threeDVisualMap,
    arrangementTarget,
    chartAxes,
    chartType,
  }

  const barOptions = useBarChartOptions(config)
  const lineOptions = useLineChartOptions(config)
  const pieOptions = usePieChartOptions(config)
  const heatmapOptions = useHeatmapChartOptions(config)
  const radarOptions = useRadarChartOptions(config)
  const bar3DOptions = useBar3DChartOptions(config)
  const line3DOptions = useLine3DChartOptions(config)
  const scatterOptions = useScatterChartOptions(config)
  const scatter3DOptions = useScatter3DChartOptions(config)

  const options = computed<EChartsOption>(() => {
    // When x, y AND z are all present, bar/line render as 3D charts.
    // Pie has no 3D equivalent, so it falls through to the 3-up pie layout.
    const use3D = is3D(
      chartData,
      threeD.value,
      arrangementTarget.value,
      chartAxes.value,
      chartType.value
    )

    switch (chartType.value) {
      case 'bar':
        return use3D ? bar3DOptions.options.value : barOptions.options.value
      case 'line':
        return use3D ? line3DOptions.options.value : lineOptions.options.value
      case 'pie':
        return pieOptions.options.value
      case 'heatmap':
        return heatmapOptions.options.value
      case 'radar':
        return radarOptions.options.value
      case 'scatter':
        return use3D ? scatter3DOptions.options.value : scatterOptions.options.value
      default:
        return use3D ? bar3DOptions.options.value : barOptions.options.value
    }
  })

  return { options }
}
