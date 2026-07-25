import { ref, type Ref } from 'vue'
import type { ChartData, ChartType, Render3D, ScaleType, Sort } from '@/types'
import type { BaseChartConfig } from '@/composables/charts/baseChartOptions'

const noSort: Sort = { enabled: false, order: 'asc' }

/** Install `window.devicePixelRatio` for node-env chart option tests. Returns cleanup. */
export function installDevicePixelRatio(ratio = 1): () => void {
  const g = globalThis as unknown as { window?: { devicePixelRatio: number } }
  const previous = g.window
  g.window = { devicePixelRatio: ratio }
  return () => {
    if (previous === undefined) {
      delete (globalThis as { window?: unknown }).window
    } else {
      g.window = previous
    }
  }
}

export type ChartDataOverrides = Partial<ChartData>

export function emptyChartData(overrides: ChartDataOverrides = {}): ChartData {
  return {
    title: 'chart',
    statType: 'val',
    yAxis: [],
    zAxis: [],
    series: [],
    points: [],
    axisLabels: {},
    ...overrides,
  }
}

export function makeGroupedChartData(overrides: ChartDataOverrides = {}): ChartData {
  return emptyChartData({
    title: 'revenue',
    statType: 'sum',
    yAxis: ['Hardware', 'Software'],
    series: [
      { xAxis: 'West', values: [10, 30], benchmarkId: '' },
      { xAxis: 'East', values: [20, 40], benchmarkId: '' },
    ],
    axisLabels: { x: 'region', y: 'category' },
    ...overrides,
  })
}

export function makeMixedChartData(overrides: ChartDataOverrides = {}): ChartData {
  return emptyChartData({
    title: 'region vs tax',
    statType: 'mixed',
    axisLabels: { x: 'region', y: 'tax' },
    xCategories: ['West', 'South'],
    mixedTuples: [
      [0, 1926.35],
      [1, 447.38],
    ],
    ...overrides,
  })
}

export function makeValueChartData(overrides: ChartDataOverrides = {}): ChartData {
  return emptyChartData({
    title: 'price · latency',
    statType: 'value',
    axisLabels: { x: 'price', y: 'latency' },
    valueTuples: [
      [100, 12],
      [200, 8],
    ],
    ...overrides,
  })
}

export function makePieChartData(overrides: ChartDataOverrides = {}): ChartData {
  return emptyChartData({
    title: 'share',
    statType: 'count',
    series: [
      { xAxis: 'A', values: [10], benchmarkId: '' },
      { xAxis: 'B', values: [20], benchmarkId: '' },
      { xAxis: 'C', values: [30], benchmarkId: '' },
    ],
    axisLabels: { x: 'group' },
    ...overrides,
  })
}

export function makeRadarChartData(overrides: ChartDataOverrides = {}): ChartData {
  return emptyChartData({
    title: 'profile',
    statType: 'score',
    yAxis: ['speed', 'memory', 'allocs'],
    series: [
      { xAxis: 'algo-a', values: [10, 20, 5], benchmarkId: '' },
      { xAxis: 'algo-b', values: [15, 12, 8], benchmarkId: '' },
    ],
    axisLabels: { x: 'algo', y: 'metric' },
    ...overrides,
  })
}

export function makeHeatmapChartData(overrides: ChartDataOverrides = {}): ChartData {
  return emptyChartData({
    title: 'matrix',
    statType: 'ns',
    yAxis: ['y1', 'y2'],
    series: [
      { xAxis: 'x1', values: [1, 2], benchmarkId: '' },
      { xAxis: 'x2', values: [3, 4], benchmarkId: '' },
    ],
    axisLabels: { x: 'x', y: 'y' },
    ...overrides,
  })
}

export const continuousRender3D: Render3D = {
  mode: 'continuous',
  xValues: [],
  yValues: [],
  zValues: [],
  barSeries: [{ name: 'pts', data: [{ value: [1, 2, 3] }, { value: [4, 5, 6] }] }],
  lineSeries: [{ name: 'pts', data: [{ value: [1, 2, 3] }, { value: [4, 5, 6] }] }],
  cellTotals: {},
}

export const groupedRender3D: Render3D = {
  mode: 'grouped',
  xValues: ['x1', 'x2'],
  yValues: ['y1'],
  zValues: ['zA', 'zB'],
  barSeries: [
    { name: 'zA', data: [{ value: [0, 0, 5] }] },
    { name: 'zB', data: [{ value: [0, 0, 7] }] },
  ],
  lineSeries: [
    { name: 'zA', data: [{ value: [0, 0, 5] }] },
    { name: 'zB', data: [{ value: [0, 0, 7] }] },
  ],
  cellTotals: { '0,0': 12 },
}

export type BaseConfigOverrides = {
  chartData?: ChartData | Ref<ChartData>
  sort?: Sort | Ref<Sort>
  showLabels?: boolean | Ref<boolean>
  isDark?: boolean | Ref<boolean>
  scale?: ScaleType | Ref<ScaleType>
  stack?: boolean | Ref<boolean>
  threeDRotate?: boolean | Ref<boolean>
  threeD?: boolean | Ref<boolean>
  threeDVisualMap?: boolean | Ref<boolean>
  visualMap?: boolean | Ref<boolean>
  smooth?: boolean | Ref<boolean>
  horizontal?: boolean | Ref<boolean>
  visibleZ?: Record<string, boolean> | Ref<Record<string, boolean>>
  arrangementTarget?: string | Ref<string>
  chartType?: ChartType | Ref<ChartType>
}

function asRef<T>(value: T | Ref<T> | undefined, fallback: T): Ref<T> {
  if (value !== undefined && typeof value === 'object' && value !== null && 'value' in value) {
    return value as Ref<T>
  }
  return ref((value as T | undefined) ?? fallback) as Ref<T>
}

/** Build a BaseChartConfig with sensible defaults; pass plain values or refs. */
export function baseConfig(overrides: BaseConfigOverrides = {}): BaseChartConfig {
  const chartData = asRef(overrides.chartData, emptyChartData())
  return {
    chartData,
    sort: asRef(overrides.sort, noSort),
    showLabels: asRef(overrides.showLabels, false),
    isDark: asRef(overrides.isDark, false),
    scale: asRef(overrides.scale, 'linear' as ScaleType),
    stack: asRef(overrides.stack, false),
    threeDRotate: asRef(overrides.threeDRotate, false),
    threeD: asRef(overrides.threeD, false),
    threeDVisualMap: asRef(overrides.threeDVisualMap, false),
    visualMap: asRef(overrides.visualMap, false),
    smooth: asRef(overrides.smooth, false),
    horizontal: asRef(overrides.horizontal, false),
    visibleZ: asRef(overrides.visibleZ, {}),
    arrangementTarget: asRef(overrides.arrangementTarget, 'xy'),
    chartType: asRef(overrides.chartType, 'bar' as ChartType),
  }
}
