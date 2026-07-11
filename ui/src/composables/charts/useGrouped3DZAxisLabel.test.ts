import { afterAll, beforeAll, describe, expect, it } from 'vitest'
import { ref } from 'vue'
import type { ChartData, Render3D } from '@/types'
import type { BaseChartConfig } from './baseChartOptions'
import { useBar3DChartOptions } from './useBar3DChartOptions'
import { useLine3DChartOptions } from './useLine3DChartOptions'
import { useScatter3DChartOptions } from './useScatter3DChartOptions'

const originalDPR = (globalThis as { window?: { devicePixelRatio: number } }).window
  ?.devicePixelRatio
beforeAll(() => {
  ;(globalThis as unknown as { window: { devicePixelRatio: number } }).window = {
    devicePixelRatio: 1,
  }
})
afterAll(() => {
  if (originalDPR === undefined) {
    delete (globalThis as { window?: unknown }).window
  } else {
    ;(globalThis as unknown as { window: { devicePixelRatio: number } }).window = {
      devicePixelRatio: originalDPR,
    }
  }
})

const groupedRender: Render3D = {
  mode: 'grouped',
  xValues: ['East'],
  yValues: ['Widget'],
  zValues: ['A', 'B'],
  barSeries: [
    { name: 'A', data: [{ value: [0, 0, 10] }] },
    { name: 'B', data: [{ value: [0, 0, 5] }] },
  ],
  lineSeries: [
    { name: 'A', data: [{ value: [0, 0, 10] }] },
    { name: 'B', data: [{ value: [0, 0, 5] }] },
  ],
  cellTotals: { '0,0': 15 },
}

const valueRender: Render3D = {
  ...groupedRender,
  mode: 'value',
  zValues: [],
  barSeries: [{ name: 'sales', data: [{ value: [0, 0, 15] }] }],
  lineSeries: [{ name: 'sales', data: [{ value: [0, 0, 15] }] }],
}

const makeConfig = (render3D: Render3D): BaseChartConfig => ({
  chartData: ref({
    title: 'sales',
    statType: 'sum',
    yAxis: ['Widget'],
    zAxis: ['A', 'B'],
    series: [],
    points: [
      { xAxis: 'East', yAxis: 'Widget', zAxis: 'A', value: 10 },
      { xAxis: 'East', yAxis: 'Widget', zAxis: 'B', value: 5 },
    ],
    axisLabels: { x: 'region', y: 'product', z: 'category' },
    render3D,
  } satisfies ChartData),
  sort: ref({ enabled: false, order: 'asc' }),
  showLabels: ref(false),
  isDark: ref(false),
  scale: ref('linear'),
  threeDRotate: ref(false),
  visibleZ: ref({}),
  threeD: ref(true),
  threeDVisualMap: ref(false),
})

const chartFactories = [
  ['bar', useBar3DChartOptions],
  ['line', useLine3DChartOptions],
  ['scatter', useScatter3DChartOptions],
] as const

describe('grouped 3D z axis labels', () => {
  it.each(chartFactories)(
    'keeps the grouped z-axis name invisible for %s charts (merge-safe)',
    (_name, useOptions) => {
      const { options } = useOptions(makeConfig(groupedRender))
      const zAxis = options.value.zAxis3D as {
        name?: string
        nameGap?: number
        nameTextStyle?: { color?: string }
      }

      // Non-empty name + nameGap preserve framing; transparent text hides the label.
      expect(zAxis.name).toBe('category')
      expect(zAxis.nameGap).toBe(25)
      expect(zAxis.nameTextStyle?.color).toBe('transparent')
    }
  )

  it.each(chartFactories)('keeps the value-mode z-axis name for %s charts', (_name, useOptions) => {
    const { options } = useOptions(makeConfig(valueRender))
    const zAxis = options.value.zAxis3D as {
      name?: string
      nameTextStyle?: { color?: string }
    }

    expect(zAxis.name).toBe('sales')
    expect(zAxis.nameTextStyle?.color).not.toBe('transparent')
  })
})
