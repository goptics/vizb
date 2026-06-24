// Framework-free chart transforms. These are the CPU-heavy parts of the chart
// pipeline (grouping rows into series/points, and building the 3D render grid),
// extracted so they can run inside a Web Worker off the main thread. No Vue,
// no echarts — pure data in, plain (clone-safe) data out.
import type {
  DataPoint,
  Axis,
  AxisLabels,
  ChartData,
  Point3D,
  SeriesData,
  Stat,
  Sort,
  SortOrder,
  Render3D,
  Series3DData,
  ScaleType,
} from '../types'
import {
  arrangementHasChartZ,
  sourceFieldForChartAxis,
  swapAxisLabels,
  translateAxisKey,
  type AxisKey,
} from './swap'

export type CanonicalAxisOrders = {
  x?: string[]
  y?: string[]
  z?: string[]
}

const fieldValue = (row: DataPoint, field: AxisKey): string => {
  if (field === 'name') return row.name ?? ''
  const rec = row as unknown as Record<string, string | undefined>
  return rec[field] ?? ''
}

// First-seen category order for a raw field — stable across group/arrangement changes.
export function canonicalValuesForField(raw: DataPoint[], field: AxisKey): string[] {
  const order: string[] = []
  const seen = new Set<string>()
  for (const row of raw) {
    const val = fieldValue(row, field)
    if (val && !seen.has(val)) {
      seen.add(val)
      order.push(val)
    }
  }
  return order
}

export function canonicalAxisOrders(
  raw: DataPoint[],
  identityKeys: AxisKey[],
  targetKeys: AxisKey[]
): CanonicalAxisOrders {
  const axes = ['xAxis', 'yAxis', 'zAxis'] as const
  const keyFor: Record<(typeof axes)[number], keyof CanonicalAxisOrders> = {
    xAxis: 'x',
    yAxis: 'y',
    zAxis: 'z',
  }
  const result: CanonicalAxisOrders = {}
  for (const axis of axes) {
    const field = sourceFieldForChartAxis(identityKeys, targetKeys, axis)
    if (field) result[keyFor[axis]] = canonicalValuesForField(raw, field)
  }
  return result
}

export function canonicalAxisOrdersFromStrings(
  raw: DataPoint[],
  identityString: string,
  targetString: string
): CanonicalAxisOrders {
  return canonicalAxisOrders(raw, translateAxisKey(identityString), translateAxisKey(targetString))
}

const applyCanonicalOrder = (values: string[], canonical: string[] | undefined): string[] => {
  if (!canonical?.length) return values
  const present = new Set(values)
  return canonical.filter((v) => present.has(v))
}

const toStatSignature = (stat: Stat): string => {
  if (!stat.per) {
    return `${stat.type}-${stat.unit}`
  }

  if (!stat.unit) {
    return stat.type
  }

  return `${stat.type}-${stat.unit}-${stat.per}`
}

// Sort series in place by their summed value across all y, mirroring the old
// main-thread `useSortedSeriesData`. Totals are computed once (not per compare).
function sortSeriesByTotal(series: SeriesData[], order: SortOrder): void {
  const totals = new Map<SeriesData, number>()
  for (const s of series)
    totals.set(
      s,
      s.values.reduce<number>((sum, v) => sum + (v ?? 0), 0)
    )
  series.sort((a, b) => {
    const diff = totals.get(a)! - totals.get(b)!
    return order === 'asc' ? diff : -diff
  })
}

export type ChartSignature = { signature: string; statTemplate: Stat }

// Enumerate the unique stat signatures in the dataset, in first-seen order.
// Each signature is one chart's stable identity; this order is the chart order.
// Cheap (one pass) — runs on both the worker and the main thread.
export function listChartSignatures(data: DataPoint[]): ChartSignature[] {
  if (!data?.length) return []

  const uniqueStats = data.reduce((acc, benchmark) => {
    for (const stat of benchmark.stats || []) {
      const signature = toStatSignature(stat)
      if (!acc.has(signature)) acc.set(signature, stat)
    }
    return acc
  }, new Map<string, Stat>())

  return Array.from(uniqueStats.entries()).map(([signature, statTemplate]) => ({
    signature,
    statTemplate,
  }))
}

// Build the ChartData for a single stat signature: group rows into a y→x value
// map plus a flat Point3D list and the per-dimension category sets, then attach
// the 3D render payload when the chart has x, y and z.
export function buildChartForSignature(
  data: DataPoint[],
  signature: string,
  statTemplate: Stat,
  labels: AxisLabels | undefined,
  sort: Sort,
  showLabels = false,
  scale: ScaleType = 'linear',
  canonical?: CanonicalAxisOrders,
  threeD = false
): ChartData {
  const dataMap = new Map<string, Map<string, number>>()
  const countMap = new Map<string, Map<string, number>>()
  const xAxisSet = new Set<string>()
  const yAxisSet = new Set<string>()
  const zAxisSet = new Set<string>()
  const points: Point3D[] = []

  for (const benchmarkData of data) {
    const { xAxis = '', yAxis = '', zAxis = '' } = benchmarkData
    const matchingStat = benchmarkData.stats?.find((s) => toStatSignature(s) === signature)
    const value = matchingStat?.value
    if (value === undefined) continue

    yAxisSet.add(yAxis)
    xAxisSet.add(xAxis)
    zAxisSet.add(zAxis)
    points.push({ xAxis, yAxis, zAxis, value })

    // Accumulate sum + count per (yAxis, xAxis) so duplicates are averaged below,
    // never silently overwritten. Grouped CSV/JSON arrives pre-summed from the Go
    // layer (one point per key → mean == value), so this only meaningfully fires
    // for benchmark count=N repeats, where the mean is the correct measurement.
    if (!dataMap.has(yAxis)) {
      dataMap.set(yAxis, new Map())
      countMap.set(yAxis, new Map())
    }
    const yMap = dataMap.get(yAxis)!
    const cMap = countMap.get(yAxis)!
    yMap.set(xAxis, (yMap.get(xAxis) ?? 0) + value)
    cMap.set(xAxis, (cMap.get(xAxis) ?? 0) + 1)
  }

  for (const [yAxis, xMap] of dataMap) {
    const cMap = countMap.get(yAxis)!
    for (const [xAxis, sum] of xMap) xMap.set(xAxis, sum / cMap.get(xAxis)!)
  }

  let xAxisValues = Array.from(xAxisSet)
  let yAxisValues = Array.from(yAxisSet)
  let zAxisValues = Array.from(zAxisSet)
  if (!sort.enabled && canonical) {
    xAxisValues = applyCanonicalOrder(xAxisValues, canonical.x)
    yAxisValues = applyCanonicalOrder(yAxisValues, canonical.y)
    zAxisValues = applyCanonicalOrder(zAxisValues, canonical.z)
  }

  const series: SeriesData[] = xAxisValues.map((xAxis) => ({
    xAxis,
    values: yAxisValues.map((yAxis) => dataMap.get(yAxis)?.get(xAxis) ?? null),
    benchmarkId: data[0]?.name || '',
  }))

  // Sort the 2D series here (in the worker) rather than in the chart-option
  // computed on the main thread — for a wide x-axis (up to 100k series) the
  // per-series total + sort is the expensive bit and belongs off-thread. 3D
  // charts render off `render3D` (sorted separately below), so this only shapes
  // the 2D bar/line x order; harmless for the 3D path.
  if (sort.enabled) sortSeriesByTotal(series, sort.order)

  // Descriptive stats + correlation are NOT computed here — they're off the chart
  // critical path now, computed lazily in the dedicated stats worker only when a
  // panel is opened (see composables/useStatsWorker.ts). This keeps every
  // sort/group/scale recompute from blocking the chart reply behind stat work.
  const chart: ChartData = {
    title: statTemplate.type,
    statType: statTemplate.type,
    statUnit: statTemplate.unit,
    yAxis: yAxisValues,
    zAxis: zAxisValues,
    series,
    points,
    axisLabels: labels,
  }

  if (chartIsGrouped3D(chart)) {
    chart.render3D = build3DRender(chart.points, chart.zAxis, sort, showLabels, scale, canonical)
    chart.render3D.mode = 'grouped'
  } else if (threeD && chartIsValue3DEligible(chart)) {
    chart.render3D = buildValue3DRender(chart, sort, showLabels, scale, canonical)
  }

  return chart
}

// Build one ChartData per unique stat signature. Kept as the bulk entry point
// (tests + any non-worker caller); the worker uses the per-signature builder.
export function buildChartData(
  data: DataPoint[],
  labels: AxisLabels | undefined,
  sort: Sort,
  showLabels = false,
  scale: ScaleType = 'linear'
): ChartData[] {
  return listChartSignatures(data).map(({ signature, statTemplate }) =>
    buildChartForSignature(data, signature, statTemplate, labels, sort, showLabels, scale)
  )
}

const SWAP_AXIS_KEYS = new Set(['name', 'x', 'y', 'z'])

const identityStringFromAxes = (axes: Axis[]): string =>
  axes
    .filter((a) => SWAP_AXIS_KEYS.has(a.key))
    .map((a) => (a.key === 'name' ? 'n' : a.key.charAt(0)))
    .join('')

const axisLabelsFromAxes = (axes: Axis[]): AxisLabels => {
  const out: AxisLabels = {}
  for (const a of axes) {
    const label = a.label ?? a.key
    ;(out as Record<string, string>)[a.key] = label
  }
  return out
}

const projectValueCoords = (
  row: DataPoint,
  identityString: string,
  targetString: string
): Partial<Record<'xAxis' | 'yAxis' | 'zAxis', number>> | null => {
  const identityKeys = translateAxisKey(identityString)
  const targetKeys = translateAxisKey(targetString)
  if (identityKeys.length !== targetKeys.length) return null

  const out: Partial<Record<'xAxis' | 'yAxis' | 'zAxis', number>> = {}
  for (let i = 0; i < identityKeys.length; i++) {
    const src = identityKeys[i]!
    const dst = targetKeys[i]!
    if (dst === 'name') continue
    const n = Number(fieldValue(row, src))
    if (!isFinite(n)) return null
    out[dst] = n
  }
  return out
}

export type ValuePoint3D = [number, number, number, number?]

export function valuePoints3DToSeries(points: ValuePoint3D[], title: string): Series3DData[] {
  const withMetric = (points[0]?.length ?? 0) >= 4
  return [
    {
      name: title,
      data: points.map(([x, y, z, m]) =>
        withMetric && m !== undefined ? { value: [x, y, z, m] } : { value: [x, y, z] }
      ),
    },
  ]
}

// Value-mode 3D: continuous [x,y,z] path through space (--axes x,y,z + swap with z on chart).
export function buildValueMode3DRender(
  points: ValuePoint3D[],
  title: string,
  showLabels = false,
  scale: ScaleType = 'linear'
): Render3D {
  const filtered = scale === 'log' ? points.filter(([x, y, z]) => x > 0 && y > 0 && z > 0) : points
  const withMetric = (filtered[0]?.length ?? 0) >= 4
  const seriesData = valuePoints3DToSeries(filtered, title)[0]!.data
  const cellTotals: Record<string, number> = {}
  if (showLabels) {
    filtered.forEach((p, i) => {
      const labelVal = withMetric && p[3] !== undefined ? p[3] : p[2]
      cellTotals[String(i)] = labelVal ?? 0
    })
  }

  return {
    mode: 'continuous',
    xValues: [],
    yValues: [],
    zValues: [],
    barSeries: [{ name: title, data: seriesData }],
    lineSeries: [{ name: title, data: seriesData }],
    cellTotals,
  }
}

// Build one ChartData for a value-mode dataset (--axes pipeline). Coordinates are
// projected onto chart axes per the active swap (identity → target). Swap with z
// on a chart axis yields continuous 3D; otherwise 2D valueTuples for chart x vs y.
export function buildValueModeChart(
  data: DataPoint[],
  axes: Axis[],
  identityString?: string,
  targetString?: string,
  opts?: { scale?: ScaleType; showLabels?: boolean; threeD?: boolean }
): ChartData {
  const identity = identityString ?? identityStringFromAxes(axes)
  const target = targetString ?? identity
  const scale = opts?.scale ?? 'linear'
  const baseLabels = axisLabelsFromAxes(axes)
  const labels = { ...(swapAxisLabels(identity, target, baseLabels) ?? baseLabels) }
  const use3D = (opts?.threeD ?? true) && arrangementHasChartZ(target)

  const valueTuples: [number, number][] = []
  const valuePoints3D: ValuePoint3D[] = []

  for (const row of data) {
    const coords = projectValueCoords(row, identity, target)
    if (!coords) continue

    if (use3D) {
      const cx = coords.xAxis
      const cy = coords.yAxis
      const cz = coords.zAxis
      if (cx === undefined || cy === undefined || cz === undefined) continue
      if (scale === 'log' && (cx <= 0 || cy <= 0 || cz <= 0)) continue
      const metricRaw = row.metric
      const metricNum = metricRaw !== undefined && metricRaw !== '' ? Number(metricRaw) : undefined
      if (metricNum !== undefined && isFinite(metricNum)) {
        valuePoints3D.push([cx, cy, cz, metricNum])
      } else {
        valuePoints3D.push([cx, cy, cz])
      }
    } else {
      const cx = coords.xAxis
      const cy = coords.yAxis
      if (cx === undefined || cy === undefined) continue
      if (scale === 'log' && cy <= 0) continue
      valueTuples.push([cx, cy])
    }
  }

  const xLabel = labels.x ?? 'x'
  const yLabel = labels.y ?? 'y'
  const zLabel = labels.z ?? 'z'
  const title = use3D ? `${xLabel} · ${yLabel} · ${zLabel}` : `${xLabel} vs ${yLabel}`

  const chart: ChartData = {
    title,
    statType: 'value',
    yAxis: [],
    zAxis: [],
    series: [],
    points: [],
    axisLabels: labels,
    ...(!use3D ? { valueTuples } : {}),
    ...(valuePoints3D.length ? { valuePoints3D } : {}),
  }

  if (use3D && valuePoints3D.length) {
    chart.render3D = buildValueMode3DRender(valuePoints3D, title, opts?.showLabels ?? false, scale)
  }

  return chart
}

// Sort category values by their summed value across all points on that axis.
function sortByAxisTotal(
  values: string[],
  key: 'xAxis' | 'yAxis' | 'zAxis',
  points: Point3D[],
  order: SortOrder
): string[] {
  const totals = new Map<string, number>()
  for (const p of points) totals.set(p[key], (totals.get(p[key]) ?? 0) + p.value)
  return values.toSorted((a, b) => {
    const diff = (totals.get(a) ?? 0) - (totals.get(b) ?? 0)
    return order === 'asc' ? diff : -diff
  })
}

// Aggregate points into per-(x,y) cells for a single z series. Many rows can share
// the same (x,y,z); we average them (matching the 2D path) so benchmark count=N
// repeats render as their mean rather than a meaningless sum. Grouped CSV/JSON is
// pre-summed in the Go layer (one point per cell → mean == value).
function cellsFor(
  points: Point3D[],
  z: string,
  xIndex: Map<string, number>,
  yIndex: Map<string, number>
): Map<string, number> {
  const cells = new Map<string, number>()
  const counts = new Map<string, number>()
  for (const p of points) {
    if (p.zAxis !== z) continue
    const xi = xIndex.get(p.xAxis)
    const yi = yIndex.get(p.yAxis)
    if (xi === undefined || yi === undefined) continue
    const key = `${xi},${yi}`
    cells.set(key, (cells.get(key) ?? 0) + p.value)
    counts.set(key, (counts.get(key) ?? 0) + 1)
  }
  for (const [key, sum] of cells) cells.set(key, sum / counts.get(key)!)
  return cells
}

const sparseFromCells = (cells: Map<string, number>): { value: number[] }[] =>
  Array.from(cells, ([key, value]) => {
    const [xi, yi] = key.split(',').map(Number) as [number, number]
    return { value: [xi, yi, value] }
  })

const gridFromCells = (
  cells: Map<string, number>,
  xIndex: Map<string, number>,
  yIndex: Map<string, number>
): { value: number[] }[] => {
  const grid: { value: number[] }[] = []
  for (const xi of xIndex.values()) {
    for (const yi of yIndex.values()) {
      grid.push({ value: [xi, yi, cells.get(`${xi},${yi}`) ?? 0] })
    }
  }
  return grid
}

// Build the 3D render payload: sorted axis categories plus per-z series data for
// bar3D (full grid — keeps stacked bars seated) and line3D (sparse — a 0-grid
// would drag every line to the floor).
export function build3DRender(
  points: Point3D[],
  zAxisAll: string[],
  sort: Sort,
  showLabels = false,
  scale: ScaleType = 'linear',
  canonical?: CanonicalAxisOrders
): Render3D {
  let xValues = Array.from(new Set(points.map((p) => p.xAxis)))
  let yValues = Array.from(new Set(points.map((p) => p.yAxis)))
  let zValues = zAxisAll.filter((z) => z !== '')

  if (!sort.enabled && canonical) {
    xValues = applyCanonicalOrder(xValues, canonical.x)
    yValues = applyCanonicalOrder(yValues, canonical.y)
    zValues = applyCanonicalOrder(zValues, canonical.z)
  }

  if (sort.enabled) {
    xValues = sortByAxisTotal(xValues, 'xAxis', points, sort.order)
    yValues = sortByAxisTotal(yValues, 'yAxis', points, sort.order)
    zValues = sortByAxisTotal(zValues, 'zAxis', points, sort.order)
  }

  const xIndex = new Map(xValues.map((v, i) => [v, i]))
  const yIndex = new Map(yValues.map((v, i) => [v, i]))

  // A log z-axis can't plot 0/negative values. bar3D's full 0-filled grid would
  // be invalid, so under log we drop non-positive cells and make bar sparse too
  // (same intent as the 2D log path nulling values <= 0).
  const isLog = scale === 'log'

  const barSeries: Series3DData[] = []
  const lineSeries: Series3DData[] = []
  for (const z of zValues) {
    const cells = cellsFor(points, z, xIndex, yIndex)
    if (isLog) for (const [k, v] of cells) if (v <= 0) cells.delete(k)
    barSeries.push({
      name: z,
      data: isLog ? sparseFromCells(cells) : gridFromCells(cells, xIndex, yIndex),
    })
    lineSeries.push({ name: z, data: sparseFromCells(cells) })
  }

  const cellTotals: Record<string, number> = {}
  if (showLabels) {
    // Use lineSeries (sparse) not barSeries (full grid) — sparse only contains
    // cells with real data, so key presence distinguishes real data from 0-fill.
    for (const s of lineSeries) {
      for (const item of s.data) {
        const [xi = 0, yi = 0, v = 0] = item.value
        const key = `${xi},${yi}`
        cellTotals[key] = (cellTotals[key] ?? 0) + v
      }
    }
  }

  return { xValues, yValues, zValues, barSeries, lineSeries, cellTotals, mode: 'grouped' }
}

export const chartIsGrouped3D = (c: ChartData): boolean => {
  const hasX = c.series.some((s) => s.xAxis && s.xAxis.trim() !== '')
  const hasY = c.yAxis.length > 0 && c.yAxis[0] !== ''
  const hasZ = c.zAxis.length > 0 && c.zAxis[0] !== ''
  return hasX && hasY && hasZ
}

export const chartIsValue3DEligible = (c: ChartData): boolean => {
  const hasX = c.series.some((s) => s.xAxis && s.xAxis.trim() !== '')
  const hasY = c.yAxis.length > 0 && c.yAxis[0] !== ''
  const hasZ = c.zAxis.length > 0 && c.zAxis[0] !== ''
  return hasX && hasY && !hasZ
}

// Value-mode 3D: project the 2D x×y matrix onto a single bar3D/line3D series
// with metric height on zAxis3D (no z grouping / legend).
export function buildValue3DRender(
  chart: ChartData,
  sort: Sort,
  showLabels = false,
  scale: ScaleType = 'linear',
  canonical?: CanonicalAxisOrders
): Render3D {
  let xValues = chart.series.map((s) => s.xAxis).filter((x) => x.trim() !== '')
  let yValues = chart.yAxis.filter((y) => y !== '')

  if (!sort.enabled && canonical) {
    xValues = applyCanonicalOrder(xValues, canonical.x)
    yValues = applyCanonicalOrder(yValues, canonical.y)
  }

  if (sort.enabled) {
    const points: Point3D[] = []
    for (let yi = 0; yi < yValues.length; yi++) {
      const yAxis = yValues[yi]!
      for (const s of chart.series) {
        points.push({ xAxis: s.xAxis, yAxis, zAxis: '', value: s.values[yi] ?? 0 })
      }
    }
    xValues = sortByAxisTotal(xValues, 'xAxis', points, sort.order)
    yValues = sortByAxisTotal(yValues, 'yAxis', points, sort.order)
  }

  const xIndex = new Map(xValues.map((v, i) => [v, i]))
  const yIndex = new Map(yValues.map((v, i) => [v, i]))
  const cells = new Map<string, number>()

  for (let yi = 0; yi < yValues.length; yi++) {
    for (const s of chart.series) {
      const xi = xIndex.get(s.xAxis)
      if (xi === undefined) continue
      cells.set(`${xi},${yi}`, s.values[yi] ?? 0)
    }
  }

  const isLog = scale === 'log'
  if (isLog) for (const [k, v] of cells) if (v <= 0) cells.delete(k)

  const barData = isLog ? sparseFromCells(cells) : gridFromCells(cells, xIndex, yIndex)
  const lineData = sparseFromCells(cells)
  const seriesName = chart.title

  const cellTotals: Record<string, number> = {}
  if (showLabels) {
    for (const [key, v] of cells) cellTotals[key] = v
  }

  return {
    mode: 'value',
    xValues,
    yValues,
    zValues: [],
    barSeries: [{ name: seriesName, data: barData }],
    lineSeries: [{ name: seriesName, data: lineData }],
    cellTotals,
  }
}

// Arrangement-aware, non-mutating grouping for the chart pipeline. Given a set
// of raw DataPoint rows and a permutation of axis keys (identityKeys → targetKeys),
// project each row's field values onto the target dimensions and bucket rows into
// groups keyed by whichever target dimension is 'name'.
//
// identityKeys[i] names the source field; targetKeys[i] names the destination.
// When targetKeys[i] is 'name', that value becomes the group key and is NOT
// written onto the output row. All other values are copied to their target field.
// stats are always carried through unchanged.
//
// Groups are ordered by first-seen insertion (Map order). The 'Default' group key
// is used when no target dimension maps to 'name', or when the mapped value is
// absent/empty — matching the useDataPoint `grouped` computed's fallback.
//
// Identity case (['name','xAxis'] → ['name','xAxis']) reproduces the existing
// useDataPoint behaviour: groups by name, output rows hold xAxis (no name field).
export function projectAndGroup(
  raw: DataPoint[],
  identityKeys: AxisKey[],
  targetKeys: AxisKey[]
): { grouped: Map<string, DataPoint[]>; groupNames: string[] } {
  const grouped = new Map<string, DataPoint[]>()

  // Length mismatch: treat as identity — one 'Default' group, rows projected as-is.
  if (identityKeys.length !== targetKeys.length) {
    const rows = raw.map((row) => ({ ...row }))
    grouped.set('Default', rows)
    return { grouped, groupNames: ['Default'] }
  }

  for (const row of raw) {
    let groupKey = 'Default'
    const out: DataPoint = row.stats ? { stats: row.stats } : {}

    const rowRec = row as Record<string, unknown>
    for (let i = 0; i < identityKeys.length; i++) {
      // Non-null: the loop bound guarantees i is always in range.
      const srcKey = identityKeys[i]!
      const dstKey = targetKeys[i]!
      const val = rowRec[srcKey]
      if (dstKey === 'name') {
        // This dimension becomes the group discriminator; exclude from output row.
        if (val !== undefined && val !== null && val !== '') groupKey = val as string
      } else {
        // dstKey is 'xAxis' | 'yAxis' | 'zAxis' here; cast to satisfy the indexer.
        ;(out as Record<string, unknown>)[dstKey] = val
      }
    }

    if (!grouped.has(groupKey)) grouped.set(groupKey, [])
    grouped.get(groupKey)!.push(out)
  }

  return { grouped, groupNames: Array.from(grouped.keys()) }
}
