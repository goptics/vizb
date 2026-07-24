import { afterAll, beforeAll, describe, expect, it } from 'vitest'
import { ref } from 'vue'
import type { ChartData } from '@/types'
import type { BaseChartConfig } from './baseChartOptions'
import { useHeatmapChartOptions } from './useHeatmapChartOptions'
import { usePieChartOptions } from './usePieChartOptions'
import { useRadarChartOptions } from './useRadarChartOptions'

const originalWindow = globalThis.window
beforeAll(() => {
  ;(globalThis as unknown as { window: { devicePixelRatio: number } }).window = {
    devicePixelRatio: 1,
  }
})
afterAll(() => {
  if (originalWindow === undefined) delete (globalThis as { window?: unknown }).window
  else globalThis.window = originalWindow
})

const chart: ChartData = {
  title: 'value',
  statType: 'value',
  yAxis: [],
  zAxis: [],
  series: [
    { xAxis: 'A', values: [10], benchmarkId: '' },
    { xAxis: 'B', values: [20], benchmarkId: '' },
  ],
  points: [],
}

const config = (): BaseChartConfig => ({
  chartData: ref(chart),
  sort: ref({ enabled: false, order: 'asc' }),
  showLabels: ref(true),
  labelMode: ref('percentage'),
  chartTotal: ref(30),
  isDark: ref(false),
})

const firstLabelFormatter = (series: unknown) =>
  (series as { label: { formatter: (params: any) => string } }[])[0]!.label.formatter

describe('percentage label wiring', () => {
  it('uses percentage-only text for pie slices', () => {
    const { options } = usePieChartOptions(config())
    expect(
      firstLabelFormatter(options.value.series)({ name: 'A', value: 10, percent: 33.33 })
    ).toBe('33.33%')
  })

  it('uses percentage-only text for heatmap cells', () => {
    const heatmapChart: ChartData = { ...chart, yAxis: ['Y'], series: [chart.series[0]!] }
    const cfg = config()
    cfg.chartData.value = heatmapChart
    cfg.chartTotal!.value = 10
    const { options } = useHeatmapChartOptions(cfg)
    expect(firstLabelFormatter(options.value.series)({ data: [0, 0, 10] })).toBe('100%')
  })

  it('omits percentage labels for missing heatmap cells', () => {
    const cfg = config()
    cfg.chartData.value = {
      ...chart,
      yAxis: ['Y'],
      series: [{ xAxis: 'A', values: [null], benchmarkId: '' }],
    }
    const { options } = useHeatmapChartOptions(cfg)
    expect(firstLabelFormatter(options.value.series)({ data: [0, 0, 0] })).toBe('')
  })

  it('uses percentage-only text for radar vertices', () => {
    const { options } = useRadarChartOptions(config())
    expect(firstLabelFormatter(options.value.series)({ value: 10 })).toBe('33.33%')
  })
})
