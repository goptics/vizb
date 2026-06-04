import type { Ref } from 'vue'
import type { EChartsOption } from 'echarts'
import type { ChartData, Sort, ScaleType } from '../../types'
import { createTooltipConfig, getChartStyling } from './shared/chartConfig'
import { fontSize } from './shared/common'
import { is3D } from '../../lib/utils'

export interface BaseChartConfig {
  chartData: Ref<ChartData>
  sort: Ref<Sort>
  showLabels: Ref<boolean>
  isDark: Ref<boolean>
  scale: Ref<ScaleType>
  autoRotate: Ref<boolean>
}

export const getBaseOptions = (config: BaseChartConfig): Partial<EChartsOption> => {
  const { isDark } = config
  const { textColor, backgroundColor } = getChartStyling(isDark.value)
  // zrender's getRenderedCanvas only composites WebGL (echarts-gl) layers when
  // saveAsImage's pixelRatio <= the chart dpr; a larger ratio falls back to a
  // 2D-only redraw that drops the 3D content. Cap 3D exports at the device dpr.
  const dpr = window.devicePixelRatio || 1
  const saveAsImagePixelRatio = is3D(config.chartData) ? dpr : 2
  return {
    backgroundColor,
    tooltip: createTooltipConfig(false) as EChartsOption['tooltip'],
    toolbox: {
      show: true,
      feature: {
        saveAsImage: {
          show: true,
          type: 'jpeg',
          title: 'Save',
          pixelRatio: saveAsImagePixelRatio,
          name: config.chartData.value.title,
        },
      },
      iconStyle: {
        borderColor: textColor,
      },
      emphasis: {
        iconStyle: {
          borderColor: textColor,
        },
      },
    },
    legend: {
      show: true,
      left: 'center',
      itemWidth: 10,
      itemHeight: 10,
      textStyle: { fontSize, color: textColor },
    },
    emphasis: {
      focus: 'series',
    },
  }
}
