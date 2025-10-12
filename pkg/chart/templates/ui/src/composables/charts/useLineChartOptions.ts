import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions, formatValue } from './baseChartOptions'
import { getNextColorFor } from '../../lib/utils'

export function useLineChartOptions(config: BaseChartConfig) {
  const { chartData, sortOrder, showLabels, isDark } = config
  
  const sortedData = computed(() => {
    if (sortOrder.value === "") {
      return {
        series: chartData.value.series,
        xAxisData: chartData.value.workloads,
      };
    }

    // Check if we have multiple workloads (subjectTotals will exist)
    const hasSubjectTotals = chartData.value.subjectTotals !== undefined;

    if (hasSubjectTotals) {
      // Sort X-axis (subjects) based on their totals across all workloads
      const sortedSubjects = chartData.value.workloads
        .map((subject) => ({
          subject,
          total: chartData.value.subjectTotals![subject] || 0,
        }))
        .sort((a, b) => {
          if (sortOrder.value === "asc") return a.total - b.total;
          return b.total - a.total;
        })
        .map((item) => item.subject);

      // Rebuild series data with sorted X-axis order
      const sortedSeries = chartData.value.series.map((series) => {
        const subjectIndexMap = new Map(
          chartData.value.workloads.map((subject, idx) => [subject, idx])
        );

        return {
          ...series,
          values: sortedSubjects.map((subject) => {
            const idx = subjectIndexMap.get(subject);
            return idx !== undefined ? series.values[idx] : 0;
          }),
        };
      });

      return {
        series: sortedSeries,
        xAxisData: sortedSubjects,
      };
    }

    // Single workload case - sort series by their total values
    const seriesWithTotals = chartData.value.series.map((series) => ({
      ...series,
      total: series.values.reduce((sum, val) => sum + val, 0),
    }));
    seriesWithTotals.sort((a, b) => {
      if (sortOrder.value === "asc") return (a.total || 0) - (b.total || 0);
      return (b.total || 0) - (a.total || 0);
    });

    return {
      series: seriesWithTotals,
      xAxisData: chartData.value.workloads,
    };
  });

  const options = computed<EChartsOption>(() => {
    const { series, xAxisData } = sortedData.value;
    const hasMultipleSeries = series.length > 1;
    const hasMultipleWorkloads = chartData.value.workloads.length > 1;
    const textColor = isDark.value ? "#e5e7eb" : "#374151";
    
    // For single workload case, use subjects as x-axis data and single series
    if (!hasMultipleWorkloads) {
      const baseOptions = getBaseOptions(config);
      
      return {
        ...baseOptions,
        tooltip: {
          trigger: "item",
          formatter: (params: any) => {
            const value = formatValue(params.value, chartData.value.statUnit);
            return `${params.marker} <strong>${params.name}</strong><br/>${value}`;
          },
        },
        legend: {
          show: false // No legend for single workload case
        },
        xAxis: {
          type: "category",
          data: series.map(s => s.subject), // Subjects on x-axis
          axisLabel: {
            fontSize: 10,
            color: textColor,
            margin: 8,
          },
          axisLine: {
            lineStyle: { color: isDark.value ? "#4b5563" : "#d1d5db" },
          },
        },
        yAxis: {
          type: "value",
          splitLine: {
            show: true,
            lineStyle: {
              type: "solid",
              opacity: 0.2,
              color: isDark.value ? "#4b5563" : "#d1d5db",
            },
          },
          axisLabel: {
            color: textColor,
          },
          axisLine: {
            lineStyle: { color: isDark.value ? "#4b5563" : "#d1d5db" },
          },
        },
        dataZoom: series.length > 10 ? [
          {
            type: "inside",
            xAxisIndex: 0,
            start: 0,
            end: 100,
          },
          {
            type: "slider",
            xAxisIndex: 0,
            start: 0,
            end: 100,
            height: 20,
            bottom: "2%",
            handleStyle: {
              color: textColor,
            },
            textStyle: {
              color: textColor,
            },
          },
        ] : [],
        series: [{
          name: chartData.value.statType,
          type: "line",
          data: series.map(s => ({
            value: s.values[0] || 0,
            label: {
              show: showLabels.value,
              position: "top",
              formatter: "{c}",
              fontSize: 10,
              color: textColor,
            },
          })),
          itemStyle: { color: "#3b82f6" }, // Single color for all points
          lineStyle: { 
            width: 2,
            type: 'solid'
          },
          smooth: false,
          symbol: 'circle',
          symbolSize: 6,
          emphasis: {
            itemStyle: {
              borderWidth: 2,
              borderColor: '#fff'
            }
          }
        }],
      };
    }

    // Multiple workloads case - original logic
    const legendSpace = Math.min(
      15 + Math.floor((series.length - 1) / 15) * 4,
      35
    );
    const baseOptions = getBaseOptions(config);

    return {
      ...baseOptions,
      grid: {
        left: "3%",
        right: "3%",
        bottom: "10%",
        top: `${legendSpace}%`,
        containLabel: true,
      },
      tooltip: {
        trigger: "axis",
        axisPointer: { type: "shadow" },
        formatter: (params: any) => {
          if (!Array.isArray(params)) return "";
          let result = `<strong>${params[0].axisValue}</strong><br/>`;
          params.forEach((param: any) => {
            const value = formatValue(param.value, chartData.value.statUnit);
            result += `${param.marker} ${param.seriesName}: ${value}<br/>`;
          });
          return result;
        },
      },
      legend: hasMultipleSeries
        ? {
            ...baseOptions.legend,
            data: series.map((s) => s.subject),
          }
        : undefined,
      xAxis: {
        type: "category",
        data: xAxisData,
        axisLabel: {
          interval: "auto", // Auto-hide overlapping labels
          fontSize: 10,
          color: textColor,
          margin: 8,
        },
        axisLine: {
          lineStyle: { color: isDark.value ? "#4b5563" : "#d1d5db" },
        },
      },
      yAxis: {
        type: "value",
        splitLine: {
          show: true,
          lineStyle: {
            type: "solid",
            opacity: 0.2,
            color: isDark.value ? "#4b5563" : "#d1d5db",
          },
        },
        axisLabel: {
          color: textColor,
        },
        axisLine: {
          lineStyle: { color: isDark.value ? "#4b5563" : "#d1d5db" },
        },
      },
      dataZoom: [
          {
            type: "inside",
            xAxisIndex: 0,
            start: 0,
            end: 100,
          },
          {
            type: "slider",
            xAxisIndex: 0,
            start: 0,
            end: 100,
            height: 20,
            bottom: "2%",
            handleStyle: {
              color: textColor,
            },
            textStyle: {
              color: textColor,
            },
          },
        ],
      series: series.map((seriesData) => ({
        name: seriesData.subject,
        type: "line",
        data: seriesData.values.map((value) => ({
          value,
          label: {
            show: showLabels.value,
            position: "top",
            formatter: "{c}",
            fontSize: 10,
            color: textColor,
          },
        })),
        itemStyle: { color: getNextColorFor(seriesData.subject) },
        lineStyle: {
          width: 2,
          type: "solid",
        },
        smooth: false,
        symbol: "circle",
        symbolSize: 6,
        emphasis: {
          focus: "series",
          itemStyle: {
            borderWidth: 2,
            borderColor: "#fff",
          },
        },
      })),
    };
  });

  return { options };
}
