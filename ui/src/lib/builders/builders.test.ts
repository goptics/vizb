import { describe, it, expect } from 'vitest'
import type { ChartData, DataPoint, Axis } from '@/types'
import { noSort, ascSort, descSort, dp } from '@/test-utils'
import { builderForChart, pickBuilder, grouped, preserveRows, value, mixed } from './index'
import { builderStatType } from './types'
import { finalizeChart } from './finalize'
import type { BuildContext } from './types'
import { buildValueModeChart, buildMixedModeChart } from '../transform'

const baseCtx = (overrides: Partial<BuildContext> = {}): BuildContext => ({
  signature: 'val-',
  statTemplate: { type: 'val' },
  sort: noSort,
  showLabels: false,
  scale: 'linear',
  threeD: false,
  preserveRows: false,
  ...overrides,
})

const emptyChart = (partial: Partial<ChartData> = {}): ChartData => ({
  title: 't',
  statType: 'grouped',
  yAxis: [],
  zAxis: [],
  series: [],
  points: [],
  ...partial,
})

describe('builderStatType', () => {
  it('returns chart.statType or grouped default', () => {
    expect(builderStatType(emptyChart({ statType: 'value' }))).toBe('value')
    expect(builderStatType(emptyChart({ statType: undefined as unknown as string }))).toBe(
      'grouped'
    )
  })
})

describe('pickBuilder / builderForChart', () => {
  it('pickBuilder respects mode flags in priority order', () => {
    expect(pickBuilder({ valueMode: true, mixedMode: true, preserveRows: true })).toBe(value)
    expect(pickBuilder({ mixedMode: true, preserveRows: true })).toBe(mixed)
    expect(pickBuilder({ preserveRows: true })).toBe(preserveRows)
    expect(pickBuilder({})).toBe(grouped)
  })

  it('builderForChart prefers data shape then render3D then statType', () => {
    expect(
      builderForChart(emptyChart({ series: [{ xAxis: 'A', values: [1], benchmarkId: '' }] }))
    ).toBe(grouped)
    expect(
      builderForChart(emptyChart({ points: [{ xAxis: 'A', yAxis: '', zAxis: '', value: 1 }] }))
    ).toBe(grouped)
    expect(builderForChart(emptyChart({ valueTuples: [[1, 2]] }))).toBe(value)
    expect(builderForChart(emptyChart({ valuePoints3D: [[1, 2, 3]] }))).toBe(value)
    expect(builderForChart(emptyChart({ mixedTuples: [[0, 1]], xCategories: ['A'] }))).toBe(mixed)
    expect(builderForChart(emptyChart({ xCategories: ['A'] }))).toBe(mixed)

    expect(
      builderForChart(
        emptyChart({
          render3D: {
            mode: 'continuous',
            xValues: [],
            yValues: [],
            zValues: [],
            barSeries: [],
            lineSeries: [],
            cellTotals: {},
          },
        })
      )
    ).toBe(value)
    expect(
      builderForChart(
        emptyChart({
          render3D: {
            mode: 'mixed',
            xValues: [],
            yValues: [],
            zValues: [],
            barSeries: [],
            lineSeries: [],
            cellTotals: {},
          },
        })
      )
    ).toBe(mixed)

    expect(builderForChart(emptyChart({ statType: 'value' }))).toBe(value)
    expect(builderForChart(emptyChart({ statType: 'mixed' }))).toBe(mixed)
    expect(builderForChart(emptyChart({ statType: 'preserveRows' }))).toBe(preserveRows)
    expect(builderForChart(emptyChart({ statType: 'other' }))).toBe(grouped)
  })
})

describe('ValueBuilder', () => {
  const valueAxes: Axis[] = [
    { key: 'x', label: 'px', type: 'value' },
    { key: 'y', label: 'py', type: 'value' },
  ]
  const valueAxes3: Axis[] = [...valueAxes, { key: 'z', label: 'pz', type: 'value' }]

  it('queries 2D valueTuples with metric color and off-chart z color', () => {
    const withMetric: DataPoint[] = [{ xAxis: '1', yAxis: '2', metric: '9', stats: [] }]
    const chart = buildValueModeChart(withMetric, valueAxes, 'xy', 'xy', { threeD: false })
    expect(chart.valueTuples).toEqual([[1, 2, 9]])
    expect(value.badgeCount(chart, 'x')).toBe(1)
    expect(value.badgeCount(chart, 'y')).toBe(1)
    expect(value.badgeCount(chart, 'z')).toBe(0)
    expect(value.grandTotal(chart)).toBe(2)
    expect(value.is3D(chart)).toBe(false)
    expect(value.canOfferValue3D()).toBe(false)

    const withZ: DataPoint[] = [{ xAxis: '1', yAxis: '2', zAxis: '3', stats: [] }]
    const chartZ = buildValueModeChart(withZ, valueAxes3, 'xyz', 'xyn', { threeD: false })
    expect(chartZ.valueTuples).toEqual([[1, 2, 3]])
  })

  it('skips non-finite / log-invalid rows and empty metric', () => {
    const data: DataPoint[] = [
      { xAxis: 'bad', yAxis: '1', stats: [] },
      { xAxis: '1', yAxis: '0', stats: [] },
      { xAxis: '2', yAxis: '3', metric: '', stats: [] },
      { xAxis: '4', yAxis: '5', metric: 'nope', stats: [] },
    ]
    const chart = buildValueModeChart(data, valueAxes, 'xy', 'xy', {
      scale: 'log',
      threeD: false,
    })
    expect(chart.valueTuples).toEqual([
      [2, 3],
      [4, 5],
    ])
  })

  it('queries continuous 3D with and without metric; log filters non-positive', () => {
    const data: DataPoint[] = [
      { xAxis: '1', yAxis: '2', zAxis: '3', metric: '4', stats: [] },
      { xAxis: '1', yAxis: '2', zAxis: '3', stats: [] },
      { xAxis: '0', yAxis: '2', zAxis: '3', stats: [] },
    ]
    const chart = buildValueModeChart(data, valueAxes3, 'xyz', 'xyz', {
      threeD: true,
      showLabels: true,
    })
    expect(chart.valuePoints3D).toEqual([
      [1, 2, 3, 4],
      [1, 2, 3],
      [0, 2, 3],
    ])
    expect(chart.render3D?.mode).toBe('continuous')
    expect(value.badgeCount(chart, 'x')).toBe(2)
    expect(value.badgeCount(chart, 'z')).toBe(1)
    expect(value.grandTotal(chart)).toBe(3 + 3 + 3)
    expect(value.is3D(chart)).toBe(true)

    const logChart = buildValueModeChart(
      [{ xAxis: '1', yAxis: '2', zAxis: '0', stats: [] }],
      valueAxes3,
      'xyz',
      'xyz',
      { threeD: true, scale: 'log' }
    )
    expect(logChart.valuePoints3D).toBeUndefined()
    expect(logChart.render3D).toBeUndefined()
  })

  it('handles missing axes/identity and empty charts in query methods', () => {
    const chart = buildValueModeChart([], [])
    expect(chart.title).toBe('x vs y')
    expect(chart.valueTuples).toEqual([])
    expect(value.badgeCount(emptyChart(), 'x')).toBe(0)
    expect(value.grandTotal(emptyChart())).toBe(0)
  })

  it('uses identityStringFromAxes when identityString omitted', () => {
    const chart = buildValueModeChart([{ xAxis: '10', yAxis: '20', stats: [] }], valueAxes)
    expect(chart.valueTuples).toEqual([[10, 20]])
  })
})

describe('MixedBuilder', () => {
  const axes2: Axis[] = [
    { key: 'x', label: 'region' },
    { key: 'y', label: 'latency', type: 'value' },
  ]
  const axes3: Axis[] = [...axes2, { key: 'z', label: 'sales', type: 'value' }]

  it('queries 2D mixedTuples', () => {
    const data: DataPoint[] = [
      { xAxis: 'Asia', yAxis: '12', stats: [] },
      { xAxis: '', yAxis: '1', stats: [] },
      { xAxis: 'EU', yAxis: 'bad', stats: [] },
      { xAxis: 'Asia', yAxis: '10', stats: [] },
    ]
    const chart = buildMixedModeChart(data, axes2)
    expect(chart.mixedTuples).toEqual([
      [0, 12],
      [0, 10],
    ])
    expect(chart.xCategories).toEqual(['Asia'])
    expect(mixed.badgeCount(chart, 'x')).toBe(1)
    expect(mixed.badgeCount(chart, 'y')).toBe(2)
    expect(mixed.badgeCount(chart, 'z')).toBe(0)
    expect(mixed.grandTotal(chart)).toBe(22)
    expect(mixed.is3D(chart)).toBe(false)
    expect(mixed.canOfferValue3D()).toBe(false)
  })

  it('log scale drops non-positive y in 2D', () => {
    const chart = buildMixedModeChart(
      [
        { xAxis: 'A', yAxis: '5', stats: [] },
        { xAxis: 'B', yAxis: '0', stats: [] },
      ],
      axes2,
      { scale: 'log' }
    )
    expect(chart.mixedTuples).toEqual([[0, 5]])
  })
  it('queries mixed 3D render and fallbacks', () => {
    const data: DataPoint[] = [
      { xAxis: 'Asia', yAxis: '12', zAxis: '100', stats: [] },
      { xAxis: 'EU', yAxis: '11', zAxis: 'bad', stats: [] },
      { xAxis: 'EU', yAxis: '0', zAxis: '1', stats: [] },
    ]
    const chart = buildMixedModeChart(data, axes3, { scale: 'log' })
    // log drops y<=0; non-finite z dropped — Asia kept; EU still registered in xCategories
    expect(chart.render3D?.mode).toBe('mixed')
    expect(chart.render3D?.lineSeries[0]?.data).toEqual([{ value: [0, 12, 100] }])
    expect(mixed.badgeCount(chart, 'x')).toBe(chart.xCategories?.length ?? 0)
    expect(mixed.badgeCount(chart, 'y')).toBe(1)
    expect(mixed.badgeCount(chart, 'z')).toBe(0)
    expect(mixed.grandTotal(chart)).toBe(12)
    expect(mixed.is3D(chart)).toBe(true)
    const empty = buildMixedModeChart([], axes3)
    expect(mixed.grandTotal(empty)).toBe(0)
    expect(mixed.badgeCount(emptyChart({ xCategories: ['A'] }), 'x')).toBe(1)
    expect(mixed.badgeCount(emptyChart(), 'y')).toBe(0)
  })

  it('falls back to default axis labels', () => {
    const chart = buildMixedModeChart(
      [{ xAxis: 'A', yAxis: '1', stats: [] }],
      [
        { key: 'x', type: 'category' as never },
        { key: 'y', type: 'value' },
      ]
    )
    expect(chart.title).toBe('x vs y')
  })
})

describe('PreserveRowsBuilder', () => {
  it('category scatter path when y is empty', () => {
    const data: DataPoint[] = [
      dp('Asia', '', '', 'val', 12),
      { xAxis: 'Asia', yAxis: '', stats: [{ type: 'val' }] }, // undefined value skipped
      dp('EU', '', '', 'val', 8),
    ]
    const chart = preserveRows.build(data, baseCtx({ signature: 'val-', preserveRows: true }))
    expect(chart.mixedTuples).toEqual([
      [0, 12],
      [1, 8],
    ])
    expect(preserveRows.badgeCount(chart, 'x')).toBe(2)
    expect(preserveRows.badgeCount(chart, 'y')).toBe(2)
    expect(preserveRows.badgeCount(chart, 'z')).toBe(0)
    expect(preserveRows.grandTotal(chart)).toBe(20)
    expect(preserveRows.canOfferValue3D()).toBe(false)
  })

  it('series path when y categories are present', () => {
    const data: DataPoint[] = [
      dp('West', 'Hardware', '', 'val', 10),
      dp('West', 'Mechanical', '', 'val', 20),
      dp('East', 'Hardware', '', 'val', 30),
      { xAxis: 'East', yAxis: 'Hardware', stats: [{ type: 'val' }] },
    ]
    const chart = preserveRows.build(
      data,
      baseCtx({ signature: 'val-', preserveRows: true, labels: { x: 'region', y: 'cat' } })
    )
    expect(chart.series).toHaveLength(3)
    expect(chart.yAxis).toEqual(['Hardware', 'Mechanical'])
    expect(chart.series[0]!.values).toEqual([10, null])
    expect(preserveRows.badgeCount(chart, 'x')).toBe(3)
    expect(preserveRows.badgeCount(chart, 'y')).toBe(2)
    expect(preserveRows.badgeCount(chart, 'z')).toBe(1)
    // points path preferred for grandTotal
    expect(preserveRows.grandTotal(chart)).toBe(60)
    expect(preserveRows.is3D(chart)).toBe(false)
    expect(preserveRows.is3D(chart, { threeD: true })).toBe(true)
  })

  it('grandTotal filters visibleZ and series/mixed fallbacks', () => {
    const withZ = emptyChart({
      zAxis: ['Z1', 'Z2'],
      points: [
        { xAxis: 'A', yAxis: 'Y', zAxis: 'Z1', value: 10 },
        { xAxis: 'A', yAxis: 'Y', zAxis: 'Z2', value: 5 },
      ],
    })
    expect(preserveRows.grandTotal(withZ, { Z1: true, Z2: false })).toBe(10)

    const mixedOnly = emptyChart({
      mixedTuples: [
        [0, 3],
        [1, 4],
      ],
    })
    expect(preserveRows.grandTotal(mixedOnly)).toBe(7)

    const seriesOnly = emptyChart({
      series: [
        { xAxis: 'A', values: [1, null, 2], benchmarkId: '' },
        { xAxis: 'B', values: [3], benchmarkId: '' },
      ],
    })
    expect(preserveRows.grandTotal(seriesOnly)).toBe(6)
  })

  it('is3D requires x+y+z series shape', () => {
    const chart = emptyChart({
      yAxis: ['Y'],
      zAxis: ['Z'],
      series: [{ xAxis: 'X', values: [1], benchmarkId: '' }],
    })
    expect(preserveRows.is3D(chart)).toBe(true)
    expect(
      preserveRows.is3D(
        emptyChart({ yAxis: ['Y'], series: [{ xAxis: 'X', values: [1], benchmarkId: '' }] })
      )
    ).toBe(false)
  })
})

describe('GroupedBuilder query methods and canOfferValue3D', () => {
  it('badgeCount / grandTotal cover mixedTuples branches', () => {
    const chart = emptyChart({
      mixedTuples: [
        [0, 1],
        [1, 2],
      ],
      xCategories: ['A', 'B'],
    })
    expect(grouped.badgeCount(chart, 'x')).toBe(2)
    expect(grouped.badgeCount(chart, 'y')).toBe(2)
    expect(grouped.badgeCount(chart, 'z')).toBe(0)
    expect(grouped.grandTotal(chart)).toBe(3)

    const seriesOnly = emptyChart({
      series: [{ xAxis: 'A', values: [1, null], benchmarkId: '' }],
    })
    expect(grouped.grandTotal(seriesOnly)).toBe(1)
  })

  it('canOfferValue3D distinguishes scatter vs bar/line', () => {
    const data = [dp('a', 'b', 'z')]
    expect(grouped.canOfferValue3D('scatter', data, false, { threeD: true })).toBe(true)
    expect(grouped.canOfferValue3D('scatter', data, true)).toBe(false)
    expect(grouped.canOfferValue3D('bar', data, false)).toBe(true)
    expect(grouped.canOfferValue3D('line', data, false)).toBe(true)
    expect(grouped.canOfferValue3D('pie', data, false)).toBe(false)
    expect(grouped.canOfferValue3D('bar', [dp('a', '')], false)).toBe(false)
  })

  it('is3D matches x/y/z presence and threeD flag', () => {
    const full = emptyChart({
      yAxis: ['Y'],
      zAxis: ['Z'],
      series: [{ xAxis: 'X', values: [1], benchmarkId: '' }],
    })
    expect(grouped.is3D(full)).toBe(true)
    const xy = emptyChart({
      yAxis: ['Y'],
      series: [{ xAxis: 'X', values: [1], benchmarkId: '' }],
    })
    expect(grouped.is3D(xy, { threeD: true })).toBe(true)
    expect(grouped.is3D(xy)).toBe(false)
  })

  it('build averages duplicates and skips missing stats', () => {
    const data: DataPoint[] = [
      dp('A', 'Y1', '', 'val', 10),
      dp('A', 'Y1', '', 'val', 20),
      { xAxis: 'B', yAxis: 'Y1', stats: [{ type: 'other', value: 99 }] },
      { xAxis: 'C', yAxis: 'Y1', stats: [{ type: 'val' }] },
    ]
    const chart = grouped.build(data, baseCtx({ signature: 'val-' }))
    expect(chart.series.find((s) => s.xAxis === 'A')!.values).toEqual([15])
    expect(chart.series.find((s) => s.xAxis === 'B')).toBeUndefined()
  })
})

describe('finalizeChart edge branches', () => {
  const pieces = {
    statType: 'val',
    title: 'val',
    yAxisValues: ['Y'],
    zAxisValues: [''],
    series: [] as ChartData['series'],
    points: [] as ChartData['points'],
    xSet: new Set<string>(['A', 'B']),
  }

  it('mixedTuples without xCategories uses empty categories', () => {
    const chart = finalizeChart(
      {
        ...pieces,
        mixedTuples: [[0, 1]],
        xCategories: undefined,
      },
      baseCtx({ preserveRows: true })
    )
    // no mixedTuples attached when length ok but xCategories undefined → still spreads if length
    expect(chart.mixedTuples).toEqual([[0, 1]])
  })

  it('preserveRows series path maps x from series when not mixed', () => {
    const chart = finalizeChart(
      {
        ...pieces,
        series: [
          { xAxis: 'S1', values: [1], benchmarkId: '' },
          { xAxis: 'S2', values: [2], benchmarkId: '' },
        ],
        points: [
          { xAxis: 'S1', yAxis: 'Y', zAxis: '', value: 1 },
          { xAxis: 'S2', yAxis: 'Y', zAxis: '', value: 2 },
        ],
      },
      baseCtx({
        preserveRows: true,
        threeD: true,
        sort: noSort,
        canonical: { y: ['Y'] },
      })
    )
    expect(chart.series).toHaveLength(2)
    expect(chart.render3D?.mode).toBe('value')
  })

  it('sorts mixedTuples with missing category totals via ?? 0', () => {
    const chart = finalizeChart(
      {
        ...pieces,
        mixedTuples: [[0, 5]],
        xCategories: ['Has', 'Missing'],
      },
      baseCtx({ sort: ascSort, preserveRows: true })
    )
    // Missing total 0 sorts before Has(5) in asc
    expect(chart.xCategories).toEqual(['Missing', 'Has'])
    expect(chart.mixedTuples).toEqual([[1, 5]])
  })

  it('desc sort on mixedTuples and canonical remap with missing labels', () => {
    const chart = finalizeChart(
      {
        ...pieces,
        mixedTuples: [
          [0, 1],
          [1, 9],
        ],
        xCategories: ['A', 'B'],
      },
      baseCtx({ sort: descSort, preserveRows: true })
    )
    expect(chart.xCategories![0]).toBe('B')

    const canon = finalizeChart(
      {
        ...pieces,
        mixedTuples: [
          [0, 1],
          [1, 2],
        ],
        xCategories: ['A', 'B'],
      },
      baseCtx({
        sort: noSort,
        preserveRows: true,
        // canonical.x omits B → labelToNew misses → ?? xi
        canonical: { x: ['A', 'C'] },
      })
    )
    expect(canon.xCategories).toEqual(['A'])
    expect(canon.mixedTuples?.some(([xi]) => xi === 0 || xi === 1)).toBe(true)
  })

  it('canonical without x skips mixed remap when preserveRows', () => {
    const chart = finalizeChart(
      {
        ...pieces,
        mixedTuples: [[0, 1]],
        xCategories: ['A'],
      },
      baseCtx({
        sort: noSort,
        preserveRows: true,
        canonical: { y: ['Y'], z: [''] },
      })
    )
    expect(chart.xCategories).toEqual(['A'])
  })
})

describe('remaining builder branch edges', () => {
  it('preserveRows yOrder empty fallback is unreachable with points but empty yOrder forces Array.from', () => {
    // ySeen always fills yOrder when values exist; empty data still exercises yOrder.length ? branch false
    const chart = preserveRows.build([], baseCtx({ signature: 'val-', preserveRows: true }))
    expect(chart.series).toEqual([])
    expect(chart.mixedTuples ?? []).toEqual([])
  })

  it('mixed sort with missing category total uses ?? 0', () => {
    // two categories, tuples only for index 1 → index 0 total uses ?? 0
    const chart = finalizeChart(
      {
        statType: 'val',
        title: 'val',
        yAxisValues: [],
        zAxisValues: [],
        series: [],
        points: [],
        xSet: new Set(),
        mixedTuples: [[1, 10]],
        xCategories: ['Empty', 'Has'],
      },
      baseCtx({ sort: descSort, preserveRows: true })
    )
    expect(chart.xCategories![0]).toBe('Has')
  })

  it('value build skips rows with no numeric coords', () => {
    // identity xyz → target nnz maps x,y to name and only z lands on chart;
    // coords lack xAxis/yAxis → 3D undefined-coords guard skips the row.
    const chart = buildValueModeChart(
      [{ xAxis: '1', yAxis: '2', zAxis: '3', stats: [] }],
      [
        { key: 'x', type: 'value' },
        { key: 'y', type: 'value' },
        { key: 'z', type: 'value' },
      ],
      'xyz',
      'nnz',
      { threeD: true }
    )
    expect(chart.valueTuples ?? []).toEqual([])
    expect(chart.valuePoints3D).toBeUndefined()
  })

  it('value 2D skips row when coords lack both axes', () => {
    // identity xy → target nn maps both to name → coords {} → 2D undefined guard
    const chart = buildValueModeChart(
      [{ xAxis: '1', yAxis: '2', stats: [] }],
      [
        { key: 'x', type: 'value' },
        { key: 'y', type: 'value' },
      ],
      'xy',
      'nn',
      { threeD: false }
    )
    expect(chart.valueTuples).toEqual([])
  })

  it('mixed build skips rows with undefined axes', () => {
    const chart = buildMixedModeChart([{ yAxis: '1', stats: [] }], [])
    expect(chart.mixedTuples).toEqual([])
  })

  it('mixed 3D with all rows filtered keeps empty render', () => {
    // z axis present → 3D; non-finite y filters every row → no render3D, empty
    // xCategories; query methods fall back to 0.
    const chart = buildMixedModeChart(
      [{ xAxis: 'A', yAxis: 'bad', zAxis: '1', stats: [] }],
      [
        { key: 'x', type: 'category' },
        { key: 'y', type: 'value' },
        { key: 'z', type: 'value' },
      ]
    )
    expect(chart.mixedTuples).toBeUndefined()
    expect(chart.render3D).toBeUndefined()
    expect(mixed.badgeCount(chart, 'y')).toBe(0)
    expect(mixed.grandTotal(chart)).toBe(0)
  })

  it('mixed 3D renders line series and query methods sum the 3D data', () => {
    const chart = buildMixedModeChart(
      [
        { xAxis: 'A', yAxis: '12', zAxis: '100', stats: [] },
        { xAxis: 'B', yAxis: '11', zAxis: '1', stats: [] },
      ],
      [
        { key: 'x', type: 'category' },
        { key: 'y', type: 'value' },
        { key: 'z', type: 'value' },
      ]
    )
    expect(chart.render3D?.mode).toBe('mixed')
    expect(mixed.badgeCount(chart, 'y')).toBe(2)
    expect(mixed.grandTotal(chart)).toBe(23)
  })

  it('value 2D uses z as color when metric absent and z off-chart', () => {
    const chart = buildValueModeChart(
      [{ xAxis: '1', yAxis: '2', zAxis: '9', stats: [] }],
      [
        { key: 'x', type: 'value' },
        { key: 'y', type: 'value' },
        { key: 'z', type: 'value' },
      ],
      'xyz',
      'xyn', // z maps to name → off-chart
      { threeD: false }
    )
    expect(chart.valueTuples).toEqual([[1, 2, 9]])
  })

  it('value 2D skips z color when z is on chart and metric absent', () => {
    const chart = buildValueModeChart(
      [{ xAxis: '1', yAxis: '2', zAxis: '9', stats: [] }],
      [
        { key: 'x', type: 'value' },
        { key: 'y', type: 'value' },
        { key: 'z', type: 'value' },
      ],
      'xyz',
      'xyz', // z on chart → no color dim
      { threeD: false }
    )
    expect(chart.valueTuples).toEqual([[1, 2]])
  })
})
