import type { Ref } from 'vue'
import type { EChartsOption } from 'echarts'
import type { ChartData, SortOrder } from '../../types/benchmark'

export interface BaseChartConfig {
  chartData: Ref<ChartData>
  sortOrder: Ref<SortOrder>
  showLabels: Ref<boolean>
  isDark: Ref<boolean>
}

export const formatValue = (value: number, unit: string): string => {
  if (value === 0) return "0";
  if (unit === "ns") {
    if (value >= 1000000) return `${(value / 1000000).toFixed(2)} ms`;
    if (value >= 1000) return `${(value / 1000).toFixed(2)} μs`;
    return `${value.toFixed(0)} ns`;
  }
  if (unit === "b") {
    if (value >= 1073741824) return `${(value / 1073741824).toFixed(2)} GB`;
    if (value >= 1048576) return `${(value / 1048576).toFixed(2)} MB`;
    if (value >= 1024) return `${(value / 1024).toFixed(2)} KB`;
    return `${value.toFixed(0)} B`;
  }
  return value.toString();
};

export const getBaseOptions = (config: BaseChartConfig): Partial<EChartsOption> => {
  const { chartData, isDark } = config
  const textColor = isDark.value ? "#e5e7eb" : "#374151"
  
  return {
    backgroundColor: "transparent",
    tooltip: {
      trigger: "item",
      formatter: (params: any) => {
        const value = formatValue(params.value, chartData.value.statUnit);
        return `${params.marker} <strong>${params.name}</strong><br/>${value}`;
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
  }
}
