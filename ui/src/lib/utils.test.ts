import { describe, it, expect } from 'vitest'
import type { DataPoint } from '../types'
import type { ChartData } from '../types'
import {
  canOfferValue3D,
  computeChartGrandTotal,
  datasetHas3DEngine,
  datasetHasBothXY,
  datasetDimension,
  formatChartTotal,
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
