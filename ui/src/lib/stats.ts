// Framework-free descriptive statistics. No Vue, no echarts — pure number[] in,
// plain numbers / clone-safe objects out, so this can run inside the transform
// Web Worker alongside lib/transform.ts.
//
// All inputs are treated as a finite population (not a sample): variance/SD use
// the /n divisor, skewness/kurtosis use population moments. The one exception is
// the 95% confidence interval for the mean, which is a *sample* inference and so
// uses the sample SD (n-1) as the standard error and the Student-t critical value
// with df = n-1 (z=1.96 would be systematically too narrow for small n). Non-finite
// values (NaN / ±Inf) are dropped up front — callers pass NaN for "missing cell"
// so a zero-fill never biases a stat (see transform.ts).
import type { DescriptiveStats, SeriesProfile, CorrelationMatrix, Point3D } from '../types'

const finite = (xs: number[]): number[] => xs.filter((x) => Number.isFinite(x))

// ── Student-t critical value (for the 95% CI) ─────────────────────────────────
// The two-sided 95% critical value t*_{df} is monotone decreasing in df toward
// the normal z=1.95996. df here is always n-1 (an integer), and the value is only
// ever surfaced in a UI table, so a compact lookup with linear interpolation in
// 1/df space (where t* is near-linear) is more than enough — <0.15% max error
// across df=1…1000 vs the exact quantile, invisible at any render precision.
// Breakpoints are denser at low df where the curve is steepest.

const T95_TABLE: ReadonlyArray<readonly [df: number, t: number]> = [
  [1, 12.706205],
  [2, 4.3026527],
  [3, 3.1824463],
  [4, 2.7764451],
  [5, 2.5705818],
  [7, 2.3646243],
  [10, 2.2281389],
  [20, 2.0859634],
  [30, 2.0422725],
  [60, 2.0002978],
  [120, 1.9799304],
]
const T95_INF = 1.959964 // normal-limit critical value (df → ∞)

function tCritical95(df: number): number {
  if (df < 1) return NaN
  if (!Number.isFinite(df)) return T95_INF
  // Linear interpolation in 1/df between bracketing breakpoints; t* is
  // near-affine in 1/df, so this tracks the curve closely with few points.
  const interp = (d0: number, t0: number, d1: number, t1: number): number => {
    const u = 1 / d0
    const v = 1 / d1
    const w = 1 / df
    return t0 + ((t1 - t0) * (u - w)) / (u - v)
  }
  for (let i = 0; i < T95_TABLE.length - 1; i++) {
    const [d0, t0] = T95_TABLE[i]!
    const [d1, t1] = T95_TABLE[i + 1]!
    if (df >= d0 && df <= d1) return interp(d0, t0, d1, t1)
  }
  // Beyond the last finite breakpoint, interpolate toward the normal limit.
  return interp(
    T95_TABLE[T95_TABLE.length - 1]![0],
    T95_TABLE[T95_TABLE.length - 1]![1],
    Infinity,
    T95_INF
  )
}

// Sample variance (divisor n-1) — used only for the SEM/CI (sample inference).
const sampleVariance = (xs: number[]): number => {
  const v = finite(xs)
  if (v.length < 2) return NaN
  const m = mean(v)
  return v.reduce((s, x) => s + (x - m) ** 2, 0) / (v.length - 1)
}

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

// Median absolute deviation: median(|x - median(x)|). Returned unscaled (no
// 1.4826 consistency factor), so it is NOT σ-equivalent for normal data — the
// UI labels it as the raw MAD, not as a robust-SD surrogate.
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
  const nan = NaN
  if (!n) {
    return {
      count: 0,
      missing: xs.length,
      unique: 0,
      zeros: 0,
      negatives: 0,
      mean: nan,
      median: nan,
      mode: nan,
      geoMean: nan,
      harmMean: nan,
      trimMean: nan,
      variance: nan,
      stdDev: nan,
      cv: nan,
      sem: nan,
      cqv: nan,
      min: nan,
      max: nan,
      range: nan,
      iqr: nan,
      mad: nan,
      lowerFence: nan,
      upperFence: nan,
      outliers: 0,
      skewness: nan,
      kurtosis: nan,
      p1: nan,
      p5: nan,
      p10: nan,
      p25: nan,
      p75: nan,
      p90: nan,
      p95: nan,
      p99: nan,
      ci95Lower: nan,
      ci95Upper: nan,
    }
  }
  const sorted = v.toSorted((a, b) => a - b)
  const min = sorted[0]!
  const max = sorted[n - 1]!
  const m = mean(v)
  const sd = stdDev(v)
  const p25 = quantileSorted(sorted, 0.25)
  const p75 = quantileSorted(sorted, 0.75)
  const iqr = p75 - p25
  const lowerFence = p25 - 1.5 * iqr
  const upperFence = p75 + 1.5 * iqr
  // SEM uses the *sample* SD (n-1) since the CI is a sample inference about the
  // mean; the population `sd` above is for the spread descriptor, not inference.
  const sampleSd = Math.sqrt(sampleVariance(v))
  const sem = n >= 2 ? sampleSd / Math.sqrt(n) : nan
  const tcrit = tCritical95(n - 1)
  const allPositive = v.every((x) => x > 0)
  const trimK = Math.floor(n * 0.1)
  return {
    // counts
    count: n,
    missing: xs.length - n,
    unique: new Set(v).size,
    zeros: v.filter((x) => x === 0).length,
    negatives: v.filter((x) => x < 0).length,
    // center
    mean: m,
    median: quantileSorted(sorted, 0.5),
    mode: mode(v),
    geoMean: allPositive ? Math.exp(mean(v.map((x) => Math.log(x)))) : nan,
    harmMean: allPositive ? n / v.reduce((s, x) => s + 1 / x, 0) : nan,
    trimMean: mean(trimK > 0 ? sorted.slice(trimK, n - trimK) : sorted),
    // spread
    variance: variance(v),
    stdDev: sd,
    cv: m === 0 ? nan : sd / m,
    sem,
    // CQV (quartile dispersion coefficient) is defined for non-negative data;
    // guard against a non-positive denominator (p25+p75 ≤ 0), which otherwise
    // yields unstable/negative values for distributions spanning negative quantiles.
    cqv: p25 + p75 > 0 ? iqr / (p25 + p75) : nan,
    // extremes
    min,
    max,
    range: max - min,
    iqr,
    mad: mad(v),
    lowerFence,
    upperFence,
    outliers: v.filter((x) => x < lowerFence || x > upperFence).length,
    // shape
    skewness: skewness(v),
    kurtosis: kurtosis(v),
    // percentiles
    p1: quantileSorted(sorted, 0.01),
    p5: quantileSorted(sorted, 0.05),
    p10: quantileSorted(sorted, 0.1),
    p25,
    p75,
    p90: quantileSorted(sorted, 0.9),
    p95: quantileSorted(sorted, 0.95),
    p99: quantileSorted(sorted, 0.99),
    // 95% CI for the mean: mean ± t*_{n-1} · SEM (t critical, not z=1.96).
    ci95Lower: n >= 2 ? m - tcrit * sem : nan,
    ci95Upper: n >= 2 ? m + tcrit * sem : nan,
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
  const idx = xs.map<[number, number]>((v, i) => [v, i])
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

// Kendall's τ-b: concordant/discordant pair counting over pairwise-complete
// observations. τ-b formula handles ties; denominator = 0 → NaN (constant input).
export function kendall(a: number[], b: number[]): number {
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
  const m = xs.length
  let C = 0,
    D = 0,
    Tx = 0,
    Ty = 0
  for (let i = 0; i < m - 1; i++) {
    for (let j = i + 1; j < m; j++) {
      const dx = xs[i]! - xs[j]!
      const dy = ys[i]! - ys[j]!
      const sign = dx * dy
      if (sign > 0) C++
      else if (sign < 0) D++
      else if (dx === 0 && dy !== 0) Tx++
      else if (dy === 0 && dx !== 0) Ty++
      // joint tie (dx===0 && dy===0): contributes to neither count
    }
  }
  const den = Math.sqrt((C + D + Tx) * (C + D + Ty))
  return den === 0 ? NaN : (C - D) / den
}

// Bias-corrected distance correlation (Székely & Rizzo 2014).
// Uses double-centered pairwise Euclidean distance matrices. O(n²) time/space.
// Returns 0–1 (measures dependence, not direction). NaN for < 4 pairs or zero var.
export function distanceCorr(a: number[], b: number[]): number {
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
  const m = xs.length
  if (m < 4) return NaN

  // Bias-corrected U-centered distance matrix for a 1D vector (Székely & Rizzo 2014).
  // A[i,j] (i≠j) = |v_i - v_j| - rowSum_i/(k-2) - rowSum_j/(k-2) + grandSum/((k-1)(k-2)); diagonal = 0.
  function bcCenter(v: number[]): number[][] {
    const k = v.length
    const D: number[][] = Array.from({ length: k }, () => new Array<number>(k).fill(0))
    const rowSum = new Array<number>(k).fill(0)
    let grandSum = 0
    for (let i = 0; i < k; i++) {
      for (let j = 0; j < k; j++) {
        D[i]![j] = Math.abs(v[i]! - v[j]!)
      }
    }
    for (let i = 0; i < k; i++) {
      for (let j = 0; j < k; j++) {
        rowSum[i]! += D[i]![j]!
      }
      grandSum += rowSum[i]!
    }
    const A: number[][] = Array.from({ length: k }, () => new Array<number>(k).fill(0))
    for (let i = 0; i < k; i++) {
      for (let j = 0; j < k; j++) {
        if (i !== j) {
          A[i]![j] =
            D[i]![j]! - rowSum[i]! / (k - 2) - rowSum[j]! / (k - 2) + grandSum / ((k - 1) * (k - 2))
        }
      }
    }
    return A
  }

  const A = bcCenter(xs)
  const B = bcCenter(ys)

  // dCov²*(X,Y) = (1/(n(n-3))) * Σ_{i≠j} A[i,j]·B[i,j]
  const factor = 1 / (m * (m - 3))
  let dCovXY = 0,
    dVarX = 0,
    dVarY = 0
  for (let i = 0; i < m; i++) {
    for (let j = 0; j < m; j++) {
      if (i !== j) {
        dCovXY += A[i]![j]! * B[i]![j]!
        dVarX += A[i]![j]! * A[i]![j]!
        dVarY += B[i]![j]! * B[i]![j]!
      }
    }
  }
  dCovXY *= factor
  dVarX *= factor
  dVarY *= factor

  if (dVarX <= 0 || dVarY <= 0) return NaN
  // dCor²* = dCov²* / sqrt(dVar²*(X) · dVar²*(Y)); clamp to 0 before sqrt
  const dCor2 = dCovXY / Math.sqrt(dVarX * dVarY)
  return Math.sqrt(Math.max(0, dCor2))
}

export type CorrelationMethod = 'pearson' | 'spearman' | 'kendall' | 'dcor'

// Symmetric K×K correlation matrix over the given columns (each an aligned
// number[] of equal conceptual length). The diagonal is 1 for a non-degenerate
// column and NaN for a constant (zero-variance) column — consistent with the
// NaN that the same column produces off-diagonal, so a constant series never
// shows a misleading self-correlation of 1.
export function correlationMatrix(columns: number[][], method: CorrelationMethod): number[][] {
  const corr =
    method === 'spearman'
      ? spearman
      : method === 'kendall'
        ? kendall
        : method === 'dcor'
          ? distanceCorr
          : pearson
  const k = columns.length
  const m: number[][] = Array.from({ length: k }, () => new Array<number>(k).fill(NaN))
  // A column is degenerate when it has <2 finite values or all finite values
  // coincide (zero spread) — every method returns NaN off-diagonal for it.
  const isConstant = (col: number[]): boolean => {
    const f = finite(col)
    if (f.length < 2) return true
    const lo = Math.min(...f)
    const hi = Math.max(...f)
    return lo === hi
  }
  for (let i = 0; i < k; i++) {
    m[i]![i] = isConstant(columns[i]!) ? NaN : 1
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
const isOverlayYAxis = (yAxis: string[]): boolean =>
  yAxis.length === 0 || (yAxis.length === 1 && yAxis[0] === '')

/** Collect every value per x category when rows overlay at one category (empty y). */
export function buildOverlayColumns(points: Point3D[], seriesOrder: string[]): number[][] {
  const byX = new Map<string, number[]>()
  for (const p of points) {
    let col = byX.get(p.xAxis)
    if (!col) {
      col = []
      byX.set(p.xAxis, col)
    }
    col.push(p.value)
  }
  return seriesOrder.map((x) => byX.get(x) ?? [])
}

export function buildColumns(
  points: Point3D[],
  seriesOrder: string[],
  yAxis: string[]
): number[][] {
  if (isOverlayYAxis(yAxis)) return buildOverlayColumns(points, seriesOrder)

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

// Pick the axis to correlate along: the usable axis with the fewest entities
// (cheapest K×K matrix), breaking ties x → y → z. Single source of truth shared
// by `availableViews` (gate) and `computeCorrelation`.
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
  // The obs key joins the two other-axis values with a NUL byte so that distinct
  // tuples can never collide (e.g. ("a b","c") vs ("a","b c")); mirrors
  // buildColumnsGroupedByZ.
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

// Groups by (xAxis, zAxis) pair for 3D charts. Key uses null-byte separator so
// labels can be split cleanly without colliding on slash-containing values.
export function buildColumnsGroupedByZ(
  points: Point3D[],
  seriesOrder: string[],
  zAxis: string[],
  yAxis: string[]
): { labels: string[]; columns: number[][] } {
  const labels: string[] = []
  const labelSet = new Set<string>()
  for (const x of seriesOrder) {
    for (const z of zAxis) {
      const key = `${x}\0${z}`
      if (!labelSet.has(key)) {
        labelSet.add(key)
        labels.push(key)
      }
    }
  }
  const byKey = new Map<string, Map<string, number>>()
  for (const p of points) {
    const key = `${p.xAxis}\0${p.zAxis}`
    let row = byKey.get(key)
    if (!row) {
      row = new Map()
      byKey.set(key, row)
    }
    row.set(p.yAxis, p.value)
  }
  const columns = labels.map((key) => {
    const row = byKey.get(key)
    return yAxis.map((y) => {
      const v = row?.get(y)
      return v === undefined ? NaN : v
    })
  })
  return { labels, columns }
}

// Per-series descriptive profiles. The eager part — always computed when a stats
// panel opens. When zAxis has ≥2 distinct non-empty values (3D charts), profiles
// are split per (series × z) pair so each z-slice gets its own stats row.
export function computeDescriptive(
  points: Point3D[],
  seriesOrder: string[],
  yAxis: string[],
  zAxis: string[] = []
): SeriesProfile[] {
  if (seriesOrder.length === 0) return []
  const distinctZ = [...new Set(zAxis.filter(Boolean))]
  if (distinctZ.length >= 2) {
    const { labels, columns } = buildColumnsGroupedByZ(points, seriesOrder, distinctZ, yAxis)
    return labels.map((key, i) => {
      const [x, z] = key.split('\0') as [string, string]
      return { name: `${x} / ${z}`, stats: describe(columns[i]!) }
    })
  }
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
  const resolved: CorrelationAxis = axis && usable.includes(axis) ? axis : usable[0]!
  const labels = resolved === 'x' ? seriesOrder : resolved === 'y' ? yAxis : zAxis
  const { labels: resolvedLabels, columns } = buildCorrelationColumns(
    points,
    resolved,
    seriesOrder,
    yAxis,
    zAxis
  )
  return {
    axis: resolved,
    labels: resolvedLabels.length ? resolvedLabels : labels.slice(),
    pearson: correlationMatrix(columns, 'pearson'),
    spearman: correlationMatrix(columns, 'spearman'),
    kendall: correlationMatrix(columns, 'kendall'),
    dcor: correlationMatrix(columns, 'dcor'),
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
    seriesProfiles: computeDescriptive(points, seriesOrder, yAxis, zAxis),
    correlation: computeCorrelation(points, seriesOrder, yAxis, zAxis),
  }
}
