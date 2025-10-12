import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { type BaseChartConfig, getBaseOptions, formatValue } from './baseChartOptions'
import { getNextColorFor } from '../../lib/utils'
import {
  getChartStyling,
  createPieSeriesConfig
} from './shared'

export function usePieChartOptions(config: BaseChartConfig) {
  const { chartData, sortOrder, showLabels, isDark } = config
  
  const sortedData = computed(() => {
    // Check if we have subjectTotals (multiple workloads case)
    if (chartData.value.subjectTotals) {
      // Prepare subject data
      let subjects = Object.entries(chartData.value.subjectTotals).map(([subject, total]) => ({
        subject,
        values: [total],
        total
      }))
      
      // Sort subjects if sort order is specified
      if (sortOrder.value !== "") {
        subjects.sort((a, b) => {
          if (sortOrder.value === "asc") return a.total - b.total
          return b.total - a.total
        })
      }
      
      // Prepare workload data
      const workloadTotals = new Map<string, number>()
      chartData.value.series.forEach(series => {
        const workloadTotal = series.values.reduce((sum, val) => sum + val, 0)
        workloadTotals.set(series.subject, workloadTotal)
      })
      
      let workloads = Array.from(workloadTotals.entries()).map(([workload, total]) => ({
        subject: workload,
        values: [total],
        total
      }))
      
      // Sort workloads if sort order is specified
      if (sortOrder.value !== "") {
        workloads.sort((a, b) => {
          if (sortOrder.value === "asc") return a.total - b.total
          return b.total - a.total
        })
      }
      
      return { subjects, workloads }
    } else {
      // Single workload case: use series directly
      let seriesWithTotals = chartData.value.series.map((series) => ({
        ...series,
        total: series.values.reduce((sum, val) => sum + val, 0),
      }))

      // Sort if sort order is specified
      if (sortOrder.value !== "") {
        seriesWithTotals.sort((a, b) => {
          if (sortOrder.value === "asc") return (a.total || 0) - (b.total || 0)
          return (b.total || 0) - (a.total || 0)
        })
      }

      return { series: seriesWithTotals }
    }
  })

  const options = computed<EChartsOption>(() => {
    const sorted = sortedData.value
    const styling = getChartStyling(isDark.value)
    const baseOptions = getBaseOptions(config)

    // Check if we have both subjects and workloads (multiple workloads case)
    if ('subjects' in sorted && 'workloads' in sorted && sorted.subjects && sorted.workloads) {
      // Prepare subject pie chart data
      const subjectPieData = sorted.subjects.map(seriesData => ({
        name: seriesData.subject,
        value: seriesData.total || 0,
        itemStyle: { color: getNextColorFor(seriesData.subject) }
      }))

      // Prepare workload pie chart data
      const workloadPieData = sorted.workloads.map(seriesData => ({
        name: seriesData.subject,
        value: seriesData.total || 0,
        itemStyle: { color: getNextColorFor(seriesData.subject) }
      }))

      // Show two pie charts
      return {
        ...baseOptions,
        grid: [
          {
            top: "10%",
            bottom: "10%",
            left: "0%",
            right: "50%",
          },
          {
            top: "10%",
            bottom: "10%",
            left: "50%",
            right: "0%",
          }
        ],
        legend: { show: false },
        title: [
          {
            text: 'Subjects',
            left: '25%',
            top: '3%',
            textAlign: 'center',
            textStyle: {
              fontSize: 14,
              fontWeight: 'bold',
              color: styling.textColor,
            },
          },
          {
            text: 'Workloads',
            left: '75%',
            top: '3%',
            textAlign: 'center',
            textStyle: {
              fontSize: 14,
              fontWeight: 'bold',
              color: styling.textColor,
            },
          }
        ],
        series: [
          createPieSeriesConfig(
            `${chartData.value.statType} (Subjects)`,
            subjectPieData,
            showLabels.value,
            styling,
            ['35%', '65%'],
            ['25%', '50%'],
            (params: any) => {
              const value = formatValue(params.value, chartData.value.statUnit)
              return `${params.name}\n${value} (${params.percent}%)`
            }
          ),
          createPieSeriesConfig(
            `${chartData.value.statType} (Workloads)`,
            workloadPieData,
            showLabels.value,
            styling,
            ['35%', '65%'],
            ['75%', '50%'],
            (params: any) => {
              const value = formatValue(params.value, chartData.value.statUnit)
              return `${params.name}\n${value} (${params.percent}%)`
            }
          )
        ],
      }
    } else {
      // Single pie chart for subjects only
      const singlePieData = sorted.series.map(seriesData => ({
        name: seriesData.subject,
        value: seriesData.total || 0,
        itemStyle: { color: getNextColorFor(seriesData.subject) }
      }))

      return {
        ...baseOptions,
        grid: {
          top: "10%",
          bottom: "10%",
          left: "10%",
          right: "10%",
        },
        legend: { show: false },
        series: [createPieSeriesConfig(
          chartData.value.statType,
          singlePieData,
          showLabels.value,
          styling,
          ['30%', '60%'],
          ['50%', '50%'],
          (params: any) => {
            const value = formatValue(params.value, chartData.value.statUnit)
            return `${params.name}\n${value} (${params.percent}%)`
          }
        )],
      }
    }
  })

  return { options }
}
