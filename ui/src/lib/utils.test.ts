import { computed } from 'vue'
import { describe, it, expect } from 'vitest'
import type { DataPoint } from '../types'
import type { ChartData, Axis } from '../types'
import {
  canOfferValue3D,
  chartAxisBadgeCount,
  chartHasPlottableData,
  computeChartGrandTotal,
  bundleHas3DChunk,
  datasetHasBothXY,
  datasetDimension,
  formatChartTotal,
  isValueMode,
  isMixedMode,
  isScatterTransformMode,
  getNextColorFor,
  resetColor,
  resolveLabelMode,
} from './utils'
import { applyTheme } from './themes'

describe('theme color allocation', () => {
  it('preserves raw key indices across differently sized palettes', () => {
    applyTheme('#111,#222,#333')
    resetColor()
    expect([getNextColorFor('a'), getNextColorFor('b'), getNextColorFor('c')]).toEqual([
      '#111',
      '#222',
      '#333',
    ])

    applyTheme('#aaa,#bbb')
    expect([getNextColorFor('a'), getNextColorFor('b'), getNextColorFor('c')]).toEqual([
      '#aaa',
      '#bbb',
      '#aaa',
    ])
    expect(getNextColorFor('d')).toBe('#bbb')
  })

  it('invalidates Vue computed options when the active palette changes', () => {
    applyTheme('#123,#456')
    resetColor()
    const color = computed(() => getNextColorFor('series'))
    expect(color.value).toBe('#123')
    applyTheme('#abc,#def')
    expect(color.value).toBe('#abc')
  })
})

const dp = (x: string, y: string, z = ''): DataPoint => ({
  name: '',
  xAxis: x,
  yAxis: y,
  zAxis: z,
  stats: [],
})

describe('datasetHasBothXY', () => {
  it('is true when any row has x and y', () => {
    expect(datasetHasBothXY([dp('a', 'b')])).toBe(true)
  })

  it('is false for x-only or empty data', () => {
    expect(datasetHasBothXY([{ name: '', xAxis: 'a', yAxis: '', zAxis: '', stats: [] }])).toBe(
      false
    )
    expect(datasetHasBothXY([])).toBe(false)
  })
})

describe('bundleHas3DChunk', () => {
  it('is true when raw data has z', () => {
    expect(bundleHas3DChunk([dp('a', 'b', 'z1')])).toBe(true)
  })

  it('is true when threeD was baked via --3d', () => {
    expect(bundleHas3DChunk([dp('a', 'b')], { threeD: true })).toBe(true)
    expect(bundleHas3DChunk([dp('a', 'b')], { threeD: false })).toBe(true)
  })

  it('is false for pure x+y without --3d', () => {
    expect(bundleHas3DChunk([dp('a', 'b')])).toBe(false)
  })
})

describe('canOfferValue3D', () => {
  const zData = [dp('a', 'b', 'z1')]

  it('offers toggle for bar/line with z-data when z is off chart axes', () => {
    expect(canOfferValue3D('bar', zData, false)).toBe(true)
    expect(canOfferValue3D('line', zData, false)).toBe(true)
  })

  it('hides toggle when z is on chart axes (grouped 3D)', () => {
    expect(canOfferValue3D('bar', zData, true)).toBe(false)
  })

  it('hides toggle for non-3D-capable chart types', () => {
    expect(canOfferValue3D('pie', zData, false)).toBe(false)
  })

  it('hides toggle for pure x+y without bundled 3D engine', () => {
    expect(canOfferValue3D('bar', [dp('a', 'b')], false)).toBe(false)
  })

  it('offers toggle for x+y data when --3d was baked', () => {
    expect(canOfferValue3D('bar', [dp('a', 'b')], false, { threeD: false })).toBe(true)
  })

  it('hides toggle for --axes value mode (2-col or 3-col)', () => {
    const valueAxes2: Axis[] = [
      { key: 'x', type: 'value' },
      { key: 'y', type: 'value' },
    ]
    const valueAxes3: Axis[] = [...valueAxes2, { key: 'z', type: 'value' }]
    expect(canOfferValue3D('bar', [dp('1', '2', '3')], false, undefined, valueAxes2)).toBe(false)
    expect(canOfferValue3D('line', [dp('1', '2', '3')], false, undefined, valueAxes3)).toBe(false)
  })

  it('offers toggle for scatter with z-data when z is off chart axes', () => {
    expect(canOfferValue3D('scatter', zData, false)).toBe(true)
  })

  it('hides toggle for scatter in value axes mode', () => {
    const valueAxes: Axis[] = [
      { key: 'x', type: 'value' },
      { key: 'y', type: 'value' },
    ]
    expect(canOfferValue3D('scatter', zData, false, undefined, valueAxes)).toBe(false)
  })
})

describe('datasetDimension', () => {
  it('classifies z-bearing data as 3D', () => {
    expect(datasetDimension([dp('a', 'b', 'z')])).toBe('3D')
  })
})

describe('computeChartGrandTotal', () => {
  const chart = (partial: Partial<ChartData>): ChartData => ({
    title: 'v',
    statType: 'v',
    yAxis: [],
    zAxis: [],
    series: [],
    points: [],
    ...partial,
  })

  it('sums 1D x-only points', () => {
    const total = computeChartGrandTotal(
      chart({
        series: [
          { xAxis: 'A', values: [10], benchmarkId: '' },
          { xAxis: 'B', values: [5], benchmarkId: '' },
        ],
        points: [
          { xAxis: 'A', yAxis: '', zAxis: '', value: 10 },
          { xAxis: 'B', yAxis: '', zAxis: '', value: 5 },
        ],
      })
    )
    expect(total).toBe(15)
  })

  it('sums 2D points across x and y', () => {
    const total = computeChartGrandTotal(
      chart({
        yAxis: ['Y1'],
        series: [
          { xAxis: 'A', values: [10], benchmarkId: '' },
          { xAxis: 'B', values: [7], benchmarkId: '' },
        ],
        points: [
          { xAxis: 'A', yAxis: 'Y1', zAxis: '', value: 10 },
          { xAxis: 'B', yAxis: 'Y1', zAxis: '', value: 7 },
        ],
      })
    )
    expect(total).toBe(17)
  })

  it('sums grouped 3D points and respects legend visibility', () => {
    const data = chart({
      yAxis: ['A'],
      zAxis: ['Z1', 'Z2'],
      points: [
        { xAxis: 'E', yAxis: 'A', zAxis: 'Z1', value: 10 },
        { xAxis: 'E', yAxis: 'A', zAxis: 'Z2', value: 5 },
      ],
      render3D: {
        mode: 'grouped',
        xValues: ['E'],
        yValues: ['A'],
        zValues: ['Z1', 'Z2'],
        barSeries: [],
        lineSeries: [
          { name: 'Z1', data: [{ value: [0, 0, 10] }] },
          { name: 'Z2', data: [{ value: [0, 0, 5] }] },
        ],
        cellTotals: { '0,0': 15 },
      },
    })
    expect(computeChartGrandTotal(data)).toBe(15)
    expect(computeChartGrandTotal(data, { Z1: true, Z2: false })).toBe(10)
  })

  it('falls back to series matrix when points are empty', () => {
    const total = computeChartGrandTotal(
      chart({
        yAxis: ['Y1', 'Y2'],
        series: [
          { xAxis: 'A', values: [3, 4], benchmarkId: '' },
          { xAxis: 'B', values: [1, null], benchmarkId: '' },
        ],
      })
    )
    expect(total).toBe(8)
  })

  it('uses the displayed aggregate instead of duplicate source points', () => {
    expect(
      computeChartGrandTotal(
        chart({
          series: [{ xAxis: 'A', values: [15], benchmarkId: '' }],
          points: [
            { xAxis: 'A', yAxis: '', zAxis: '', value: 10 },
            { xAxis: 'A', yAxis: '', zAxis: '', value: 20 },
          ],
        })
      )
    ).toBe(15)
  })

  it('ignores non-finite displayed values', () => {
    expect(
      computeChartGrandTotal(
        chart({
          series: [{ xAxis: 'A', values: [2, NaN, Infinity], benchmarkId: '' }],
        })
      )
    ).toBe(2)
  })

  it('excludes non-positive marks hidden by a log scale', () => {
    expect(
      computeChartGrandTotal(
        chart({
          series: [{ xAxis: 'A', values: [10, 0, -3], benchmarkId: '' }],
        }),
        undefined,
        'log'
      )
    ).toBe(10)
  })

  it('uses the displayed label dimension for value and mixed 3D totals', () => {
    expect(
      computeChartGrandTotal(
        chart({
          statType: 'value',
          valuePoints3D: [
            [1, 2, 3, 10],
            [4, 5, 6, 20],
          ],
        })
      )
    ).toBe(30)
    expect(
      computeChartGrandTotal(
        chart({
          statType: 'mixed',
          render3D: {
            mode: 'mixed',
            xValues: ['A'],
            yValues: [],
            zValues: [],
            barSeries: [],
            lineSeries: [{ name: 'value', data: [{ value: [0, 12, 5] }] }],
            cellTotals: {},
          },
        })
      )
    ).toBe(12)
  })
})

describe('resolveLabelMode', () => {
  it('prefers labelMode and preserves legacy showLabels as value mode', () => {
    expect(resolveLabelMode({ type: 'bar', showLabels: true })).toBe('value')
    expect(resolveLabelMode({ type: 'bar', showLabels: true, labelMode: 'percentage' })).toBe(
      'percentage'
    )
    expect(resolveLabelMode({ type: 'bar', showLabels: true, labelMode: 'none' })).toBe('none')
  })
})

describe('formatChartTotal', () => {
  it('rounds to two decimals', () => {
    expect(formatChartTotal(10.126)).toBe('10.13')
  })
})

describe('chartAxisBadgeCount', () => {
  const chart = (partial: Partial<ChartData>): ChartData => ({
    title: 'v',
    statType: 'v',
    yAxis: [],
    zAxis: [],
    series: [],
    points: [],
    ...partial,
  })

  it('uses series/y/z lengths for grouped charts', () => {
    const grouped = chart({
      yAxis: ['Y1', 'Y2'],
      zAxis: ['Z1'],
      series: [
        { xAxis: 'A', values: [1, 2], benchmarkId: '' },
        { xAxis: 'B', values: [3, 4], benchmarkId: '' },
      ],
    })
    expect(chartAxisBadgeCount(grouped, 'x')).toBe(2)
    expect(chartAxisBadgeCount(grouped, 'y')).toBe(2)
    expect(chartAxisBadgeCount(grouped, 'z')).toBe(1)
  })

  it('counts unique value-mode 2D coordinates per axis', () => {
    const value2d = chart({
      statType: 'value',
      valueTuples: [
        [1, 10],
        [1, 20],
        [2, 30],
      ],
    })
    expect(chartAxisBadgeCount(value2d, 'x')).toBe(2)
    expect(chartAxisBadgeCount(value2d, 'y')).toBe(3)
    expect(chartAxisBadgeCount(value2d, 'z')).toBe(0)
  })

  it('counts unique value-mode 3D coordinates per axis', () => {
    const value3d = chart({
      statType: 'value',
      valuePoints3D: [
        [1, 2, 3],
        [1, 2, 4],
        [5, 6, 7],
      ],
    })
    expect(chartAxisBadgeCount(value3d, 'x')).toBe(2)
    expect(chartAxisBadgeCount(value3d, 'y')).toBe(2)
    expect(chartAxisBadgeCount(value3d, 'z')).toBe(3)
  })

  it('counts mixedTuples points for overlay y badge', () => {
    const overlay = chart({
      xCategories: ['West', 'East'],
      mixedTuples: [
        [0, 10],
        [0, 20],
        [1, 30],
      ],
    })
    expect(chartAxisBadgeCount(overlay, 'x')).toBe(2)
    expect(chartAxisBadgeCount(overlay, 'y')).toBe(3)
  })

  it('counts category series when the stat is named value', () => {
    const grouped = chart({
      statType: 'value',
      xCategories: ['West', 'East'],
      mixedTuples: [
        [0, 12],
        [1, 18],
      ],
    })

    expect(chartAxisBadgeCount(grouped, 'x')).toBe(2)
  })

  it.each(['value', 'mixed', 'preserveRows'])(
    'counts grouped axes when the stat is named %s',
    (statType) => {
      const grouped = chart({
        statType,
        yAxis: ['Jan', 'Feb'],
        series: [
          { xAxis: 'West', values: [12, 18], benchmarkId: '' },
          { xAxis: 'East', values: [14, 20], benchmarkId: '' },
        ],
      })

      expect(chartAxisBadgeCount(grouped, 'x')).toBe(2)
      expect(chartAxisBadgeCount(grouped, 'y')).toBe(2)
    }
  )
})

describe('chartHasPlottableData', () => {
  const chart = (partial: Partial<ChartData>): ChartData => ({
    title: 'v',
    statType: 'v',
    yAxis: [],
    zAxis: [],
    series: [],
    points: [],
    ...partial,
  })

  it('is true for value-mode tuples and 3D points', () => {
    expect(chartHasPlottableData(chart({ valueTuples: [[1, 2]] }))).toBe(true)
    expect(chartHasPlottableData(chart({ valuePoints3D: [[1, 2, 3]] }))).toBe(true)
  })

  it('is false for empty value-mode chart', () => {
    expect(chartHasPlottableData(chart({ statType: 'value' }))).toBe(false)
  })
})

describe('isValueMode', () => {
  it('returns false for undefined axes', () => {
    expect(isValueMode(undefined)).toBe(false)
  })

  it('returns false for empty axes', () => {
    expect(isValueMode([])).toBe(false)
  })

  it('returns false when no axis has type value', () => {
    const axes: Axis[] = [{ key: 'x', label: 'Price' }]
    expect(isValueMode(axes)).toBe(false)
  })

  it('returns true when any axis has type value', () => {
    const axes: Axis[] = [
      { key: 'x', label: 'Price', type: 'value' },
      { key: 'y', label: 'Latency', type: 'value' },
    ]
    expect(isValueMode(axes)).toBe(true)
  })

  it('returns false with mixed category and value axes', () => {
    const axes: Axis[] = [
      { key: 'x', label: 'Region' },
      { key: 'y', label: 'Category' },
      { key: 'z', label: 'Latency (ms)', type: 'value' },
    ]
    expect(isValueMode(axes)).toBe(false)
  })
})

describe('isMixedMode', () => {
  it('returns false for undefined or empty axes', () => {
    expect(isMixedMode(undefined)).toBe(false)
    expect(isMixedMode([])).toBe(false)
  })

  it('returns true for category x + value y', () => {
    const axes: Axis[] = [
      { key: 'x', label: 'region' },
      { key: 'y', label: 'latency', type: 'value' },
    ]
    expect(isMixedMode(axes)).toBe(true)
  })

  it('returns false for all-value axes', () => {
    const axes: Axis[] = [
      { key: 'x', label: 'x', type: 'value' },
      { key: 'y', label: 'y', type: 'value' },
    ]
    expect(isMixedMode(axes)).toBe(false)
  })

  it('returns false for all-category axes', () => {
    const axes: Axis[] = [
      { key: 'x', label: 'region' },
      { key: 'y', label: 'group' },
    ]
    expect(isMixedMode(axes)).toBe(false)
  })
})

describe('isScatterTransformMode', () => {
  it('is true for value or mixed scatter axes', () => {
    expect(
      isScatterTransformMode([
        { key: 'x', type: 'value' },
        { key: 'y', type: 'value' },
      ])
    ).toBe(true)
    expect(
      isScatterTransformMode([
        { key: 'x', label: 'region' },
        { key: 'y', type: 'value' },
      ])
    ).toBe(true)
    expect(
      isScatterTransformMode([
        { key: 'x', label: 'region' },
        { key: 'y', label: 'group' },
      ])
    ).toBe(false)
  })
})

describe('canOfferValue3D with mixed axes', () => {
  it('returns false for mixed-axis datasets', () => {
    const axes: Axis[] = [
      { key: 'x', label: 'region' },
      { key: 'y', label: 'latency', type: 'value' },
    ]
    expect(canOfferValue3D('scatter', [dp('a', 'b')], false, undefined, axes)).toBe(false)
  })
})
