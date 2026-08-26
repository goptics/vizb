import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import { ref } from 'vue'
import {
  baseConfig,
  makeMixedChartData,
  installDevicePixelRatio,
  type BaseConfigOverrides,
} from '@/test-utils'
import { LARGE_X_THRESHOLD } from './chartConfig'
import {
  scaleMixedTuples,
  createMixedModeTooltip,
  buildMixedAxes2DOptions,
  buildMixedAxes3DOptions,
} from './mixedMode'

let restoreDpr: () => void
beforeAll(() => {
  restoreDpr = installDevicePixelRatio(1)
})
afterAll(() => {
  restoreDpr()
})

describe('scaleMixedTuples', () => {
  it('returns original on linear', () => {
    const tuples: [number, number][] = [
      [0, 1],
      [1, 2],
    ]
    expect(scaleMixedTuples(tuples, 'linear')).toBe(tuples)
  })

  it('nulls non-positive y on log', () => {
    expect(
      scaleMixedTuples(
        [
          [0, 0],
          [1, -2],
          [2, 5],
        ],
        'log'
      )
    ).toEqual([
      [0, null],
      [1, null],
      [2, 5],
    ])
  })
})

describe('createMixedModeTooltip', () => {
  const cats = ['West', 'South', 'East']

  function fmt(tooltip: ReturnType<typeof createMixedModeTooltip>, params: unknown): string {
    if (!tooltip || typeof tooltip !== 'object' || !('formatter' in tooltip)) {
      throw new Error('missing formatter')
    }
    const formatter = tooltip.formatter
    if (typeof formatter !== 'function') throw new Error('formatter not fn')
    return formatter(params as never, '' as never) as string
  }

  it('uses line pointer for scatter/line and shadow for bar', () => {
    const line = createMixedModeTooltip(false, cats, 'line', 'region', 'tax')
    const bar = createMixedModeTooltip(true, cats, 'bar', 'region', 'tax')
    expect(line && typeof line === 'object' && 'axisPointer' in line).toBe(true)
    expect(bar && typeof bar === 'object' && 'axisPointer' in bar).toBe(true)
  })

  it('defaults axis names and returns empty for empty params', () => {
    const tooltip = createMixedModeTooltip(false, cats, 'scatter')
    expect(fmt(tooltip, [])).toBe('')
  })

  it('formats from name + numeric value', () => {
    const tooltip = createMixedModeTooltip(false, cats, 'bar', 'X', 'Y')
    const html = fmt(tooltip, { name: 'West', value: 12.345, marker: '•' })
    expect(html).toContain('X: West')
    expect(html).toContain('Y:')
    expect(html).toContain('•')
  })

  it('formats from array value and single non-array params', () => {
    const tooltip = createMixedModeTooltip(false, cats, 'line', 'X', 'Y')
    expect(fmt(tooltip, { name: 'South', value: [1, 3.5], marker: '' })).toContain('3.5')
  })

  it('falls back to data index category and empty y text', () => {
    const tooltip = createMixedModeTooltip(false, cats, 'scatter', 'X', 'Y')
    const fromData = fmt(tooltip, { data: [2, null] })
    expect(fromData).toContain('East')
    expect(fromData).toMatch(/Y: <b><\/b>/)

    const unknownIdx = fmt(tooltip, { data: [99, undefined] })
    expect(unknownIdx).toContain('99')

    const noIdx = fmt(tooltip, { data: undefined, value: undefined })
    expect(noIdx).toContain('X: ')
  })

  it('uses data[1] when value is neither array nor number', () => {
    const tooltip = createMixedModeTooltip(false, cats, 'bar', 'X', 'Y')
    const html = fmt(tooltip, {
      name: 'West',
      value: 'x' as unknown as number,
      data: [0, 42],
    })
    expect(html).toContain('42')
  })
})

function cfg(overrides: BaseConfigOverrides = {}) {
  return baseConfig({
    chartData: makeMixedChartData(),
    ...overrides,
  })
}

function series0(option: ReturnType<typeof buildMixedAxes2DOptions>) {
  const series = option.series
  if (!Array.isArray(series) || !series[0] || typeof series[0] !== 'object') {
    throw new Error('expected series')
  }
  return series[0] as {
    type: string
    data: unknown
    smooth?: boolean
    large?: boolean
    symbol?: string
    symbolSize?: number
    sampling?: string
    itemStyle?: { color?: string }
    label?: { formatter?: (p: { data: [number, number | null] }) => string }
  }
}

describe('buildMixedAxes2DOptions', () => {
  it('defaults to scatter series with mixedTuples', () => {
    const option = buildMixedAxes2DOptions(cfg())
    const s = series0(option)
    expect(s.type).toBe('scatter')
    expect(s.data).toEqual([
      [0, 1926.35],
      [1, 447.38],
    ])
    expect(option.legend).toEqual({ show: false })
  })

  it('coerces numeric category X to log and uses cross tooltip', () => {
    const option = buildMixedAxes2DOptions(
      cfg({
        scale: { type: 'log', axes: ['x'], base: 2 },
        chartData: makeMixedChartData({
          xCategories: ['1', '10'],
          mixedTuples: [
            [0, 5],
            [1, 0],
          ],
        }),
      }),
      'line'
    )
    expect((option.xAxis as { type?: string; logBase?: number }).type).toBe('log')
    expect((option.xAxis as { logBase?: number }).logBase).toBe(2)
    expect((option.yAxis as { type?: string }).type).toBe('value')
    expect(series0(option).data).toEqual([
      [1, 5],
      [10, 0],
    ])
    expect((option.tooltip as { axisPointer?: { type?: string } }).axisPointer?.type).toBe('cross')
  })

  it('falls back to the mixed index when log-X lookup is out of range', () => {
    const option = buildMixedAxes2DOptions(
      cfg({
        scale: { type: 'log', axes: ['x'] },
        chartData: makeMixedChartData({
          xCategories: ['1'],
          mixedTuples: [
            [0, 5],
            [9, 8],
          ],
        }),
      }),
      'line'
    )
    expect(series0(option).data).toEqual([
      [1, 5],
      [9, 8],
    ])
  })

  it('defaults scale to linear when scale ref omitted', () => {
    const full = cfg()
    const { scale: _scale, ...rest } = full
    const option = buildMixedAxes2DOptions(rest)
    expect(series0(option).type).toBe('scatter')
  })

  it('handles missing mixedTuples/xCategories', () => {
    const option = buildMixedAxes2DOptions(
      cfg({
        chartData: makeMixedChartData({
          mixedTuples: undefined,
          xCategories: undefined,
        }),
      })
    )
    expect(series0(option).data).toEqual([])
  })

  it('builds bar and line variants with smooth + symbols', () => {
    const bar = buildMixedAxes2DOptions(cfg(), 'bar')
    expect(series0(bar).type).toBe('bar')
    expect(series0(bar).large).toBe(true)

    const line = buildMixedAxes2DOptions(cfg({ smooth: true }), 'line')
    expect(series0(line).type).toBe('line')
    expect(series0(line).smooth).toBe(true)
    expect(series0(line).symbolSize).toBe(7)
  })

  it('uses large symbols and dataZoom when categories exceed threshold', () => {
    const cats = Array.from({ length: LARGE_X_THRESHOLD + 1 }, (_, i) => `c${i}`)
    const tuples = cats.map((_, i) => [i, i + 1] as [number, number])
    const option = buildMixedAxes2DOptions(
      cfg({
        chartData: makeMixedChartData({ xCategories: cats, mixedTuples: tuples }),
      }),
      'line'
    )
    const s = series0(option)
    expect(s.symbol).toBe('none')
    expect(s.sampling).toBe('lttb')
    const dataZoom = option.dataZoom as { type: string }[]
    expect(dataZoom).toHaveLength(1)
    expect(dataZoom[0]!.type).toBe('slider')

    const scatter = buildMixedAxes2DOptions(
      cfg({
        chartData: makeMixedChartData({ xCategories: cats, mixedTuples: tuples }),
      }),
      'scatter'
    )
    expect(series0(scatter).symbolSize).toBe(5)
  })

  it('omits itemStyle color when visualMap on scatter', () => {
    const option = buildMixedAxes2DOptions(cfg({ visualMap: true }), 'scatter')
    expect(series0(option).itemStyle).toBeUndefined()
    expect(option.visualMap).toBeTruthy()
  })

  it('applies log scale and symbol overrides', () => {
    const option = buildMixedAxes2DOptions(
      {
        ...cfg({
          chartData: makeMixedChartData({
            mixedTuples: [
              [0, 0],
              [1, 10],
            ],
          }),
          scale: 'log',
        }),
        symbol: ref('triangle'),
        symbolSize: ref(11),
      },
      'scatter'
    )
    expect(series0(option).data).toEqual([
      [0, null],
      [1, 10],
    ])
    expect(series0(option).symbol).toBe('triangle')
    expect(series0(option).symbolSize).toBe(11)
  })

  it('formats labels for number/null/undefined', () => {
    const option = buildMixedAxes2DOptions(cfg({ showLabels: true }))
    const formatter = series0(option).label?.formatter
    expect(formatter!({ data: [0, 2.456] })).toBe('2.456')
    expect(formatter!({ data: [0, null] })).toBe('')
    expect(formatter!({ data: [0, undefined as unknown as number] })).toBe('')
  })

  it('uses inside zoom for small category axes', () => {
    const option = buildMixedAxes2DOptions(cfg())
    expect(option.dataZoom).toEqual([
      { type: 'inside', xAxisIndex: 0 },
      { type: 'inside', yAxisIndex: 0 },
    ])
  })

  it('omits inside zoom for small mixed line axes', () => {
    expect(buildMixedAxes2DOptions(cfg(), 'line').dataZoom).toBeUndefined()
  })

  it('does not reserve the category slider band for mixed scatter log-X', () => {
    const cats = Array.from({ length: LARGE_X_THRESHOLD + 1 }, (_, i) => String(i + 1))
    const tuples = cats.map((_, i) => [i, i + 1] as [number, number])
    const option = buildMixedAxes2DOptions(
      cfg({
        scale: { type: 'log', axes: ['x'] },
        chartData: makeMixedChartData({ xCategories: cats, mixedTuples: tuples }),
      }),
      'scatter'
    )
    const grid = option.grid as { bottom?: number; containLabel?: boolean }
    expect(grid.bottom).toBe(28)
    expect(grid.containLabel).toBe(true)
    expect(option.dataZoom).toEqual([
      { type: 'inside', xAxisIndex: 0 },
      { type: 'inside', yAxisIndex: 0 },
    ])
  })
})

describe('buildMixedAxes3DOptions', () => {
  function mixed3dConfig(overrides: BaseConfigOverrides = {}) {
    return baseConfig({
      chartData: makeMixedChartData({
        axisLabels: { x: 'region', y: 'tax', z: 'score' },
        render3D: {
          mode: 'mixed',
          xValues: ['West', 'South'],
          yValues: [],
          zValues: [],
          barSeries: [],
          lineSeries: [
            {
              name: 'pts',
              data: [{ value: [0, 1, 2] }, { value: [1, 3, 4] }],
            },
          ],
          cellTotals: {},
        },
      }),
      ...overrides,
    })
  }

  function seriesList(option: ReturnType<typeof buildMixedAxes3DOptions>) {
    const series = option.series
    if (!Array.isArray(series)) throw new Error('series')
    return series as Array<Record<string, unknown>>
  }

  it('defaults to scatter3D with symbol props', () => {
    const option = buildMixedAxes3DOptions({
      ...mixed3dConfig(),
      symbol: ref('diamond'),
      symbolSize: ref(6),
    })
    const s = seriesList(option)[0]!
    expect(s.type).toBe('scatter3D')
    expect(s.symbol).toBe('diamond')
    expect(s.symbolSize).toBe(6)
    expect(s.itemStyle).toBeTruthy()
    expect(option.legend).toEqual({ show: false })
    expect(option.xAxis3D).toMatchObject({ type: 'category', data: ['West', 'South'] })
    expect(option.yAxis3D).toMatchObject({ type: 'value' })
    expect(option.zAxis3D).toMatchObject({ type: 'value' })
  })

  it('defaults scale/threeDRotate when optional refs omitted', () => {
    const full = mixed3dConfig()
    const { scale: _s, threeDRotate: _r, ...rest } = full
    const option = buildMixedAxes3DOptions(rest)
    expect(seriesList(option)[0]!.type).toBe('scatter3D')
    expect(option.yAxis3D).toMatchObject({ type: 'value' })
  })
  it('builds bar3D and line3D series', () => {
    const bar = buildMixedAxes3DOptions(mixed3dConfig(), 'bar3D')
    expect(seriesList(bar)[0]).toMatchObject({
      type: 'bar3D',
      bevelSize: 0.3,
      shading: 'lambert',
    })

    const line = buildMixedAxes3DOptions(mixed3dConfig(), 'line3D')
    expect(seriesList(line)[0]).toMatchObject({
      type: 'line3D',
      lineStyle: { width: 3 },
    })
  })

  it('enables visualMap and log axes, empty series data path', () => {
    const option = buildMixedAxes3DOptions(
      mixed3dConfig({
        threeDVisualMap: true,
        threeDRotate: true,
        scale: 'log',
        chartData: makeMixedChartData({
          render3D: {
            mode: 'mixed',
            xValues: ['a'],
            yValues: [],
            zValues: [],
            barSeries: [],
            lineSeries: [],
            cellTotals: {},
          },
        }),
      }),
      'scatter3D'
    )
    expect(option.yAxis3D).toMatchObject({ type: 'log' })
    expect(option.zAxis3D).toMatchObject({ type: 'log' })
    expect(option.yAxis3D).toMatchObject({ logBase: 10 })
    expect(seriesList(option)).toEqual([])
    expect(option.visualMap).toBeTruthy()
  })

  it('logs mixed 3D y only when axes is y', () => {
    const option = buildMixedAxes3DOptions(
      mixed3dConfig({ scale: { type: 'log', axes: ['y'] } }),
      'scatter3D'
    )
    expect(option.yAxis3D).toMatchObject({ type: 'log' })
    expect(option.zAxis3D).toMatchObject({ type: 'value' })
    expect(option.xAxis3D).toMatchObject({ type: 'category' })
  })

  it('omits itemStyle on scatter/bar when visualMap on', () => {
    const scatter = buildMixedAxes3DOptions(mixed3dConfig({ threeDVisualMap: true }), 'scatter3D')
    expect(seriesList(scatter)[0]!.itemStyle).toBeUndefined()

    const bar = buildMixedAxes3DOptions(mixed3dConfig({ threeDVisualMap: true }), 'bar3D')
    expect(seriesList(bar)[0]!.itemStyle).toBeUndefined()
  })

  it('handles empty first series data and default chart type', () => {
    const option = buildMixedAxes3DOptions(
      mixed3dConfig({
        chartData: makeMixedChartData({
          render3D: {
            mode: 'mixed',
            xValues: [],
            yValues: [],
            zValues: [],
            barSeries: [],
            lineSeries: [{ name: 'empty', data: [] }],
            cellTotals: {},
          },
        }),
      })
    )
    expect(seriesList(option)[0]!.type).toBe('scatter3D')
  })
})
