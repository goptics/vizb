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
  kendall,
  distanceCorr,
  correlationMatrix,
  computeProfiles,
  buildColumns,
  availableViews,
  computeDescriptive,
  computeCorrelation,
  selectCorrelationAxis,
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
    expect(d.zeros).toBe(0)
    expect(d.negatives).toBe(0)
    expect(d.outliers).toBe(0)
    expect(d.mean).toBeNaN()
    expect(d.geoMean).toBeNaN()
    expect(d.harmMean).toBeNaN()
    expect(d.ci95Lower).toBeNaN()
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
  it('zeros and negatives count correctly', () => {
    const d = describeStats([-2, -1, 0, 0, 1, 2])
    expect(d.zeros).toBe(2)
    expect(d.negatives).toBe(2)
  })
  it('geoMean and harmMean NaN when any value ≤ 0', () => {
    expect(describeStats([-1, 2, 3]).geoMean).toBeNaN()
    expect(describeStats([0, 2, 3]).harmMean).toBeNaN()
  })
  it('geoMean and harmMean correct for positive data', () => {
    // geoMean([1,4]) = sqrt(4) = 2; harmMean([1,4]) = 2/(1+1/4) = 8/5
    const d = describeStats([1, 4])
    expect(d.geoMean).toBeCloseTo(2, P)
    expect(d.harmMean).toBeCloseTo(8 / 5, P)
  })
  it('trimMean trims 10% from each end', () => {
    // [1..10]: trim 1 from each → [2..9] → mean = 5.5
    const d = describeStats([1, 2, 3, 4, 5, 6, 7, 8, 9, 10])
    expect(d.trimMean).toBeCloseTo(5.5, P)
  })
  it('sem = stdDev / sqrt(n)', () => {
    const d = describeStats([2, 4, 6])
    expect(d.sem).toBeCloseTo(d.stdDev / Math.sqrt(3), P)
  })
  it('ci95 bounds = mean ± 1.96·sem', () => {
    const d = describeStats([2, 4, 6])
    expect(d.ci95Lower).toBeCloseTo(d.mean - 1.96 * d.sem, P)
    expect(d.ci95Upper).toBeCloseTo(d.mean + 1.96 * d.sem, P)
  })
  it('fences and outlier count', () => {
    // [1,2,3,4,5]: q1=2, q3=4, iqr=2, lf=−1, uf=7 → no outliers
    const d = describeStats([1, 2, 3, 4, 5])
    expect(d.lowerFence).toBeCloseTo(-1, P)
    expect(d.upperFence).toBeCloseTo(7, P)
    expect(d.outliers).toBe(0)
    // adding an outlier at 100
    const d2 = describeStats([1, 2, 3, 4, 5, 100])
    expect(d2.outliers).toBe(1)
  })
  it('p1/p10/p90/p99 are computed', () => {
    const d = describeStats([1, 2, 3, 4, 5, 6, 7, 8, 9, 10])
    expect(d.p1).toBeCloseTo(1.09, 1)
    expect(d.p10).toBeCloseTo(1.9, 1)
    expect(d.p90).toBeCloseTo(9.1, 1)
    expect(d.p99).toBeCloseTo(9.91, 1)
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

describe('kendall', () => {
  it('perfect positive monotonic → 1', () =>
    expect(kendall([1, 2, 3], [2, 4, 6])).toBeCloseTo(1, P))
  it('perfect negative monotonic → -1', () =>
    expect(kendall([1, 2, 3], [6, 4, 2])).toBeCloseTo(-1, P))
  it('ties in x → τ-b (denominator reduced)', () => {
    // xs=[1,2,2,3] ys=[1,2,3,4]: C=5, D=0, Tx=1, Ty=0
    // τ_b = 5 / sqrt((5+0+1)(5+0+0)) = 5/sqrt(30)
    expect(kendall([1, 2, 2, 3], [1, 2, 3, 4])).toBeCloseTo(5 / Math.sqrt(30), P)
  })
  it('constant input (all ties) → NaN', () => expect(kendall([1, 1, 1], [1, 2, 3])).toBeNaN())
  it('< 2 complete pairs → NaN', () => expect(kendall([1], [2])).toBeNaN())
  it('pairwise-complete: NaN positions skipped', () =>
    expect(kendall([1, NaN, 3], [2, 5, 6])).toBeCloseTo(1, P))
})

describe('distanceCorr', () => {
  it('identical vectors → 1', () =>
    expect(distanceCorr([1, 2, 3, 4, 5], [1, 2, 3, 4, 5])).toBeCloseTo(1, P))
  it('detects V-shape (non-linear) dependence that pearson misses', () => {
    // 15-point centered parabola: pearson=0, dcor>0 (U-centering requires n≥15 for this shape)
    const xs = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14]
    const ys = [49, 36, 25, 16, 9, 4, 1, 0, 1, 4, 9, 16, 25, 36, 49]
    expect(distanceCorr(xs, ys)).toBeGreaterThan(0)
    expect(pearson(xs, ys)).toBeCloseTo(0, P)
  })
  it('constant input → NaN', () => expect(distanceCorr([1, 1, 1, 1], [1, 2, 3, 4])).toBeNaN())
  it('< 4 complete pairs → NaN', () => {
    expect(distanceCorr([1, 2], [3, 4])).toBeNaN() // n=2
    expect(distanceCorr([1, 2, 3], [3, 4, 5])).toBeNaN() // n=3 (1/(m*(m-3)) undefined)
  })
  it('output is in [0, 1]', () => {
    const v = distanceCorr([1, 3, 2, 5, 4], [4, 2, 5, 1, 3])
    expect(v).toBeGreaterThanOrEqual(0)
    expect(v).toBeLessThanOrEqual(1)
  })
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
  it('kendall: diagonal=1, symmetric, perfect relationships', () => {
    const m = correlationMatrix(
      [
        [1, 2, 3],
        [2, 4, 6],
        [3, 2, 1],
      ],
      'kendall'
    )
    expect(m[0]![0]).toBeCloseTo(1, P) // diagonal
    expect(m[0]![1]).toBeCloseTo(m[1]![0]!, P) // symmetric
    expect(m[0]![1]).toBeCloseTo(1, P) // [1,2,3] and [2,4,6] perfectly correlated
    expect(m[0]![2]).toBeCloseTo(-1, P) // [3,2,1] perfectly anti-correlated
  })
  it('dcor: diagonal=1, symmetric, non-negative off-diagonal', () => {
    // Need at least 4 points per column for dcor; use [1,2,3,4,5]
    const m = correlationMatrix(
      [
        [1, 2, 3, 4, 5],
        [2, 4, 6, 8, 10],
        [5, 4, 3, 2, 1],
      ],
      'dcor'
    )
    expect(m[0]![0]).toBeCloseTo(1, P)
    expect(m[0]![1]).toBeCloseTo(m[1]![0]!, P)
    expect(m[0]![1]).toBeGreaterThanOrEqual(0) // dcor ≥ 0
    expect(m[0]![2]).toBeGreaterThanOrEqual(0)
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

  it('computes correlation on the x axis when x and y both have ≥2 entities', () => {
    const points = pts(['A', 'p', 1], ['A', 'q', 2], ['B', 'p', 4], ['B', 'q', 5])
    const { correlation } = computeProfiles(points, ['A', 'B'], ['p', 'q'])
    // Both x=2 and y=2 are usable; auto-picks one (smallest, stable sort → x or y, both size 2)
    expect(correlation).toBeDefined()
    expect(correlation!.labels).toHaveLength(2)
  })

  it('falls back to y when only 1 series exists', () => {
    const points = pts(['A', 'p', 1], ['A', 'q', 2], ['A', 'r', 3])
    const { seriesProfiles, correlation } = computeProfiles(points, ['A'], ['p', 'q', 'r'])
    expect(seriesProfiles).toHaveLength(1)
    // x=1 (<2), y=3 (≥2) → y is usable
    expect(correlation).toBeDefined()
    expect(correlation!.axis).toBe('y')
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

describe('buildColumns', () => {
  const pts = (...t: [string, string, number][]): Point3D[] =>
    t.map(([xAxis, yAxis, value]) => ({ xAxis, yAxis, zAxis: '', value }))

  it('NaN-aligns columns by (series, yAxis); last point wins', () => {
    const cols = buildColumns(
      pts(['A', 'p', 1], ['A', 'q', 9], ['A', 'q', 2], ['B', 'p', 4]),
      ['A', 'B'],
      ['p', 'q', 'r']
    )
    expect(cols[0]).toEqual([1, 2, NaN]) // A: q overwritten 9→2, r absent
    expect(cols[1]![0]).toBe(4)
    expect(cols[1]![1]).toBeNaN() // B has no q
    expect(cols[1]![2]).toBeNaN()
  })
})

describe('selectCorrelationAxis', () => {
  const ids = (k: number) => Array.from({ length: k }, (_, i) => `S${i}`)
  const cat = (n: number) => Array.from({ length: n }, (_, i) => `c${i}`)

  it('picks the smallest axis (min entity count) when multiple qualify', () => {
    // 5 series, 4 categories → y (4) is smaller, auto-pick = y
    const sel = selectCorrelationAxis(ids(5), cat(4), [])
    expect(sel.axis).toBe('y')
    expect(sel.labels).toEqual(cat(4))
  })

  it('picks x when it is the only axis with ≥2 entities', () => {
    const sel = selectCorrelationAxis(ids(2), cat(1), [])
    expect(sel.axis).toBe('x')
    expect(sel.labels).toEqual(ids(2))
  })

  it('returns null when no axis has ≥2 entities', () => {
    expect(selectCorrelationAxis(ids(1), cat(1), []).axis).toBeNull()
    expect(selectCorrelationAxis([], [], []).axis).toBeNull()
  })

  it('no cap — large axis counts are usable (user can switch)', () => {
    const sel = selectCorrelationAxis(ids(5000), cat(3), [])
    expect(sel.axis).toBe('y') // 3 < 5000, y is smallest
    expect(sel.labels).toEqual(cat(3))
  })

  it('3D — picks smallest among x/y/z', () => {
    const sel = selectCorrelationAxis(ids(300), cat(250), cat(4))
    expect(sel.axis).toBe('z')
    expect(sel.labels).toEqual(cat(4))
  })
})

describe('availableViews', () => {
  const series = (k: number) => Array.from({ length: k }, (_, i) => `S${i}`)
  const num = (n: number) => Array.from({ length: n }, (_, i) => String(i + 1))

  it('correlation available when at least one axis has ≥2 entities', () => {
    expect(availableViews(series(1), num(1)).correlation).toBe(false) // neither axis ≥2
    expect(availableViews(series(2), num(1)).correlation).toBe(true) // x has 2
    expect(availableViews(series(1), num(2)).correlation).toBe(true) // y has 2
    expect(availableViews(series(2), num(3)).correlation).toBe(true)
    expect(availableViews(series(5000), num(5000)).correlation).toBe(true) // no cap
  })
})

describe('computeCorrelation axis pick', () => {
  const pts2d = (...t: [string, string, number][]): Point3D[] =>
    t.map(([xAxis, yAxis, value]) => ({ xAxis, yAxis, zAxis: '', value }))

  it('default 2D path correlates the series (axis x)', () => {
    const points = pts2d(
      ['A', 'p', 1],
      ['A', 'q', 2],
      ['A', 'r', 3],
      ['B', 'p', 2],
      ['B', 'q', 4],
      ['B', 'r', 6]
    )
    const corr = computeCorrelation(points, ['A', 'B'], ['p', 'q', 'r'])
    expect(corr!.axis).toBe('x')
    expect(corr!.labels).toEqual(['A', 'B'])
    expect(corr!.pearson[0]![1]).toBeCloseTo(1, P) // B = 2·A → perfectly correlated
  })

  it('result carries kendall and dcor matrices', () => {
    // Need at least 4 categories for dcor (≥4 complete pairs per series)
    const points = pts2d(
      ['A', 'p', 1],
      ['A', 'q', 2],
      ['A', 'r', 3],
      ['A', 's', 4],
      ['B', 'p', 2],
      ['B', 'q', 4],
      ['B', 'r', 6],
      ['B', 's', 8]
    )
    const corr = computeCorrelation(points, ['A', 'B'], ['p', 'q', 'r', 's'])
    expect(corr!.kendall).toBeDefined()
    expect(corr!.dcor).toBeDefined()
    expect(corr!.kendall[0]![0]).toBeCloseTo(1, P) // diagonal
    expect(corr!.dcor[0]![0]).toBeCloseTo(1, P)
    expect(corr!.kendall[0]![1]).toBeCloseTo(corr!.kendall[1]![0]!, P) // symmetric
    expect(corr!.dcor[0]![1]).toBeCloseTo(corr!.dcor[1]![0]!, P)
  })

  it('transposes to y (categories) when the series axis is too wide', () => {
    // 201 series (> cap) × 2 perfectly-(anti)correlated categories.
    const order = Array.from({ length: 201 }, (_, i) => `s${i}`)
    const points: Point3D[] = []
    for (let i = 0; i < order.length; i++) {
      points.push({ xAxis: order[i]!, yAxis: 'L1', zAxis: '', value: i })
      points.push({ xAxis: order[i]!, yAxis: 'L2', zAxis: '', value: -i })
    }
    const corr = computeCorrelation(points, order, ['L1', 'L2'])
    expect(corr!.axis).toBe('y')
    expect(corr!.labels).toEqual(['L1', 'L2'])
    expect(corr!.pearson[0]![1]).toBeCloseTo(-1, P) // L2 = −L1 across the 201 series
  })
})

describe('lazy compute pieces match computeProfiles', () => {
  const pts = (...t: [string, string, number][]): Point3D[] =>
    t.map(([xAxis, yAxis, value]) => ({ xAxis, yAxis, zAxis: '', value }))
  const labels = ['1', '2', '3', '4', '5', '6', '7', '8']
  const points = [
    ...labels.map((l): [string, string, number] => ['A', l, Number(l) * 2]),
    ...labels.map((l): [string, string, number] => ['B', l, Number(l) ** 2]),
  ]
  const order = ['A', 'B']

  it('each piece equals the corresponding field of the one-shot bundle', () => {
    const all = computeProfiles(pts(...points), order, labels)
    expect(computeDescriptive(pts(...points), order, labels)).toEqual(all.seriesProfiles)
    expect(computeCorrelation(pts(...points), order, labels)).toEqual(all.correlation)
  })

  it('lazy pieces return undefined when no axis has ≥2 entities', () => {
    const single = pts(['A', 'p', 1])
    expect(computeCorrelation(single, ['A'], ['p'])).toBeUndefined() // x=1, y=1 → no usable axis
  })
})
