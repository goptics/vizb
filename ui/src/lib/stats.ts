// Framework-free descriptive statistics. No Vue, no echarts — pure number[] in,
// plain numbers / clone-safe objects out, so this can run inside the transform
// Web Worker alongside lib/transform.ts.
//
// All inputs are treated as a finite population (not a sample): variance/SD use
// the /n divisor, skewness/kurtosis use population moments. Non-finite values
// (NaN / ±Inf) are dropped up front — callers pass NaN for "missing cell" so a
// zero-fill never biases a stat (see transform.ts).
import type {
  DescriptiveStats,
  SeriesProfile,
  CorrelationMatrix,
  Point3D,
} from '../types'

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

// Per-series NaN-aligned columns straight from the point cloud — the shared
// substrate every stat computes over. One column per series (in `seriesOrder`),
// each indexed by `yAxis` with NaN for an absent (x,y) cell so the chart's
// zero-fill never biases a stat. Last point wins per (x,y), mirroring the
// transform's dataMap collapse.
export function buildColumns(
  points: Point3D[],
  seriesOrder: string[],
  yAxis: string[]
): number[][] {
  const byX = new Map<string, Map<string, number>>()
  for (const p of points) {
    let row = byX.get(p.xAxis)
    if (!row) {
      row = new Map()
      byX.set(p.xAxis, row)
    }
    row.set(p.yAxis, p.value)
  }
  return seriesOrder.map((x) => {
    const row = byX.get(x)
    return yAxis.map((y) => {
      const v = row?.get(y)
      return v === undefined ? NaN : v
    })
  })
}

// Which axis supplies the correlation matrix's entities (its rows/cols). 'x' = the
// series-identity axis (`seriesOrder`); 'y' = the per-series category axis; 'z' =
// the third (3D-only) axis. The observation vector for each entity is its values
// across the tuple of the *other two* axes.
export type CorrelationAxis = 'x' | 'y' | 'z'

// All axes with ≥2 distinct entities, sorted by entity count ascending so the
// auto-pick ([0]) is always the smallest (cheapest) matrix. User can switch to
// any axis including large ones.
export function usableCorrelationAxes(
  seriesOrder: string[],
  yAxis: string[],
  zAxis: string[]
): CorrelationAxis[] {
  const counts: [CorrelationAxis, number][] = [
    ['x', seriesOrder.length],
    ['y', yAxis.length],
    ['z', zAxis.length],
  ]
  return counts
    .filter(([, n]) => n >= 2)
    .sort(([, a], [, b]) => a - b)
    .map(([ax]) => ax)
}

// Pick the axis to correlate along: prefer x (series), fall back to y, then z.
// Single source of truth shared by `availableViews` (gate) and `computeCorrelation`.
export function selectCorrelationAxis(
  seriesOrder: string[],
  yAxis: string[],
  zAxis: string[]
): { axis: CorrelationAxis | null; labels: string[] } {
  const axes = usableCorrelationAxes(seriesOrder, yAxis, zAxis)
  if (axes.length === 0) return { axis: null, labels: [] }
  const axis = axes[0]!
  const labels = axis === 'x' ? seriesOrder : axis === 'y' ? yAxis : zAxis
  return { axis, labels }
}

// One NaN-aligned column per entity along the chosen correlation axis, indexed by a
// stable union of the other two axes' value tuples. Generalises `buildColumns` (the
// x-axis, yAxis-indexed case) to any entity axis so the same `correlationMatrix` can
// fit over series, categories, or the z dimension. Last point wins per (entity, obs).
export function buildCorrelationColumns(
  points: Point3D[],
  axis: CorrelationAxis,
  seriesOrder: string[],
  yAxis: string[],
  zAxis: string[]
): { labels: string[]; columns: number[][] } {
  // The entity value and the (otherA, otherB) observation values for one point.
  const parts = (p: Point3D): [entity: string, obsA: string, obsB: string] =>
    axis === 'x'
      ? [p.xAxis, p.yAxis, p.zAxis]
      : axis === 'y'
        ? [p.yAxis, p.xAxis, p.zAxis]
        : [p.zAxis, p.xAxis, p.yAxis]
  const labels = axis === 'x' ? seriesOrder : axis === 'y' ? yAxis : zAxis

  // entity → (obsKey → value); plus a stable first-seen ordering of obs keys.
  const byEntity = new Map<string, Map<string, number>>()
  const obsKeys: string[] = []
  const seenObs = new Set<string>()
  for (const p of points) {
    const [entity, obsA, obsB] = parts(p)
    const obsKey = `${obsA} ${obsB}`
    if (!seenObs.has(obsKey)) {
      seenObs.add(obsKey)
      obsKeys.push(obsKey)
    }
    let row = byEntity.get(entity)
    if (!row) {
      row = new Map()
      byEntity.set(entity, row)
    }
    row.set(obsKey, p.value)
  }

  const columns = labels.map((entity) => {
    const row = byEntity.get(entity)
    return obsKeys.map((k) => {
      const v = row?.get(k)
      return v === undefined ? NaN : v
    })
  })
  return { labels, columns }
}

// Cheap, math-free precondition flags: which stat views are valid for this data
// shape. O(K) — counts only, no pairwise loops — so the panel can render its tabs
// (and the compute functions can guard) without paying for any heavy computation.
// Single source of truth for the gates.
export function availableViews(
  seriesOrder: string[],
  yAxis: string[],
  zAxis: string[] = []
): {
  correlation: boolean
} {
  return {
    // Available when some axis (x → y → z) yields a fittable matrix.
    correlation: selectCorrelationAxis(seriesOrder, yAxis, zAxis).axis !== null,
  }
}

// Per-series descriptive profiles. The eager part — always computed when a stats
// panel opens.
export function computeDescriptive(
  points: Point3D[],
  seriesOrder: string[],
  yAxis: string[]
): SeriesProfile[] {
  if (seriesOrder.length === 0) return []
  const columns = buildColumns(points, seriesOrder, yAxis)
  return seriesOrder.map((name, i) => ({ name, stats: describe(columns[i]!) }))
}

// Cross-entity correlation matrix (both methods). `axis` overrides the auto-pick
// when it's among the usable axes; otherwise falls back to the first usable.
// Returns undefined when no axis yields a fittable matrix.
export function computeCorrelation(
  points: Point3D[],
  seriesOrder: string[],
  yAxis: string[],
  zAxis: string[] = [],
  axis?: CorrelationAxis
): CorrelationMatrix | undefined {
  const usable = usableCorrelationAxes(seriesOrder, yAxis, zAxis)
  if (usable.length === 0) return undefined
  const resolved: CorrelationAxis = (axis && usable.includes(axis)) ? axis : usable[0]!
  const labels = resolved === 'x' ? seriesOrder : resolved === 'y' ? yAxis : zAxis
  const { labels: resolvedLabels, columns } = buildCorrelationColumns(points, resolved, seriesOrder, yAxis, zAxis)
  return {
    axis: resolved,
    labels: resolvedLabels.length ? resolvedLabels : labels.slice(),
    pearson: correlationMatrix(columns, 'pearson'),
    spearman: correlationMatrix(columns, 'spearman'),
  }
}

// Build the full per-chart stat bundle in one pass — descriptive profiles +
// correlation. Retained as a thin wrapper over the lazy pieces for callers/tests
// that want everything at once.
export function computeProfiles(
  points: Point3D[],
  seriesOrder: string[],
  yAxis: string[],
  zAxis: string[] = []
): {
  seriesProfiles: SeriesProfile[]
  correlation?: CorrelationMatrix
} {
  return {
    seriesProfiles: computeDescriptive(points, seriesOrder, yAxis),
    correlation: computeCorrelation(points, seriesOrder, yAxis, zAxis),
  }
}
