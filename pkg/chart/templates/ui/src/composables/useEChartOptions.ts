import { computed, type Ref } from 'vue'
import type { ChartData, SeriesData } from '../types/benchmark'
import type { SortOrder } from '../types/benchmark'
import type { EChartsOption } from 'echarts'

const COLOR_PALETTE = [
  '#5470c6', '#91cc75', '#fac858', '#ee6666', '#73c0de',
  '#3ba272', '#fc8452', '#9a60b4', '#ea7ccc', '#d14a61',
  '#6e7074', '#546570', '#c4ccd3', '#626c91', '#a0a7e6',
  '#c4ebad', '#96dee8'
]

export function useEChartOptions(
  chartData: Ref<ChartData>,
  sortOrder: Ref<SortOrder>,
  showLabels: Ref<boolean>,
  isDark: Ref<boolean>
) {
  const formatValue = (value: number, unit: string): string => {
    if (value === 0) return '0'
    if (unit === 'ns') {
      if (value >= 1000000) return `${(value / 1000000).toFixed(2)} ms`
      if (value >= 1000) return `${(value / 1000).toFixed(2)} μs`
      return `${value.toFixed(0)} ns`
    }
    if (unit === 'B') {
      if (value >= 1073741824) return `${(value / 1073741824).toFixed(2)} GB`
      if (value >= 1048576) return `${(value / 1048576).toFixed(2)} MB`
      if (value >= 1024) return `${(value / 1024).toFixed(2)} KB`
      return `${value.toFixed(0)} B`
    }
    return value.toString()
  }

  const sortedSeries = computed<SeriesData[]>(() => {
    if (sortOrder.value === 'default') return chartData.value.series
    const seriesWithTotals = chartData.value.series.map(series => ({
      ...series,
      total: series.values.reduce((sum, val) => sum + val, 0)
    }))
    seriesWithTotals.sort((a, b) => {
      if (sortOrder.value === 'asc') return a.total - b.total
      return b.total - a.total
    })
    return seriesWithTotals
  })

  const options = computed<EChartsOption>(() => {
    const series = sortedSeries.value
    const hasMultipleSeries = series.length > 1
    const legendSpace = Math.min(15 + Math.floor((series.length - 1) / 15) * 4, 35)

    return {
      backgroundColor: 'transparent',
      grid: {
        left: '3%',
        right: '3%',
        bottom: '10%',
        top: `${legendSpace}%`,
        containLabel: true
      },
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'shadow' },
        formatter: (params: any) => {
          if (!Array.isArray(params)) return ''
          let result = `<strong>${params[0].axisValue}</strong><br/>`
          params.forEach((param: any) => {
            const value = formatValue(param.value, chartData.value.statUnit)
            result += `${param.marker} ${param.seriesName}: ${value}<br/>`
          })
          return result
        }
      },
      toolbox: {
        show: true,
        right: '2%',
        feature: {
          saveAsImage: { show: true, type: 'png', title: 'Save', pixelRatio: 2 },
          dataZoom: { show: true, yAxisIndex: 'none', title: { zoom: 'Area Zoom', back: 'Reset Zoom' } }
        }
      },
      dataZoom: [
        { type: 'slider', show: series.length > 10, xAxisIndex: 0, start: 0, end: 100, bottom: 0, height: 20 },
        { type: 'inside', xAxisIndex: 0, start: 0, end: 100 }
      ],
      legend: hasMultipleSeries ? {
        show: true, top: '7%', left: 'center', itemWidth: 10, itemHeight: 10,
        textStyle: { fontSize: 12 }, data: series.map(s => s.subject)
      } : undefined,
      xAxis: {
        type: 'category',
        data: chartData.value.workloads,
        axisLabel: { interval: 0, rotate: 0, fontSize: 11 }
      },
      yAxis: {
        type: 'value',
        splitLine: { show: true, lineStyle: { type: 'solid', opacity: 0.2 } }
      },
      series: series.map((seriesData, index) => ({
        name: seriesData.subject,
        type: 'bar',
        data: seriesData.values.map(value => ({
          value,
          label: {
            show: showLabels.value,
            position: 'top',
            formatter: () => formatValue(value, chartData.value.statUnit),
            fontSize: 10
          }
        })),
        itemStyle: { color: COLOR_PALETTE[index % COLOR_PALETTE.length] }
      }))
    }
  })

  return { options }
}
