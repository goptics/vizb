import { describe, it, expect } from 'vitest'
import {
  mean,
  median,
  quantileSorted,
  mode,
  variance,
  stdDev,
  skewness,
  kurtosis,
  mad,
  describe as describeStats,
  pearson,
  spearman,
  correlationMatrix,
  computeProfiles,
} from './stats'
import type { Point3D } from '../types'

const P = 10 // toBeCloseTo precision

describe('mean / median', () => {
  it('mean averages', () => expect(mean([1, 2, 3, 4])).toBeCloseTo(2.5, P))
  it('mean of empty is NaN', () => expect(mean([])).toBeNaN())
  it('median odd', () => expect(median([3, 1, 2])).toBeCloseTo(2, P))
  it('median even interpolates', () => expect(median([1, 2, 3, 4])).toBeCloseTo(2.5, P))
})

describe('quantileSorted (type-7)', () => {
  const s = [1, 2, 3, 4, 5]
  it('p=0 → min', () => expect(quantileSorted(s, 0)).toBeCloseTo(1, P))
  it('p=0.25', () => expect(quantileSorted(s, 0.25)).toBeCloseTo(2, P))
  it('p=0.5', () => expect(quantileSorted(s, 0.5)).toBeCloseTo(3, P))
  it('p=0.75', () => expect(quantileSorted(s, 0.75)).toBeCloseTo(4, P))
  it('p=1 → max', () => expect(quantileSorted(s, 1)).toBeCloseTo(5, P))
  it('single element', () => expect(quantileSorted([42], 0.5)).toBeCloseTo(42, P))
  it('empty → NaN', () => expect(quantileSorted([], 0.5)).toBeNaN())
})

describe('mode', () => {
  it('most frequent', () => expect(mode([1, 2, 2, 3])).toBe(2))
  it('tie resolves to smallest', () => expect(mode([1, 2, 2, 3, 3])).toBe(2))
  it('all unique → smallest', () => expect(mode([5, 3, 9])).toBe(3))
  it('empty → NaN', () => expect(mode([])).toBeNaN())
})

describe('variance / stdDev (population, /n)', () => {
  it('variance', () => expect(variance([1, 2, 3, 4, 5])).toBeCloseTo(2, P))
  it('stdDev', () => expect(stdDev([1, 2, 3, 4, 5])).toBeCloseTo(Math.sqrt(2), P))
  it('constant → 0', () => expect(variance([2, 2, 2])).toBeCloseTo(0, P))
})

describe('skewness / kurtosis (population moments)', () => {
  it('symmetric → skew 0', () => expect(skewness([1, 2, 3, 4, 5])).toBeCloseTo(0, P))
  it('excess kurtosis', () => expect(kurtosis([1, 2, 3, 4, 5])).toBeCloseTo(-1.3, P))
  it('n<2 → NaN', () => {
    expect(skewness([1])).toBeNaN()
    expect(kurtosis([1])).toBeNaN()
  })
  it('constant → 0', () => {
    expect(skewness([5, 5, 5])).toBe(0)
    expect(kurtosis([5, 5, 5])).toBe(0)
  })
})

describe('mad', () => {
  it('median absolute deviation', () => expect(mad([1, 2, 3, 4, 5])).toBeCloseTo(1, P))
  it('empty → NaN', () => expect(mad([])).toBeNaN())
})

describe('describe', () => {
  it('empty (all-NaN) reports missing=length', () => {
    const d = describeStats([NaN, NaN])
    expect(d.count).toBe(0)
    expect(d.missing).toBe(2)
    expect(d.unique).toBe(0)
    expect(d.mean).toBeNaN()
  })
  it('drops non-finite, computes count/missing/unique/range/iqr', () => {
    const d = describeStats([1, 2, NaN, 4])
    expect(d.count).toBe(3)
    expect(d.missing).toBe(1)
    expect(d.unique).toBe(3)
    expect(d.min).toBeCloseTo(1, P)
    expect(d.max).toBeCloseTo(4, P)
    expect(d.range).toBeCloseTo(3, P)
    expect(d.iqr).toBeCloseTo(1.5, P)
    expect(d.mean).toBeCloseTo(7 / 3, P)
  })
})

describe('pearson', () => {
  it('perfect positive', () => expect(pearson([1, 2, 3], [2, 4, 6])).toBeCloseTo(1, P))
  it('perfect negative', () => expect(pearson([1, 2, 3], [6, 4, 2])).toBeCloseTo(-1, P))
  it('constant → NaN', () => expect(pearson([1, 2, 3], [5, 5, 5])).toBeNaN())
  it('<2 complete pairs → NaN', () => expect(pearson([1], [2])).toBeNaN())
  it('pairwise-complete only', () => expect(pearson([1, NaN, 3], [2, 5, 6])).toBeCloseTo(1, P))
})

describe('spearman', () => {
  it('monotonic non-linear → 1', () =>
    expect(spearman([1, 2, 3, 4], [1, 4, 9, 16])).toBeCloseTo(1, P))
  it('tie-averaged ranks', () =>
    // ranks_a=[1,2.5,2.5,4] vs ranks_b=[1,2,3,4] → 4.5/sqrt(22.5)
    expect(spearman([1, 2, 2, 3], [1, 2, 3, 4])).toBeCloseTo(0.9486832980505138, P))
})

describe('correlationMatrix', () => {
  it('diagonal 1, symmetric, signed off-diagonal', () => {
    const m = correlationMatrix(
      [
        [1, 2, 3],
        [2, 4, 6],
        [3, 2, 1],
      ],
      'pearson'
    )
    expect(m[0]![0]).toBeCloseTo(1, P)
    expect(m[1]![1]).toBeCloseTo(1, P)
    expect(m[0]![1]).toBeCloseTo(1, P)
    expect(m[0]![2]).toBeCloseTo(-1, P)
    expect(m[0]![1]).toBeCloseTo(m[1]![0]!, P)
  })
  it('NaN preserved for constant column', () => {
    const m = correlationMatrix(
      [
        [1, 2, 3],
        [5, 5, 5],
      ],
      'pearson'
    )
    expect(m[0]![1]).toBeNaN()
    expect(m[1]![0]).toBeNaN()
    expect(m[0]![0]).toBeCloseTo(1, P)
  })
})

describe('computeProfiles', () => {
  const pts = (...t: [string, string, number][]): Point3D[] =>
    t.map(([xAxis, yAxis, value]) => ({ xAxis, yAxis, zAxis: '', value }))

  it('NaN-fills absent (x,y) cells; correlation present when shape allows', () => {
    const points = pts(['A', 'p', 1], ['A', 'q', 2], ['B', 'p', 4], ['B', 'q', 5], ['B', 'r', 6])
    const { seriesProfiles, correlation } = computeProfiles(points, ['A', 'B'], ['p', 'q', 'r'])
    expect(seriesProfiles.map((s) => s.name)).toEqual(['A', 'B'])
    expect(seriesProfiles[0]!.stats.count).toBe(2) // A has no 'r'
    expect(seriesProfiles[0]!.stats.missing).toBe(1)
    expect(seriesProfiles[1]!.stats.count).toBe(3)
    expect(correlation).toBeDefined()
    expect(correlation!.labels).toEqual(['A', 'B'])
  })

  it('omits correlation with <3 categories', () => {
    const points = pts(['A', 'p', 1], ['A', 'q', 2], ['B', 'p', 4], ['B', 'q', 5])
    expect(computeProfiles(points, ['A', 'B'], ['p', 'q']).correlation).toBeUndefined()
  })

  it('omits correlation with a single series', () => {
    const points = pts(['A', 'p', 1], ['A', 'q', 2], ['A', 'r', 3])
    const { seriesProfiles, correlation } = computeProfiles(points, ['A'], ['p', 'q', 'r'])
    expect(seriesProfiles).toHaveLength(1)
    expect(correlation).toBeUndefined()
  })

  it('last point wins per (x,y)', () => {
    const points = pts(['A', 'p', 1], ['A', 'p', 9])
    const { seriesProfiles } = computeProfiles(points, ['A'], ['p'])
    expect(seriesProfiles[0]!.stats.count).toBe(1)
    expect(seriesProfiles[0]!.stats.mean).toBeCloseTo(9, P)
  })

  it('empty series order → no profiles', () => {
    expect(computeProfiles([], [], []).seriesProfiles).toEqual([])
  })
})
