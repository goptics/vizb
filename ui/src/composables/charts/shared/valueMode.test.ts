import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import { ref } from 'vue'
import {
  baseConfig,
  makeValueChartData,
  installDevicePixelRatio,
  type BaseConfigOverrides,
} from '@/test-utils'
import { LARGE_X_THRESHOLD } from './chartConfig'
import { sortValueTuples, scaleValueTuples, buildValueAxes2DOptions } from './valueMode'

let restoreDpr: () => void
beforeAll(() => {
  restoreDpr = installDevicePixelRatio(1)
})
afterAll(() => {
  restoreDpr()
})

describe('sortValueTuples', () => {
  const tuples: [number, number, number?][] = [
    [1, 30],
    [2, 10],
    [3, 20],
  ]

  it('returns original when sort disabled', () => {
    expect(sortValueTuples(tuples, false, 'asc')).toBe(tuples)
  })

  it('sorts by y ascending and descending', () => {
    expect(sortValueTuples(tuples, true, 'asc').map((t) => t[1])).toEqual([10, 20, 30])
    expect(sortValueTuples(tuples, true, 'desc').map((t) => t[1])).toEqual([30, 20, 10])
  })
})

describe('scaleValueTuples', () => {
  it('returns original on linear', () => {
    const tuples: [number, number, number?][] = [
      [1, 0],
      [2, 5],
    ]
    expect(scaleValueTuples(tuples, 'linear')).toBe(tuples)
  })

  it('nulls non-positive y on log and keeps color dim', () => {
    const tuples: [number, number, number?][] = [
      [1, 0, 9],
      [2, -1, 8],
      [3, 4, 7],
    ]
    expect(scaleValueTuples(tuples, 'log')).toEqual([
      [1, null, 9],
      [2, null, 8],
      [3, 4, 7],
    ])
  })

  it('x-log drops non-positive x and leaves y intact', () => {
    const tuples: [number, number, number?][] = [
      [0, 0],
      [2, -1],
    ]
    expect(scaleValueTuples(tuples, 'linear', 'log')).toEqual([[2, -1]])
  })
})

function cfg(overrides: BaseConfigOverrides = {}) {
  return baseConfig({
    chartData: makeValueChartData(),
    ...overrides,
  })
}

function seriesOf(option: ReturnType<typeof buildValueAxes2DOptions>) {
  const series = option.series
  if (!Array.isArray(series) || !series[0] || typeof series[0] !== 'object') {
    throw new Error('expected series[0]')
  }
  return series[0] as {
    type: string
    data: unknown
    smooth?: boolean
    large?: boolean
    largeThreshold?: number
    symbol?: string
    symbolSize?: number
    sampling?: string
    itemStyle?: { color?: string }
    label?: { formatter?: (p: { data: [number, number | null, number?] }) => string }
  }
}

describe('buildValueAxes2DOptions', () => {
  it('defaults to scatter and maps valueTuples', () => {
    const option = buildValueAxes2DOptions(cfg())
    const s = seriesOf(option)
    expect(s.type).toBe('scatter')
    expect(s.data).toEqual([
      [100, 12],
      [200, 8],
    ])
    expect(option.legend).toEqual({ show: false })
    expect(option.dataZoom).toEqual([
      { type: 'inside', xAxisIndex: 0 },
      { type: 'inside', yAxisIndex: 0 },
    ])
  })

  it('defaults scale to linear when scale ref omitted', () => {
    const full = cfg()
    const { scale: _scale, ...rest } = full
    const option = buildValueAxes2DOptions(rest)
    expect(seriesOf(option).data).toEqual([
      [100, 12],
      [200, 8],
    ])
  })

  it('uses empty tuples when valueTuples missing', () => {
    const option = buildValueAxes2DOptions(
      cfg({ chartData: makeValueChartData({ valueTuples: undefined }) })
    )
    expect(seriesOf(option).data).toEqual([])
  })

  it('logs X independently and keeps non-positive y', () => {
    const option = buildValueAxes2DOptions(
      cfg({
        chartData: makeValueChartData({
          valueTuples: [
            [10, 0],
            [100, 8],
          ],
        }),
        scale: { type: 'log', axes: ['x'], base: 2 },
      }),
      'line'
    )
    expect((option.xAxis as { type?: string; logBase?: number }).type).toBe('log')
    expect((option.xAxis as { logBase?: number }).logBase).toBe(2)
    expect((option.yAxis as { type?: string }).type).toBe('value')
    expect(seriesOf(option).data).toEqual([
      [10, 0],
      [100, 8],
    ])
  })

  it('sorts when enabled and scales log y', () => {
    const option = buildValueAxes2DOptions(
      cfg({
        chartData: makeValueChartData({
          valueTuples: [
            [1, 0],
            [2, 20],
            [3, 5],
          ],
        }),
        sort: { enabled: true, order: 'desc' },
        scale: 'log',
      }),
      'line'
    )
    expect(seriesOf(option).data).toEqual([
      [2, 20],
      [3, 5],
      [1, null],
    ])
  })

  it('sorts ascending when enabled', () => {
    const option = buildValueAxes2DOptions(
      cfg({
        chartData: makeValueChartData({
          valueTuples: [
            [1, 30],
            [2, 10],
          ],
        }),
        sort: { enabled: true, order: 'asc' },
      })
    )
    expect(seriesOf(option).data).toEqual([
      [2, 10],
      [1, 30],
    ])
  })

  it('maps unknown chart types to scatter and returns empty seriesSymbol for bar', () => {
    const pie = buildValueAxes2DOptions(cfg(), 'pie')
    expect(seriesOf(pie).type).toBe('scatter')

    const bar = buildValueAxes2DOptions(cfg(), 'bar')
    const s = seriesOf(bar)
    expect(s.type).toBe('bar')
    expect(s.large).toBe(true)
    expect(s.largeThreshold).toBe(2000)
    expect(s.symbol).toBeUndefined()
  })

  it('applies large scatter/line symbols when point count exceeds threshold', () => {
    const many = Array.from(
      { length: LARGE_X_THRESHOLD + 1 },
      (_, i) => [i, i + 1] as [number, number]
    )
    const scatter = buildValueAxes2DOptions(
      cfg({ chartData: makeValueChartData({ valueTuples: many }) }),
      'scatter'
    )
    expect(seriesOf(scatter).symbolSize).toBe(5)

    const line = buildValueAxes2DOptions(
      cfg({ chartData: makeValueChartData({ valueTuples: many }) }),
      'line'
    )
    const ls = seriesOf(line)
    expect(ls.symbol).toBe('none')
    expect(ls.sampling).toBe('lttb')
  })

  it('uses default symbols for small scatter/line and honors overrides', () => {
    const option = buildValueAxes2DOptions(
      {
        ...cfg(),
        symbol: ref('diamond'),
        symbolSize: ref(15),
      },
      'scatter'
    )
    const s = seriesOf(option)
    expect(s.symbol).toBe('diamond')
    expect(s.symbolSize).toBe(15)

    const line = buildValueAxes2DOptions(cfg(), 'line')
    expect(seriesOf(line).symbolSize).toBe(7)
    expect(line.dataZoom).toBeUndefined()
  })

  it('keeps inside zoom on scatter and omits it on value-mode line', () => {
    expect(buildValueAxes2DOptions(cfg(), 'scatter').dataZoom).toEqual([
      { type: 'inside', xAxisIndex: 0 },
      { type: 'inside', yAxisIndex: 0 },
    ])
    expect(buildValueAxes2DOptions(cfg(), 'line').dataZoom).toBeUndefined()
  })

  it('enables smooth lines and visualMap color dimension from third tuple slot', () => {
    const option = buildValueAxes2DOptions(
      cfg({
        chartData: makeValueChartData({
          valueTuples: [
            [1, 2, 9],
            [3, 4, 1],
          ],
        }),
        visualMap: true,
        smooth: true,
      }),
      'scatter'
    )
    const s = seriesOf(option)
    expect(s.itemStyle).toBeUndefined()
    expect(s.large).toBe(false)
    expect(option.visualMap).toBeTruthy()

    const line = buildValueAxes2DOptions(cfg({ smooth: true }), 'line')
    expect(seriesOf(line).smooth).toBe(true)
  })

  it('formats labels for number, null, and undefined y', () => {
    const option = buildValueAxes2DOptions(cfg({ showLabels: true }))
    const formatter = seriesOf(option).label?.formatter
    expect(formatter).toBeTypeOf('function')
    expect(formatter!({ data: [1, 1.234] })).toBe('1.234')
    expect(formatter!({ data: [1, null] })).toBe('')
    expect(formatter!({ data: [1, undefined as unknown as number] })).toBe('')
  })

  it('uses cross pointer for scatter/line and color dim 1 without third slot', () => {
    const scatter = buildValueAxes2DOptions(cfg({ visualMap: true }), 'scatter')
    expect(scatter.tooltip).toBeTruthy()
    const s = seriesOf(scatter)
    expect(s.itemStyle).toBeUndefined()

    const bar = buildValueAxes2DOptions(cfg(), 'bar')
    expect(seriesOf(bar).itemStyle?.color).toBeTruthy()
  })
})
