import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import { ref } from 'vue'
import type { ChartData } from '@/types'
import type { BaseChartConfig } from './baseChartOptions'
import { useBarChartOptions } from './useBarChartOptions'

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

const makeMixedChartData = (): ChartData => ({
  title: 'region vs tax',
  statType: 'mixed',
  yAxis: [],
  zAxis: [],
  series: [],
  points: [],
  axisLabels: { x: 'region', y: 'tax' },
  xCategories: ['West', 'South'],
  mixedTuples: [
    [0, 1926.35],
    [1, 447.38],
  ],
})

const makeMixedConfig = (): BaseChartConfig => ({
  chartData: ref(makeMixedChartData()),
  sort: ref({ enabled: false, order: 'asc' }),
  showLabels: ref(false),
  isDark: ref(false),
})

const makeStackedGroupedChartData = (): ChartData => ({
  title: 'revenue',
  statType: 'sum',
  yAxis: ['Hardware', 'Software'],
  zAxis: [],
  series: [
    { xAxis: 'West', values: [10, 30], benchmarkId: '' },
    { xAxis: 'East', values: [20, 40], benchmarkId: '' },
  ],
  points: [],
  axisLabels: { x: 'region', y: 'category' },
})

const makeStackedGroupedConfig = (
  stack = false,
  horizontal = false,
  showLabels = true,
  isDark = false
): BaseChartConfig => ({
  chartData: ref(makeStackedGroupedChartData()),
  sort: ref({ enabled: false, order: 'asc' }),
  showLabels: ref(showLabels),
  isDark: ref(isDark),
  scale: ref<'linear' | 'log'>('linear'),
  stack: ref(stack),
  horizontal: ref(horizontal),
})

describe('useBarChartOptions — grouped mode', () => {
  it('emits stacked bar series when stack is enabled', () => {
    const { options } = useBarChartOptions(makeStackedGroupedConfig(true))
    const series = options.value.series as { stack?: string }[]
    expect(series).toHaveLength(2)
    expect(series.every((s) => s.stack === 'total')).toBe(true)
  })

  it('does not stack grouped bars by default', () => {
    const { options } = useBarChartOptions(makeStackedGroupedConfig(false))
    const series = options.value.series as { stack?: string }[]
    expect(series.every((s) => s.stack === undefined)).toBe(true)
  })

  it.each([
    [false, false],
    [true, true],
  ])(
    'centers readable labels inside stacked bars when horizontal is %s and dark mode is %s',
    (horizontal, isDark) => {
      const { options } = useBarChartOptions(
        makeStackedGroupedConfig(true, horizontal, true, isDark)
      )
      const series = options.value.series as {
        label: {
          show: boolean
          position: string
          color: string
          textBorderColor?: string
          textBorderWidth?: number
        }
      }[]

      expect(series.every((s) => s.label.show)).toBe(true)
      expect(series.every((s) => s.label.position === 'inside')).toBe(true)
      expect(series.every((s) => s.label.color === '#fff')).toBe(true)
      expect(series.every((s) => s.label.textBorderColor === 'rgba(0,0,0,0.5)')).toBe(true)
      expect(series.every((s) => s.label.textBorderWidth === 2)).toBe(true)
    }
  )

  it.each([
    [false, 'top'],
    [true, 'right'],
  ])('keeps labels at the bar tip when horizontal is %s', (horizontal, position) => {
    const { options } = useBarChartOptions(makeStackedGroupedConfig(false, horizontal))
    const series = options.value.series as {
      label: { position: string; color: string; textBorderWidth?: number }
    }[]

    expect(series.every((s) => s.label.position === position)).toBe(true)
    expect(series.every((s) => s.label.color === '#374151')).toBe(true)
    expect(series.every((s) => s.label.textBorderWidth === undefined)).toBe(true)
  })

  it('keeps labels hidden when stacking is enabled and show labels is off', () => {
    const { options } = useBarChartOptions(makeStackedGroupedConfig(true, false, false))
    const series = options.value.series as { label: { show: boolean; position: string } }[]

    expect(series.every((s) => !s.label.show)).toBe(true)
    expect(series.every((s) => s.label.position === 'inside')).toBe(true)
  })
})

describe('useBarChartOptions — mixed mode', () => {
  it('emits bar series with mixedTuples as data', () => {
    const { options } = useBarChartOptions(makeMixedConfig())
    const s = (options.value.series as { type: string; data: [number, number][] }[])[0]!
    expect(s.type).toBe('bar')
    expect(s.data).toEqual([
      [0, 1926.35],
      [1, 447.38],
    ])
  })

  it('emits category xAxis', () => {
    const { options } = useBarChartOptions(makeMixedConfig())
    expect((options.value.xAxis as { type: string; data: string[] }).type).toBe('category')
    expect((options.value.xAxis as { data: string[] }).data).toEqual(['West', 'South'])
  })

  it('uses axis trigger with themed shadow pointer for category x', () => {
    const { options } = useBarChartOptions(makeMixedConfig())
    const tooltip = options.value.tooltip as {
      trigger?: string
      axisPointer?: {
        type?: string
        snap?: boolean
        shadowStyle?: { color?: string; opacity?: number }
      }
    }
    expect(tooltip.trigger).toBe('axis')
    expect(tooltip.axisPointer?.type).toBe('shadow')
    expect(tooltip.axisPointer?.snap).toBeUndefined()
    expect(tooltip.axisPointer?.shadowStyle?.color).toBe('#d1d5db')
    expect(tooltip.axisPointer?.shadowStyle?.opacity).toBe(0.4)
  })
})

const makeSimpleChartData = (): ChartData => ({
  title: 'items',
  statType: 'counts',
  yAxis: [],
  zAxis: [],
  series: [
    { xAxis: 'A', values: [10], benchmarkId: '' },
    { xAxis: 'B', values: [20], benchmarkId: '' },
    { xAxis: 'C', values: [30], benchmarkId: '' },
  ],
  points: [],
  axisLabels: { x: 'category', y: 'value' },
})

const makeSimpleConfig = (horizontal: boolean): BaseChartConfig => ({
  chartData: ref(makeSimpleChartData()),
  sort: ref({ enabled: false, order: 'asc' }),
  showLabels: ref(false),
  isDark: ref(false),
  horizontal: ref(horizontal),
})

const makeGroupedChartData = (): ChartData => ({
  title: 'regions',
  statType: 'counts',
  yAxis: ['North', 'South'],
  zAxis: [],
  series: [
    { xAxis: 'A', values: [10, 20], benchmarkId: '' },
    { xAxis: 'B', values: [15, 25], benchmarkId: '' },
  ],
  points: [],
  axisLabels: { x: 'category', y: 'region' },
})

const makeGroupedConfig = (horizontal: boolean): BaseChartConfig => ({
  chartData: ref(makeGroupedChartData()),
  sort: ref({ enabled: false, order: 'asc' }),
  showLabels: ref(false),
  isDark: ref(false),
  horizontal: ref(horizontal),
})

describe('useBarChartOptions — horizontal mode', () => {
  it('renders horizontal 1D bars with value xAxis and category yAxis', () => {
    const { options } = useBarChartOptions(makeSimpleConfig(true))
    const opt = options.value
    expect((opt.xAxis as { type: string }).type).toBe('value')
    expect((opt.yAxis as { type: string }).type).toBe('category')
    expect((opt.yAxis as { data: string[] }).data).toEqual(['A', 'B', 'C'])
    expect((opt.yAxis as { name?: string }).name).toBe('category')
    expect((opt.xAxis as { name?: string }).name).toBeUndefined()
  })

  it('renders horizontal grouped bars with correct series', () => {
    const { options } = useBarChartOptions(makeGroupedConfig(true))
    const opt = options.value
    expect((opt.xAxis as { type: string }).type).toBe('value')
    expect((opt.yAxis as { type: string }).type).toBe('category')
    expect((opt.yAxis as { name?: string }).name).toBe('category')
    expect((opt.xAxis as { name?: string }).name).toBeUndefined()
    const series = opt.series as { type: string; name: string; data: number[] }[]
    expect(series.length).toBe(2)
    expect(series[0]!.name).toBe('North')
    expect(series[1]!.name).toBe('South')
    expect(series[0]!.data).toEqual([10, 15])
  })

  it('places grouped horizontal legend at bottom center', () => {
    const { options } = useBarChartOptions(makeGroupedConfig(true))
    const legend = options.value.legend as { left?: string; bottom?: number; top?: number }
    expect(legend.left).toBe('center')
    expect(legend.bottom).toBe(0)
    expect(legend.top).toBeUndefined()
    expect(options.value.title).toBeUndefined()
  })

  it('renders vertical bars by default (horizontal not set)', () => {
    const { options } = useBarChartOptions(makeSimpleConfig(false))
    const opt = options.value
    expect((opt.xAxis as { type: string }).type).toBe('category')
    expect((opt.yAxis as { type: string }).type).toBe('value')
  })
})

const makeValueChartData = (): ChartData => ({
  title: 'price · latency',
  statType: 'value',
  yAxis: [],
  zAxis: [],
  series: [],
  points: [],
  axisLabels: { x: 'price', y: 'latency' },
  valueTuples: [
    [100, 12],
    [200, 8],
  ],
})

describe('useBarChartOptions — value mode and branches', () => {
  it('emits bar series for valueTuples', () => {
    const { options } = useBarChartOptions({
      chartData: ref(makeValueChartData()),
      sort: ref({ enabled: false, order: 'asc' }),
      showLabels: ref(false),
      isDark: ref(false),
    })
    const s = (options.value.series as { type: string; data: unknown[] }[])[0]!
    expect(s.type).toBe('bar')
    expect(s.data).toHaveLength(2)
  })

  it('sorts single-x grouped series by value', () => {
    const chartData: ChartData = {
      title: 'one',
      statType: 'sum',
      yAxis: ['High', 'Low'],
      zAxis: [],
      series: [{ xAxis: 'only', values: [30, 10], benchmarkId: '' }],
      points: [],
      axisLabels: { x: 'x', y: 'y' },
    }
    const asc = useBarChartOptions({
      chartData: ref(chartData),
      sort: ref({ enabled: true, order: 'asc' }),
      showLabels: ref(false),
      isDark: ref(false),
    })
    const desc = useBarChartOptions({
      chartData: ref(chartData),
      sort: ref({ enabled: true, order: 'desc' }),
      showLabels: ref(false),
      isDark: ref(false),
    })
    expect((asc.options.value.series as { name: string }[]).map((s) => s.name)).toEqual([
      'Low',
      'High',
    ])
    expect((desc.options.value.series as { name: string }[]).map((s) => s.name)).toEqual([
      'High',
      'Low',
    ])
  })

  it('nulls non-positive values on log scale', () => {
    const chartData: ChartData = {
      title: 'log',
      statType: 'v',
      yAxis: [],
      zAxis: [],
      series: [
        { xAxis: 'a', values: [0], benchmarkId: '' },
        { xAxis: 'b', values: [-1], benchmarkId: '' },
        { xAxis: 'c', values: [5], benchmarkId: '' },
        { xAxis: 'd', values: [null], benchmarkId: '' },
      ],
      points: [],
    }
    const { options } = useBarChartOptions({
      chartData: ref(chartData),
      sort: ref({ enabled: false, order: 'asc' }),
      showLabels: ref(false),
      isDark: ref(false),
      scale: ref('log'),
    })
    const series = options.value.series as { data: (number | null)[] }[]
    expect(series[0]!.data).toEqual([null, null, 5, null])
  })

  it('adds horizontal dataZoom for large simple categories', () => {
    const many = Array.from({ length: 60 }, (_, i) => `c${i}`)
    const chartData: ChartData = {
      title: 'wide',
      statType: 'v',
      yAxis: [],
      zAxis: [],
      series: many.map((x, i) => ({ xAxis: x, values: [i + 1], benchmarkId: '' })),
      points: [],
      axisLabels: { x: 'cat' },
    }
    const { options } = useBarChartOptions({
      chartData: ref(chartData),
      sort: ref({ enabled: false, order: 'asc' }),
      showLabels: ref(false),
      isDark: ref(false),
      horizontal: ref(true),
    })
    expect(options.value.dataZoom).toBeDefined()
    expect((options.value.grid as { right?: number }).right).toBe(44)
  })

  it('adds vertical dataZoom for large simple categories', () => {
    const many = Array.from({ length: 60 }, (_, i) => `c${i}`)
    const chartData: ChartData = {
      title: 'wide',
      statType: 'v',
      yAxis: [],
      zAxis: [],
      series: many.map((x, i) => ({ xAxis: x, values: [i + 1], benchmarkId: '' })),
      points: [],
      axisLabels: { x: 'cat' },
    }
    const { options } = useBarChartOptions({
      chartData: ref(chartData),
      sort: ref({ enabled: false, order: 'asc' }),
      showLabels: ref(false),
      isDark: ref(false),
    })
    expect(options.value.dataZoom).toBeDefined()
  })

  it('handles horizontal simple bars without x label', () => {
    const chartData = makeSimpleChartData()
    chartData.axisLabels = {}
    const { options } = useBarChartOptions({
      chartData: ref(chartData),
      sort: ref({ enabled: false, order: 'asc' }),
      showLabels: ref(false),
      isDark: ref(false),
      horizontal: ref(true),
    })
    expect((options.value.grid as { left?: string | number }).left).toBe('3%')
  })

  it('adds legend title for multi-series vertical bars with y label', () => {
    const { options } = useBarChartOptions(makeGroupedConfig(false))
    expect(options.value.title).toBeDefined()
  })

  it('does not stack when scale is log even if stack is true', () => {
    const { options } = useBarChartOptions({
      chartData: ref(makeStackedGroupedChartData()),
      sort: ref({ enabled: false, order: 'asc' }),
      showLabels: ref(false),
      isDark: ref(false),
      scale: ref('log'),
      stack: ref(true),
    })
    const series = options.value.series as { stack?: string }[]
    expect(series.every((s) => s.stack === undefined)).toBe(true)
  })

  it('adds dataZoom for large horizontal grouped categories', () => {
    const many = Array.from({ length: 60 }, (_, i) => `c${i}`)
    const chartData: ChartData = {
      title: 'wide',
      statType: 'v',
      yAxis: ['N', 'S'],
      zAxis: [],
      series: many.map((x) => ({ xAxis: x, values: [1, 2], benchmarkId: '' })),
      points: [],
      axisLabels: { x: 'cat', y: 'reg' },
    }
    const { options } = useBarChartOptions({
      chartData: ref(chartData),
      sort: ref({ enabled: false, order: 'asc' }),
      showLabels: ref(false),
      isDark: ref(false),
      horizontal: ref(true),
    })
    expect(options.value.dataZoom).toBeDefined()
  })
})

describe('useBarChartOptions — optional branch edges', () => {
  it('handles missing values and single-series horizontal without y multi', () => {
    const chartData: ChartData = {
      title: 't',
      statType: 'v',
      yAxis: ['Only'],
      zAxis: [],
      series: [
        { xAxis: 'A', values: [], benchmarkId: '' },
        { xAxis: 'B', values: [null], benchmarkId: '' },
      ],
      points: [],
      axisLabels: {},
    }
    const { options } = useBarChartOptions({
      chartData: ref(chartData),
      sort: ref({ enabled: true, order: 'asc' }),
      showLabels: ref(false),
      isDark: ref(false),
      horizontal: ref(true),
    })
    // single y group → no multi legend bottom band beyond default
    expect((options.value.series as { data: (number | null)[] }[])[0]!.data).toEqual([null, null])
  })

  it('sorts with empty data slots and large vertical grouped dataZoom', () => {
    const many = Array.from({ length: 60 }, (_, i) => `c${i}`)
    const chartData: ChartData = {
      title: 't',
      statType: 'v',
      yAxis: ['A', 'B'],
      zAxis: [],
      series: [{ xAxis: 'only', values: [null, 5], benchmarkId: '' }],
      points: [],
      axisLabels: { x: 'x' }, // no y label → no legend title
    }
    const sorted = useBarChartOptions({
      chartData: ref(chartData),
      sort: ref({ enabled: true, order: 'asc' }),
      showLabels: ref(false),
      isDark: ref(false),
    })
    expect((sorted.options.value.series as { name: string }[]).map((s) => s.name)).toEqual([
      'A',
      'B',
    ])
    expect(sorted.options.value.title).toBeUndefined()

    const wide: ChartData = {
      title: 't',
      statType: 'v',
      yAxis: ['A', 'B'],
      zAxis: [],
      series: many.map((x) => ({ xAxis: x, values: [1, 2], benchmarkId: '' })),
      points: [],
      axisLabels: { x: 'x', y: 'y' },
    }
    const { options } = useBarChartOptions({
      chartData: ref(wide),
      sort: ref({ enabled: false, order: 'asc' }),
      showLabels: ref(false),
      isDark: ref(false),
    })
    expect(options.value.dataZoom).toBeDefined()
    expect(options.value.title).toBeDefined()
  })

  it('uses default scale when scale ref omitted', () => {
    const { options } = useBarChartOptions({
      chartData: ref(makeSimpleChartData()),
      sort: ref({ enabled: false, order: 'asc' }),
      showLabels: ref(false),
      isDark: ref(false),
    })
    expect((options.value.series as unknown[]).length).toBe(1)
  })
})

describe('useBarChartOptions — nullish value edges', () => {
  it('maps missing simple values via ?? null', () => {
    const chartData: ChartData = {
      title: 't',
      statType: 'v',
      yAxis: [],
      zAxis: [],
      series: [{ xAxis: 'A', values: [], benchmarkId: '' }],
      points: [],
    }
    const { options } = useBarChartOptions({
      chartData: ref(chartData),
      sort: ref({ enabled: false, order: 'asc' }),
      showLabels: ref(false),
      isDark: ref(false),
    })
    expect((options.value.series as { data: (number | null)[] }[])[0]!.data).toEqual([null])
  })

  it('sort compares undefined data slots as 0', () => {
    const chartData: ChartData = {
      title: 't',
      statType: 'v',
      yAxis: ['A', 'B'],
      zAxis: [],
      series: [{ xAxis: 'only', values: [], benchmarkId: '' }],
      points: [],
    }
    const { options } = useBarChartOptions({
      chartData: ref(chartData),
      sort: ref({ enabled: true, order: 'desc' }),
      showLabels: ref(false),
      isDark: ref(false),
    })
    expect((options.value.series as { name: string }[]).map((s) => s.name)).toEqual(['A', 'B'])
  })
})

describe('useBarChartOptions — horizontal nullish values', () => {
  it('maps missing horizontal simple values via ?? null', () => {
    const chartData: ChartData = {
      title: 't',
      statType: 'v',
      yAxis: [],
      zAxis: [],
      series: [{ xAxis: 'A', values: [], benchmarkId: '' }],
      points: [],
      axisLabels: { x: 'cat' },
    }
    const { options } = useBarChartOptions({
      chartData: ref(chartData),
      sort: ref({ enabled: false, order: 'asc' }),
      showLabels: ref(false),
      isDark: ref(false),
      horizontal: ref(true),
    })
    expect((options.value.series as { data: (number | null)[] }[])[0]!.data).toEqual([null])
  })
})

describe('useBarChartOptions — borderRadius', () => {
  const makeMultiSeriesChartData = (): ChartData => ({
    title: 'stacked test',
    statType: 'sum',
    yAxis: ['Series A', 'Series B', 'Series C'],
    zAxis: [],
    series: [
      { xAxis: 'X1', values: [10, 20, 30], benchmarkId: '' },
      { xAxis: 'X2', values: [15, 25, 35], benchmarkId: '' },
    ],
    points: [],
    axisLabels: { x: 'category', y: 'series' },
  })

  type SeriesStyle = { itemStyle?: { borderRadius?: number[] } }

  it('omits borderRadius when undefined or 0', () => {
    for (const borderRadius of [undefined, 0] as const) {
      const { options } = useBarChartOptions({
        chartData: ref(makeSimpleChartData()),
        sort: ref({ enabled: false, order: 'asc' }),
        showLabels: ref(false),
        isDark: ref(false),
        ...(borderRadius === undefined ? {} : { borderRadius: ref(borderRadius) }),
      })
      const series = options.value.series as SeriesStyle[]
      expect(series[0]?.itemStyle?.borderRadius).toBeUndefined()
    }
  })

  it('applies free-outer corners to non-stacked bars', () => {
    const { options } = useBarChartOptions({
      chartData: ref(makeSimpleChartData()),
      sort: ref({ enabled: false, order: 'asc' }),
      showLabels: ref(false),
      isDark: ref(false),
      borderRadius: ref(8),
    })
    const series = options.value.series as SeriesStyle[]
    expect(series[0]?.itemStyle?.borderRadius).toEqual([8, 8, 0, 0])
  })

  it('applies radius to every series when not stacked', () => {
    const { options } = useBarChartOptions({
      chartData: ref(makeGroupedChartData()),
      sort: ref({ enabled: false, order: 'asc' }),
      showLabels: ref(false),
      isDark: ref(false),
      borderRadius: ref(8),
      stack: ref(false),
    })
    const series = options.value.series as SeriesStyle[]
    expect(series).toHaveLength(2)
    expect(series[0]?.itemStyle?.borderRadius).toEqual([8, 8, 0, 0])
    expect(series[1]?.itemStyle?.borderRadius).toEqual([8, 8, 0, 0])
  })

  it('applies radius only to the top stack segment', () => {
    const { options } = useBarChartOptions({
      chartData: ref(makeMultiSeriesChartData()),
      sort: ref({ enabled: false, order: 'asc' }),
      showLabels: ref(false),
      isDark: ref(false),
      borderRadius: ref(8),
      stack: ref(true),
    })
    const series = options.value.series as SeriesStyle[]
    expect(series).toHaveLength(3)
    expect(series[0]?.itemStyle?.borderRadius).toBeUndefined()
    expect(series[1]?.itemStyle?.borderRadius).toBeUndefined()
    expect(series[2]?.itemStyle?.borderRadius).toEqual([8, 8, 0, 0])
  })

  it('uses free-outer corners for horizontal bars', () => {
    const { options } = useBarChartOptions({
      chartData: ref(makeSimpleChartData()),
      sort: ref({ enabled: false, order: 'asc' }),
      showLabels: ref(false),
      isDark: ref(false),
      borderRadius: ref(8),
      horizontal: ref(true),
    })
    const series = options.value.series as SeriesStyle[]
    expect(series[0]?.itemStyle?.borderRadius).toEqual([0, 8, 8, 0])
  })

  it('applies horizontal free-outer corners only to the top stack segment', () => {
    const { options } = useBarChartOptions({
      chartData: ref(makeMultiSeriesChartData()),
      sort: ref({ enabled: false, order: 'asc' }),
      showLabels: ref(false),
      isDark: ref(false),
      borderRadius: ref(8),
      stack: ref(true),
      horizontal: ref(true),
    })
    const series = options.value.series as SeriesStyle[]
    expect(series[0]?.itemStyle?.borderRadius).toBeUndefined()
    expect(series[1]?.itemStyle?.borderRadius).toBeUndefined()
    expect(series[2]?.itemStyle?.borderRadius).toEqual([0, 8, 8, 0])
  })

  it('applies borderRadius in mixed and value modes', () => {
    for (const chartData of [makeMixedChartData(), makeValueChartData()]) {
      const { options } = useBarChartOptions({
        chartData: ref(chartData),
        sort: ref({ enabled: false, order: 'asc' }),
        showLabels: ref(false),
        isDark: ref(false),
        borderRadius: ref(8),
      })
      const series = options.value.series as (SeriesStyle & { type: string })[]
      expect(series[0]?.type).toBe('bar')
      expect(series[0]?.itemStyle?.borderRadius).toEqual([8, 8, 0, 0])
    }
  })
})
