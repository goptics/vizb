// Framework-free descriptive statistics. No Vue, no echarts — pure number[] in,
// plain numbers / clone-safe objects out, so this can run inside the transform
// Web Worker alongside lib/transform.ts.
//
// All inputs are treated as a finite population (not a sample): variance/SD use
// the /n divisor, skewness/kurtosis use population moments. Non-finite values
// (NaN / ±Inf) are dropped up front — callers pass NaN for "missing cell" so a
// zero-fill never biases a stat (see transform.ts).
import type { DescriptiveStats, SeriesProfile, CorrelationMatrix, Point3D } from '../types'

const finite = (xs: number[]): number[] => xs.filter((x) => Number.isFinite(x))

export const mean = (xs: number[]): number =>
  xs.length ? xs.reduce((s, x) => s + x, 0) / xs.length : NaN

// Quantile of an already-ascending-sorted array, linear interpolation between
// closest ranks (numpy/“type 7” default). p in [0,1].
export function quantileSorted(sorted: number[], p: number): number {
  const n = sorted.length
  if (!n) return NaN
  if (n === 1) return sorted[0]!
  const idx = p * (n - 1)
  const lo = Math.floor(idx)
  const hi = Math.ceil(idx)
  const frac = idx - lo
  return sorted[lo]! + (sorted[hi]! - sorted[lo]!) * frac
}

export const median = (xs: number[]): number => {
  const s = finite(xs).toSorted((a, b) => a - b)
  return quantileSorted(s, 0.5)
}

// Most frequent value; ties resolved to the smallest. All-unique data has no
// real mode — we return the smallest value (frequency 1) rather than NaN so the
// table always has a number, matching how pandas/D-Tale surface a modal value.
export function mode(xs: number[]): number {
  const v = finite(xs)
  if (!v.length) return NaN
  const counts = new Map<number, number>()
  for (const x of v) counts.set(x, (counts.get(x) ?? 0) + 1)
  let best = v[0]!
  let bestCount = 0
  for (const [val, c] of counts) {
    if (c > bestCount || (c === bestCount && val < best)) {
      best = val
      bestCount = c
    }
  }
  return best
}

// Population variance (divisor n).
export function variance(xs: number[]): number {
  const v = finite(xs)
  if (!v.length) return NaN
  const m = mean(v)
  return v.reduce((s, x) => s + (x - m) ** 2, 0) / v.length
}

export const stdDev = (xs: number[]): number => Math.sqrt(variance(xs))

// Population skewness (Fisher-Pearson moment coefficient g1).
export function skewness(xs: number[]): number {
  const v = finite(xs)
  const n = v.length
  if (n < 2) return NaN
  const m = mean(v)
  const sd = stdDev(v)
  if (sd === 0) return 0
  const m3 = v.reduce((s, x) => s + (x - m) ** 3, 0) / n
  return m3 / sd ** 3
}

// Population excess kurtosis (g2): 0 for a normal distribution.
export function kurtosis(xs: number[]): number {
  const v = finite(xs)
  const n = v.length
  if (n < 2) return NaN
  const m = mean(v)
  const sd = stdDev(v)
  if (sd === 0) return 0
  const m4 = v.reduce((s, x) => s + (x - m) ** 4, 0) / n
  return m4 / sd ** 4 - 3
}

// Median absolute deviation: median(|x - median(x)|).
export function mad(xs: number[]): number {
  const v = finite(xs)
  if (!v.length) return NaN
  const med = median(v)
  return median(v.map((x) => Math.abs(x - med)))
}

// One pass over the (filtered, sorted) data producing every descriptive stat,
// so callers don't re-sort/re-reduce per metric.
export function describe(xs: number[]): DescriptiveStats {
  const v = finite(xs)
  const n = v.length
  if (!n) {
    const nan = NaN
    return {
      count: 0,
      missing: xs.length,
      unique: 0,
      mean: nan, median: nan, mode: nan,
      variance: nan, stdDev: nan, min: nan, max: nan, range: nan,
      iqr: nan, mad: nan, cv: nan, skewness: nan, kurtosis: nan,
      p5: nan, p25: nan, p75: nan, p95: nan,
    }
  }
  const sorted = v.toSorted((a, b) => a - b)
  const min = sorted[0]!
  const max = sorted[n - 1]!
  const m = mean(v)
  const sd = stdDev(v)
  const p25 = quantileSorted(sorted, 0.25)
  const p75 = quantileSorted(sorted, 0.75)
  return {
    count: n,
    missing: xs.length - n,
    unique: new Set(v).size,
    mean: m,
    median: quantileSorted(sorted, 0.5),
    mode: mode(v),
    variance: variance(v),
    stdDev: sd,
    min,
    max,
    range: max - min,
    iqr: p75 - p25,
    mad: mad(v),
    cv: m === 0 ? NaN : sd / m,
    skewness: skewness(v),
    kurtosis: kurtosis(v),
    p5: quantileSorted(sorted, 0.05),
    p25,
    p75,
    p95: quantileSorted(sorted, 0.95),
  }
}

// Pearson correlation over pairwise-complete observations (indices where both a
// and b are finite). Returns NaN when fewer than 2 complete pairs or either side
// is constant.
export function pearson(a: number[], b: number[]): number {
  const xs: number[] = []
  const ys: number[] = []
  const n = Math.min(a.length, b.length)
  for (let i = 0; i < n; i++) {
    const x = a[i]!
    const y = b[i]!
    if (Number.isFinite(x) && Number.isFinite(y)) {
      xs.push(x)
      ys.push(y)
    }
  }
  if (xs.length < 2) return NaN
  const mx = mean(xs)
  const my = mean(ys)
  let num = 0
  let dx2 = 0
  let dy2 = 0
  for (let i = 0; i < xs.length; i++) {
    const dx = xs[i]! - mx
    const dy = ys[i]! - my
    num += dx * dy
    dx2 += dx * dx
    dy2 += dy * dy
  }
  const den = Math.sqrt(dx2 * dy2)
  return den === 0 ? NaN : num / den
}

// Fractional ranks (ties share the average rank). Used by Spearman.
function ranks(xs: number[]): number[] {
  const idx = xs.map((v, i) => [v, i] as [number, number])
  idx.sort((p, q) => p[0] - q[0])
  const r = new Array<number>(xs.length)
  let i = 0
  while (i < idx.length) {
    let j = i
    while (j + 1 < idx.length && idx[j + 1]![0] === idx[i]![0]) j++
    const avg = (i + j) / 2 + 1 // average 1-based rank across the tie block
    for (let k = i; k <= j; k++) r[idx[k]![1]] = avg
    i = j + 1
  }
  return r
}

// Spearman = Pearson on the rank-transformed, pairwise-complete data.
export function spearman(a: number[], b: number[]): number {
  const xs: number[] = []
  const ys: number[] = []
  const n = Math.min(a.length, b.length)
  for (let i = 0; i < n; i++) {
    const x = a[i]!
    const y = b[i]!
    if (Number.isFinite(x) && Number.isFinite(y)) {
      xs.push(x)
      ys.push(y)
    }
  }
  if (xs.length < 2) return NaN
  return pearson(ranks(xs), ranks(ys))
}

export type CorrelationMethod = 'pearson' | 'spearman'

// Symmetric K×K correlation matrix over the given columns (each an aligned
// number[] of equal conceptual length). Diagonal is 1.
export function correlationMatrix(
  columns: number[][],
  method: CorrelationMethod
): number[][] {
  const corr = method === 'spearman' ? spearman : pearson
  const k = columns.length
  const m: number[][] = Array.from({ length: k }, () => new Array<number>(k).fill(NaN))
  for (let i = 0; i < k; i++) {
    m[i]![i] = 1
    for (let j = i + 1; j < k; j++) {
      const c = corr(columns[i]!, columns[j]!)
      m[i]![j] = c
      m[j]![i] = c
    }
  }
  return m
}

// Correlation is O(K²·N) and a K×K heatmap stops being legible well past this,
// so it's capped tighter than the (now virtualized, uncapped) descriptive table.
const CORR_MAX_SERIES = 60

// Build the chart's per-series descriptive profiles + cross-series correlation
// straight from its point cloud — the pure core the stats worker runs. One column
// per series (in `seriesOrder`), each indexed by `yAxis` with NaN for an absent
// (x,y) cell so the chart's zero-fill never biases a stat. Last point wins per
// (x,y), mirroring the transform's dataMap collapse.
export function computeProfiles(
  points: Point3D[],
  seriesOrder: string[],
  yAxis: string[]
): { seriesProfiles: SeriesProfile[]; correlation?: CorrelationMatrix } {
  if (seriesOrder.length === 0) return { seriesProfiles: [] }

  const byX = new Map<string, Map<string, number>>()
  for (const p of points) {
    let row = byX.get(p.xAxis)
    if (!row) {
      row = new Map()
      byX.set(p.xAxis, row)
    }
    row.set(p.yAxis, p.value)
  }

  const columns = seriesOrder.map((x) => {
    const row = byX.get(x)
    return yAxis.map((y) => {
      const v = row?.get(y)
      return v === undefined ? NaN : v
    })
  })

  const seriesProfiles: SeriesProfile[] = seriesOrder.map((name, i) => ({
    name,
    stats: describe(columns[i]!),
  }))

  let correlation: CorrelationMatrix | undefined
  // Need ≥2 series and ≥3 shared categories for a correlation to mean anything.
  if (seriesOrder.length >= 2 && seriesOrder.length <= CORR_MAX_SERIES && yAxis.length >= 3) {
    correlation = {
      labels: seriesOrder.slice(),
      pearson: correlationMatrix(columns, 'pearson'),
      spearman: correlationMatrix(columns, 'spearman'),
    }
  }

  return { seriesProfiles, correlation }
}
