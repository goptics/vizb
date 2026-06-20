import { describe, it, expect } from 'vitest'
import type { DataPoint } from '../types'
import type { ChartData, Axis } from '../types'
import {
  canOfferValue3D,
  chartAxisBadgeCount,
  chartHasPlottableData,
  computeChartGrandTotal,
  datasetHas3DEngine,
  datasetHasBothXY,
  datasetDimension,
  formatChartTotal,
  isValueMode,
  isHybridMode,
  isScatterTransformMode,
} from './utils'

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

describe('datasetHas3DEngine', () => {
  it('is true when raw data has z', () => {
    expect(datasetHas3DEngine([dp('a', 'b', 'z1')])).toBe(true)
  })

  it('is true when threeD was baked via --3d', () => {
    expect(datasetHas3DEngine([dp('a', 'b')], { threeD: true })).toBe(true)
    expect(datasetHas3DEngine([dp('a', 'b')], { threeD: false })).toBe(true)
  })

  it('is false for pure x+y without --3d', () => {
    expect(datasetHas3DEngine([dp('a', 'b')])).toBe(false)
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
        series: [{ xAxis: 'A', values: [10], benchmarkId: '' }],
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
        series: [{ xAxis: 'A', values: [10], benchmarkId: '' }],
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

  it('returns false with mixed category and value axes (hybrid, not value mode)', () => {
    const axes: Axis[] = [
      { key: 'x', label: 'Region' },
      { key: 'y', label: 'Category' },
      { key: 'z', label: 'Latency (ms)', type: 'value' },
    ]
    expect(isValueMode(axes)).toBe(false)
  })
})

describe('isHybridMode', () => {
  it('returns false for undefined or empty axes', () => {
    expect(isHybridMode(undefined)).toBe(false)
    expect(isHybridMode([])).toBe(false)
  })

  it('returns false for pure value mode', () => {
    const axes: Axis[] = [
      { key: 'x', label: 'x', type: 'value' },
      { key: 'y', label: 'y', type: 'value' },
    ]
    expect(isHybridMode(axes)).toBe(false)
  })

  it('returns true for 2 category axes + z value axis', () => {
    const axes: Axis[] = [
      { key: 'x', label: 'Region' },
      { key: 'y', label: 'Category' },
      { key: 'z', label: 'Latency (ms)', type: 'value' },
    ]
    expect(isHybridMode(axes)).toBe(true)
  })

  it('returns false when value axis is not z', () => {
    const axes: Axis[] = [
      { key: 'x', label: 'Name' },
      { key: 'y', label: 'Score', type: 'value' },
    ]
    expect(isHybridMode(axes)).toBe(false)
  })
})

describe('isScatterTransformMode', () => {
  const hybridAxes: Axis[] = [
    { key: 'x', label: 'Region' },
    { key: 'y', label: 'Category' },
    { key: 'z', label: 'Latency (ms)', type: 'value' },
  ]

  it('is true for scatter with value or hybrid axes', () => {
    expect(
      isScatterTransformMode('scatter', [
        { key: 'x', type: 'value' },
        { key: 'y', type: 'value' },
      ])
    ).toBe(true)
    expect(isScatterTransformMode('scatter', hybridAxes)).toBe(true)
  })

  it('is false for non-scatter chart types', () => {
    expect(
      isScatterTransformMode('bar', [
        { key: 'x', type: 'value' },
        { key: 'y', type: 'value' },
      ])
    ).toBe(false)
    expect(isScatterTransformMode('line', hybridAxes)).toBe(false)
  })
})
