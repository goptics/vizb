import { describe, it, expect, afterAll } from 'vitest'
import { ref } from 'vue'
import type { ChartData, ScaleInput } from '@/types'
import type { BaseChartConfig } from './baseChartOptions'
import { useLineChartOptions } from './useLineChartOptions'
import { installDevicePixelRatio } from '@/test-utils'
import { LARGE_X_THRESHOLD, VALUE_MODE_GRID_TOP } from './shared/chartConfig'

let restoreDpr = installDevicePixelRatio()
afterAll(() => restoreDpr())

const makeMixedChartData = (): ChartData => ({
  title: 'region vs latency',
  statType: 'mixed',
  yAxis: [],
  zAxis: [],
  series: [],
  points: [],
  axisLabels: { x: 'region', y: 'latency' },
  xCategories: ['Asia', 'EU'],
  mixedTuples: [
    [0, 12],
    [1, 11],
  ],
})

const makeMixedConfig = (): BaseChartConfig => ({
  chartData: ref(makeMixedChartData()),
  sort: ref({ enabled: false, order: 'asc' }),
  showLabels: ref(false),
  isDark: ref(false),
})

const makeSmoothMixedConfig = (): BaseChartConfig => ({
  ...makeMixedConfig(),
  smooth: ref(true),
})

const makeValueChartData = (): ChartData => ({
  title: 'price vs latency',
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

const makeValueConfig = (): BaseChartConfig => ({
  chartData: ref(makeValueChartData()),
  sort: ref({ enabled: false, order: 'asc' }),
  showLabels: ref(false),
  isDark: ref(false),
})

const makeSmoothValueConfig = (): BaseChartConfig => ({
  ...makeValueConfig(),
  smooth: ref(true),
})

const makeGroupedChartData = (): ChartData => ({
  title: 'sales by month',
  statType: 'grouped',
  yAxis: ['North', 'South'],
  zAxis: [],
  series: [
    { xAxis: 'Jan', values: [10, 8], benchmarkId: 'jan' },
    { xAxis: 'Feb', values: [12, 9], benchmarkId: 'feb' },
  ],
  points: [],
  axisLabels: { x: 'month', y: 'region' },
})

const makeGroupedConfig = (opts: { smooth?: boolean; stack?: boolean } = {}): BaseChartConfig => ({
  chartData: ref(makeGroupedChartData()),
  sort: ref({ enabled: false, order: 'asc' }),
  showLabels: ref(false),
  isDark: ref(false),
  scale: ref<ScaleInput>('linear'),
  smooth: ref(opts.smooth ?? false),
  stack: ref(opts.stack ?? false),
})

const makeNumericStepChartData = (
  yAxis: string[] = ['train', 'val'],
  points: [string, number[]][] = [
    ['1', [10, 8]],
    ['2', [0, 9]],
    ['4', [12, 6]],
  ]
): ChartData => ({
  title: 'loss vs step',
  statType: 'grouped',
  yAxis,
  zAxis: [],
  series: points.map(([xAxis, values]) => ({ xAxis, values, benchmarkId: xAxis })),
  points: [],
  axisLabels: { x: 'step', y: 'split' },
})

const makeNumericStepConfig = (
  scale: ScaleInput,
  data: ChartData = makeNumericStepChartData()
) => ({
  chartData: ref(data),
  sort: ref({ enabled: false, order: 'asc' as const }),
  showLabels: ref(false),
  isDark: ref(false),
  scale: ref(scale),
  smooth: ref(false),
  stack: ref(false),
})

const axisOf = (options: { xAxis?: unknown; yAxis?: unknown }) => ({
  xAxis: options.xAxis as { type?: string; logBase?: number; data?: string[] },
  yAxis: options.yAxis as { type?: string; logBase?: number },
})

const tooltipOf = (options: { tooltip?: unknown }) =>
  options.tooltip as {
    trigger?: string
    axisPointer?: { type?: string; snap?: boolean }
    formatter?: (params: unknown) => string
  }

describe('useLineChartOptions — simple 1D', () => {
  it('skips the legend top band when there is no legend', () => {
    const { options } = useLineChartOptions({
      chartData: ref({
        title: 'items',
        statType: 'counts',
        yAxis: [],
        zAxis: [],
        series: [
          { xAxis: 'A', values: [10], benchmarkId: 'a' },
          { xAxis: 'B', values: [20], benchmarkId: 'b' },
        ],
        points: [],
        axisLabels: { x: 'category' },
      }),
      sort: ref({ enabled: false, order: 'asc' as const }),
      showLabels: ref(false),
      isDark: ref(false),
      scale: ref<ScaleInput>('linear'),
      smooth: ref(false),
      stack: ref(false),
    })
    const grid = options.value.grid as {
      top?: number | string
      bottom?: number
      containLabel?: boolean
    }
    expect((options.value.legend as { show?: boolean }).show).toBe(false)
    expect(grid.top).toBe(VALUE_MODE_GRID_TOP)
    expect(grid.bottom).toBe(28)
    expect(grid.containLabel).toBe(true)
  })

  it('keeps the category slider band on large-X 1D lines with no legend top', () => {
    const { options } = useLineChartOptions({
      chartData: ref({
        title: 'items',
        statType: 'counts',
        yAxis: [],
        zAxis: [],
        series: Array.from({ length: LARGE_X_THRESHOLD + 1 }, (_, i) => ({
          xAxis: `x${i}`,
          values: [i],
          benchmarkId: `x${i}`,
        })),
        points: [],
        axisLabels: { x: 'category' },
      }),
      sort: ref({ enabled: false, order: 'asc' as const }),
      showLabels: ref(false),
      isDark: ref(false),
      scale: ref<ScaleInput>('linear'),
      smooth: ref(false),
      stack: ref(false),
    })
    const grid = options.value.grid as { top?: number; bottom?: number; containLabel?: boolean }
    expect(grid.top).toBe(VALUE_MODE_GRID_TOP)
    expect(grid.bottom).toBe(100)
    expect(grid.containLabel).toBe(false)
  })

  it('uses no-legend top on numeric log-X 1D lines', () => {
    const { options } = useLineChartOptions(
      makeNumericStepConfig(
        { type: 'log', axes: ['x'] },
        {
          title: 'steps',
          statType: 'grouped',
          yAxis: [],
          zAxis: [],
          series: [
            { xAxis: '1', values: [10], benchmarkId: '1' },
            { xAxis: '2', values: [0], benchmarkId: '2' },
          ],
          points: [],
          axisLabels: { x: 'step' },
        }
      )
    )
    expect((options.value.grid as { top?: number }).top).toBe(VALUE_MODE_GRID_TOP)
  })
})

describe('useLineChartOptions — grouped mode', () => {
  it('uses a compact pixel legend band, not the category % top', () => {
    const { options } = useLineChartOptions(makeGroupedConfig())
    const grid = options.value.grid as {
      top?: number | string
      bottom?: number
      containLabel?: boolean
    }
    expect(grid.top).toBe(48)
    expect(grid.bottom).toBe(28)
    expect(grid.containLabel).toBe(true)
  })

  it('grows the legend band when grouped series wrap', () => {
    const yAxis = Array.from({ length: 16 }, (_, i) => `s${i}`)
    const data = makeGroupedChartData()
    data.yAxis = yAxis
    data.series = data.series.map((s) => ({ ...s, values: yAxis.map(() => 1) }))
    const { options } = useLineChartOptions({
      ...makeGroupedConfig(),
      chartData: ref(data),
    })
    expect((options.value.grid as { top?: number }).top).toBe(48 + 16)
  })

  it('keeps the category slider band on large-X grouped lines with compact top', () => {
    const data = makeGroupedChartData()
    data.series = Array.from({ length: LARGE_X_THRESHOLD + 1 }, (_, i) => ({
      xAxis: `x${i}`,
      values: [10, 8],
      benchmarkId: `x${i}`,
    }))
    const { options } = useLineChartOptions({
      ...makeGroupedConfig(),
      chartData: ref(data),
    })
    const grid = options.value.grid as { top?: number; bottom?: number; containLabel?: boolean }
    expect(grid.top).toBe(48)
    expect(grid.bottom).toBe(100)
    expect(grid.containLabel).toBe(false)
  })

  it('emits stacked area line series when stack is enabled', () => {
    const { options } = useLineChartOptions(makeGroupedConfig({ stack: true }))
    const series = options.value.series as { stack?: string; areaStyle?: Record<string, never> }[]
    expect(series).toHaveLength(2)
    expect(series.every((s) => s.stack === 'total')).toBe(true)
    expect(series.every((s) => s.areaStyle !== undefined)).toBe(true)
  })

  it('does not stack grouped lines by default', () => {
    const { options } = useLineChartOptions(makeGroupedConfig())
    const series = options.value.series as { stack?: string | null; areaStyle?: unknown }[]
    expect(series.every((s) => s.stack === null)).toBe(true)
    expect(series.every((s) => s.areaStyle === null)).toBe(true)
  })
})

describe('useLineChartOptions — mixed mode', () => {
  it('emits line series with mixedTuples as data', () => {
    const { options } = useLineChartOptions(makeMixedConfig())
    const s = (options.value.series as { type: string; data: [number, number][] }[])[0]!
    expect(s.type).toBe('line')
    expect(s.data).toEqual([
      [0, 12],
      [1, 11],
    ])
  })

  it('emits smooth line series when enabled', () => {
    const { options } = useLineChartOptions(makeSmoothMixedConfig())
    const s = (options.value.series as { smooth?: boolean }[])[0]!
    expect(s.smooth).toBe(true)
  })

  it('uses axis trigger with themed snap line pointer for category x', () => {
    const { options } = useLineChartOptions(makeMixedConfig())
    const tooltip = options.value.tooltip as {
      trigger?: string
      axisPointer?: { type?: string; snap?: boolean; lineStyle?: { color?: string } }
    }
    expect(tooltip.trigger).toBe('axis')
    expect(tooltip.axisPointer?.type).toBe('line')
    expect(tooltip.axisPointer?.snap).toBe(true)
    expect(tooltip.axisPointer?.lineStyle?.color).toBe('#d1d5db')
  })
})

describe('useLineChartOptions — axes value mode', () => {
  it('uses cross axisPointer when both axes are value type', () => {
    const { options } = useLineChartOptions(makeValueConfig())
    expect((options.value.tooltip as { axisPointer?: { type?: string } }).axisPointer?.type).toBe(
      'cross'
    )
  })

  it('emits smooth line series when enabled', () => {
    const { options } = useLineChartOptions(makeSmoothValueConfig())
    const s = (options.value.series as { smooth?: boolean }[])[0]!
    expect(s.smooth).toBe(true)
  })
})

describe('useLineChartOptions — per-axis log scale', () => {
  it('string log logs the default value axis only', () => {
    const { options } = useLineChartOptions(makeNumericStepConfig('log'))
    const { xAxis, yAxis } = axisOf(options.value)
    expect(xAxis.type).toBe('category')
    expect(yAxis.type).toBe('log')
    expect(yAxis.logBase).toBe(10)
  })

  it('object axes x on numeric-step grouped lines sets log X and value Y with legend', () => {
    const { options } = useLineChartOptions(makeNumericStepConfig({ type: 'log', axes: ['x'] }))
    const { xAxis, yAxis } = axisOf(options.value)
    expect(xAxis.type).toBe('log')
    expect(yAxis.type).toBe('value')
    const legend = options.value.legend as { show?: boolean; data?: string[] }
    expect(legend.show).not.toBe(false)
    expect(legend.data).toEqual(['train', 'val'])
    const series = options.value.series as { name: string; data: [number, number | null][] }[]
    expect(series).toHaveLength(2)
    expect(series[0]!.data[0]).toEqual([1, 10])
  })

  it('base 2 sets xAxis.logBase to 2', () => {
    const { options } = useLineChartOptions(
      makeNumericStepConfig({ type: 'log', axes: ['x'], base: 2 })
    )
    expect(axisOf(options.value).xAxis.logBase).toBe(2)
  })

  it('X-log does not drop y <= 0', () => {
    const { options } = useLineChartOptions(makeNumericStepConfig({ type: 'log', axes: ['x'] }))
    const series = options.value.series as { data: [number, number | null][] }[]
    expect(series[0]!.data).toEqual([
      [1, 10],
      [2, 0],
      [4, 12],
    ])
  })

  it('Y-log still nulls y <= 0', () => {
    const { options } = useLineChartOptions(makeNumericStepConfig({ type: 'log', axes: ['y'] }))
    const series = options.value.series as { data: (number | null)[] }[]
    expect(series[0]!.data).toEqual([10, null, 12])
    expect(axisOf(options.value).xAxis.type).toBe('category')
    expect(axisOf(options.value).yAxis.type).toBe('log')
  })

  it('non-numeric category X stays category when axes includes x', () => {
    const { options } = useLineChartOptions({
      ...makeGroupedConfig(),
      scale: ref<ScaleInput>({ type: 'log', axes: ['x'] }),
    })
    const { xAxis, yAxis } = axisOf(options.value)
    expect(xAxis.type).toBe('category')
    expect(xAxis.data).toEqual(['Jan', 'Feb'])
    expect(yAxis.type).toBe('value')
  })

  it('uses a compact pixel legend band on numeric log-X, not the category % top', () => {
    const { options } = useLineChartOptions(makeNumericStepConfig({ type: 'log', axes: ['x'] }))
    const grid = options.value.grid as {
      top?: number | string
      bottom?: number
      containLabel?: boolean
    }
    expect(grid.top).toBe(48)
    expect(grid.bottom).toBe(28)
    expect(grid.containLabel).toBe(true)
  })

  it('does not reserve the category slider band when numeric log-X has many steps', () => {
    const many = Array.from({ length: LARGE_X_THRESHOLD + 1 }, (_, i) => [
      String(i + 1),
      [10, 8],
    ]) as [string, number[]][]
    const { options } = useLineChartOptions(
      makeNumericStepConfig(
        { type: 'log', axes: ['x'] },
        makeNumericStepChartData(['train', 'val'], many)
      )
    )
    const grid = options.value.grid as { top?: number; bottom?: number; containLabel?: boolean }
    expect(grid.top).toBe(48)
    expect(grid.bottom).toBe(28)
    expect(grid.containLabel).toBe(true)
  })

  it('grouped log-X with 2+ y series uses axis tooltip not cross', () => {
    const { options } = useLineChartOptions(makeNumericStepConfig({ type: 'log', axes: ['x'] }))
    const tooltip = tooltipOf(options.value)
    expect(tooltip.trigger).toBe('axis')
    expect(tooltip.axisPointer?.type).not.toBe('cross')
    expect(tooltip.axisPointer?.type).toBe('line')
    expect(tooltip.axisPointer?.snap).toBe(true)
    const html = tooltip.formatter?.([
      { name: '1', seriesName: 'train', value: [1, 10], marker: 'a', color: '#a00' },
      { name: '1', seriesName: 'val', value: [1, 8], marker: 'b', color: '#00a' },
    ])
    expect(html).toContain('train')
    expect(html).toContain('val')
    expect(html).toContain('10')
    expect(html).toContain('8')
  })

  it('same coerce-X path with one series uses cross', () => {
    const { options } = useLineChartOptions(
      makeNumericStepConfig(
        { type: 'log', axes: ['x'] },
        makeNumericStepChartData(
          ['train'],
          [
            ['1', [10]],
            ['2', [0]],
            ['4', [12]],
          ]
        )
      )
    )
    const tooltip = tooltipOf(options.value)
    expect(tooltip.axisPointer?.type).toBe('cross')
    expect(tooltip.trigger).toBe('item')
  })

  it('x-only numeric log-X uses cross and [x, y] pairs', () => {
    const { options } = useLineChartOptions(
      makeNumericStepConfig(
        { type: 'log', axes: ['x'] },
        {
          title: 'steps',
          statType: 'grouped',
          yAxis: [],
          zAxis: [],
          series: [
            { xAxis: '1', values: [10], benchmarkId: '1' },
            { xAxis: '2', values: [0], benchmarkId: '2' },
          ],
          points: [],
          axisLabels: { x: 'step' },
        }
      )
    )
    expect(axisOf(options.value).xAxis.type).toBe('log')
    const series = options.value.series as { data: [number, number | null][] }[]
    expect(series[0]!.data).toEqual([
      [1, 10],
      [2, 0],
    ])
    expect(tooltipOf(options.value).axisPointer?.type).toBe('cross')
  })
})

describe('useLineChartOptions — grouped mode', () => {
  it('preserves straight lines by default', () => {
    const { options } = useLineChartOptions(makeGroupedConfig())
    const series = options.value.series as { smooth?: boolean }[]
    expect(series.every((s) => s.smooth === false)).toBe(true)
  })

  it('emits smooth on every grouped line series when enabled', () => {
    const { options } = useLineChartOptions(makeGroupedConfig({ smooth: true }))
    const series = options.value.series as { smooth?: boolean }[]
    expect(series).toHaveLength(2)
    expect(series.every((s) => s.smooth === true)).toBe(true)
  })
})
