import type { Ref } from "vue";
import type { EChartsOption } from "echarts";
import type { ChartData, Sort } from "../../types/benchmark";
import { createTooltipConfig, getChartStyling } from "./shared/chartConfig";
import { fontSize } from "./shared/common";

export interface BaseChartConfig {
  chartData: Ref<ChartData>;
  sort: Ref<Sort>;
  showLabels: Ref<boolean>;
  isDark: Ref<boolean>;
}

export const getBaseOptions = (
  config: BaseChartConfig
): Partial<EChartsOption> => {
  const { isDark } = config;
  const { textColor, backgroundColor } = getChartStyling(isDark.value);
  return {
    backgroundColor,
    tooltip: createTooltipConfig(false) as EChartsOption['tooltip'],
    toolbox: {
      show: true,
      feature: {
        saveAsImage: {
          show: true,
          type: "jpeg",
          title: "Save",
          pixelRatio: 2,
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
      left: "center",
      itemWidth: 10,
      itemHeight: 10,
      textStyle: { fontSize, color: textColor },
    },
    emphasis: {
      focus: "series",
    },
  };
};
