import type { Ref } from 'vue'
import type { EChartsOption } from 'echarts'
import type { ChartData, Sort } from '../../types/benchmark'
import { getChartStyling } from './shared/chartConfig'

export interface BaseChartConfig {
  chartData: Ref<ChartData>
  sort: Ref<Sort>
  showLabels: Ref<boolean>
  isDark: Ref<boolean>
}

export const getBaseOptions = (config: BaseChartConfig): Partial<EChartsOption> => {
  const { isDark } = config
  const { textColor } = getChartStyling(isDark.value)
  
  return {
    backgroundColor: "transparent",
    tooltip: {
      trigger: "item",
      formatter: (params: any) => {
        return `${params.marker} <strong>${params.name}</strong><br/>${params.value}`;
      },
    },
    toolbox: {
      show: true,
      feature: {
        saveAsImage: {
          show: true,
          type: "jpeg",
          title: "Save",
          pixelRatio: 2,
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
      left: "center",
      itemWidth: 10,
      itemHeight: 10,
      textStyle: { fontSize: 12, color: textColor },
    },
  };
}
