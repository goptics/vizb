import { describe, it, expect } from 'vitest'
import { ref } from 'vue'
import type { ChartData, Point3D } from '@/types'
import {
  fontSize,
  sortBy,
  sortByTotal,
  sortByValue,
  adjustForLogScaleLine,
  useSortedSeriesData,
  resolveLogScale,
  computeSeriesTotals,
  sortByAxisTotal,
} from './common'

describe('fontSize', () => {
  it('is 12', () => {
    expect(fontSize).toBe(12)
  })
})

describe('sortBy / sortByTotal / sortByValue', () => {
  it('sorts ascending and descending by key', () => {
    const items = [
      { total: 3, value: 30 },
      { total: 1, value: 10 },
      { total: 2, value: 20 },
    ]
    expect([...items].sort(sortByTotal('asc')).map((i) => i.total)).toEqual([1, 2, 3])
    expect([...items].sort(sortByTotal('desc')).map((i) => i.total)).toEqual([3, 2, 1])
    expect([...items].sort(sortByValue('asc')).map((i) => i.value)).toEqual([10, 20, 30])
    expect([...items].sort(sortByValue('desc')).map((i) => i.value)).toEqual([30, 20, 10])
  })

  it('works for custom keys via sortBy factory', () => {
    const byScore = sortBy('score')
    const rows = [{ score: 5 }, { score: 1 }, { score: 9 }]
    expect([...rows].sort(byScore('asc')).map((r) => r.score)).toEqual([1, 5, 9])
    expect([...rows].sort(byScore('desc')).map((r) => r.score)).toEqual([9, 5, 1])
  })
})

describe('adjustForLogScaleLine', () => {
  it('passes null through', () => {
    expect(adjustForLogScaleLine(null, 'log')).toBeNull()
    expect(adjustForLogScaleLine(null, 'linear')).toBeNull()
  })

  it('returns value unchanged on linear scale', () => {
    expect(adjustForLogScaleLine(0, 'linear')).toBe(0)
    expect(adjustForLogScaleLine(-3, 'linear')).toBe(-3)
    expect(adjustForLogScaleLine(4, 'linear')).toBe(4)
  })

  it('nulls non-positive values on log scale', () => {
    expect(adjustForLogScaleLine(0, 'log')).toBeNull()
    expect(adjustForLogScaleLine(-1, 'log')).toBeNull()
    expect(adjustForLogScaleLine(2, 'log')).toBe(2)
  })
})

describe('useSortedSeriesData', () => {
  it('maps series and xAxisData without re-sorting', () => {
    const chartData = ref<ChartData>({
      title: 't',
      statType: 'avg',
      yAxis: [],
      zAxis: [],
      series: [
        { xAxis: 'b', values: [1], benchmarkId: '' },
        { xAxis: 'a', values: [2], benchmarkId: '' },
      ],
      points: [],
    })
    const sort = ref({ enabled: true, order: 'asc' as const })
    const result = useSortedSeriesData(chartData, sort)
    expect(result.value.series).toBe(chartData.value.series)
    expect(result.value.xAxisData).toEqual(['b', 'a'])
    expect(typeof result.value.hasYAxis).toBe('boolean')
  })
})

describe('resolveLogScale', () => {
  it('falls back to linear when empty', () => {
    expect(resolveLogScale('log', [])).toBe('linear')
  })

  it('keeps log when any value is positive', () => {
    expect(resolveLogScale('log', [0.2, 0.5])).toBe('log')
    expect(resolveLogScale('log', [-5, 0, 3])).toBe('log')
  })

  it('falls back when only non-positive values exist', () => {
    expect(resolveLogScale('log', [-5, 0, null])).toBe('linear')
  })

  it('passes linear through unchanged', () => {
    expect(resolveLogScale('linear', [1, 2])).toBe('linear')
  })
})

describe('computeSeriesTotals', () => {
  it('sums plain numbers and nulls', () => {
    const totals = computeSeriesTotals([
      { name: 'a', data: [1, 2, null] },
      { name: 'b', data: [10] },
    ])
    expect(totals.get('a')).toBe(3)
    expect(totals.get('b')).toBe(10)
  })

  it('tolerates legacy { value } items and missing value', () => {
    const totals = computeSeriesTotals([
      { name: 'legacy', data: [{ value: 4 }, { value: 6 }, {}, null] },
    ])
    expect(totals.get('legacy')).toBe(10)
  })

  it('sums y from coerced [x, y] pairs', () => {
    const totals = computeSeriesTotals([
      {
        name: 'train',
        data: [
          [1, 10],
          [2, null],
          [4, 5],
        ],
      },
    ])
    expect(totals.get('train')).toBe(15)
  })
})

describe('sortByAxisTotal', () => {
  const points: Point3D[] = [
    { xAxis: 'x1', yAxis: 'y1', zAxis: 'z1', value: 10 },
    { xAxis: 'x1', yAxis: 'y2', zAxis: 'z1', value: 5 },
    { xAxis: 'x2', yAxis: 'y1', zAxis: 'z2', value: 3 },
    { xAxis: 'x3', yAxis: 'y1', zAxis: 'z1', value: 20 },
  ]

  it('sorts x categories by total asc/desc', () => {
    expect(sortByAxisTotal(['x1', 'x2', 'x3'], 'xAxis', points, 'asc')).toEqual(['x2', 'x1', 'x3'])
    expect(sortByAxisTotal(['x1', 'x2', 'x3'], 'xAxis', points, 'desc')).toEqual(['x3', 'x1', 'x2'])
  })

  it('treats missing totals as 0 on either side of the compare', () => {
    expect(sortByAxisTotal(['ghostA', 'ghostB', 'x2'], 'xAxis', points, 'asc')).toEqual([
      'ghostA',
      'ghostB',
      'x2',
    ])
    expect(sortByAxisTotal(['x2', 'ghostA'], 'xAxis', points, 'desc')).toEqual(['x2', 'ghostA'])
  })

  it('sorts by yAxis and zAxis keys', () => {
    expect(sortByAxisTotal(['y1', 'y2'], 'yAxis', points, 'desc')[0]).toBe('y1')
    expect(sortByAxisTotal(['z1', 'z2'], 'zAxis', points, 'asc')[0]).toBe('z2')
  })
})
