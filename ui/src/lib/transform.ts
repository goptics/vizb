// Framework-free chart transforms. These are the CPU-heavy parts of the chart
// pipeline (grouping rows into series/points, and building the 3D render grid),
// extracted so they can run inside a Web Worker off the main thread. No Vue,
// no echarts — pure data in, plain (clone-safe) data out.
import type {
  DataPoint,
  AxisLabels,
  ChartData,
  Point3D,
  SeriesData,
  Stat,
  Sort,
  SortOrder,
  Render3D,
  Series3DData,
} from '../types'

const toStatSignature = (stat: Stat): string => `${stat.type}-${stat.unit}-${stat.per}`

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
  showLabels = false
): ChartData {
  const dataMap = new Map<string, Map<string, number>>()
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

    if (!dataMap.has(yAxis)) dataMap.set(yAxis, new Map())
    dataMap.get(yAxis)!.set(xAxis, value)
  }

  const xAxisValues = Array.from(xAxisSet)
  const yAxisValues = Array.from(yAxisSet)
  const zAxisValues = Array.from(zAxisSet)

  const series: SeriesData[] = xAxisValues.map((xAxis) => ({
    xAxis,
    values: yAxisValues.map((yAxis) => dataMap.get(yAxis)?.get(xAxis) || 0),
    benchmarkId: data[0]?.name || '',
  }))

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

  if (chartIs3D(chart)) chart.render3D = build3DRender(chart.points, chart.zAxis, sort, showLabels)

  return chart
}

// Build one ChartData per unique stat signature. Kept as the bulk entry point
// (tests + any non-worker caller); the worker uses the per-signature builder.
export function buildChartData(data: DataPoint[], labels: AxisLabels | undefined, sort: Sort, showLabels = false): ChartData[] {
  return listChartSignatures(data).map(({ signature, statTemplate }) =>
    buildChartForSignature(data, signature, statTemplate, labels, sort, showLabels)
  )
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
  return [...values].sort((a, b) => {
    const diff = (totals.get(a) ?? 0) - (totals.get(b) ?? 0)
    return order === 'asc' ? diff : -diff
  })
}

// Aggregate points into per-(x,y) cell sums for a single z series. Many rows can
// share the same (x,y,z); summing avoids coplanar WebGL z-fighting.
function cellsFor(
  points: Point3D[],
  z: string,
  xIndex: Map<string, number>,
  yIndex: Map<string, number>
): Map<string, number> {
  const cells = new Map<string, number>()
  for (const p of points) {
    if (p.zAxis !== z) continue
    const xi = xIndex.get(p.xAxis)
    const yi = yIndex.get(p.yAxis)
    if (xi === undefined || yi === undefined) continue
    cells.set(`${xi},${yi}`, (cells.get(`${xi},${yi}`) ?? 0) + p.value)
  }
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
export function build3DRender(points: Point3D[], zAxisAll: string[], sort: Sort, showLabels = false): Render3D {
  let xValues = Array.from(new Set(points.map((p) => p.xAxis)))
  let yValues = Array.from(new Set(points.map((p) => p.yAxis)))
  let zValues = zAxisAll.filter((z) => z !== '')

  if (sort.enabled) {
    xValues = sortByAxisTotal(xValues, 'xAxis', points, sort.order)
    yValues = sortByAxisTotal(yValues, 'yAxis', points, sort.order)
    zValues = sortByAxisTotal(zValues, 'zAxis', points, sort.order)
  }

  const xIndex = new Map(xValues.map((v, i) => [v, i]))
  const yIndex = new Map(yValues.map((v, i) => [v, i]))

  const barSeries: Series3DData[] = []
  const lineSeries: Series3DData[] = []
  for (const z of zValues) {
    const cells = cellsFor(points, z, xIndex, yIndex)
    barSeries.push({ name: z, data: gridFromCells(cells, xIndex, yIndex) })
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

  return { xValues, yValues, zValues, barSeries, lineSeries, cellTotals }
}

const chartIs3D = (c: ChartData): boolean => {
  const hasX = c.series.some((s) => s.xAxis && s.xAxis.trim() !== '')
  const hasY = c.yAxis.length > 0 && c.yAxis[0] !== ''
  const hasZ = c.zAxis.length > 0 && c.zAxis[0] !== ''
  return hasX && hasY && hasZ
}
