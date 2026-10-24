import type { Better, ChartData, SeriesData, Point3D, AxisLabels, ResolvedSortOrder } from '@/types'
import type { BuildContext } from './types'
import { resolveSort } from '@/lib/sort'
import {
  applyCanonicalOrder,
  sortSeriesByTotal,
  chartIsGrouped3D,
  chartIsValue3DEligible,
  build3DRender,
  buildValue3DRender,
} from '../transform'

function sortMixedTuplesByCategoryTotal(
  mixedTuples: [number, number][],
  xCategories: string[],
  order: ResolvedSortOrder
): { mixedTuples: [number, number][]; xCategories: string[] } {
  const totals = new Map<number, number>()
  for (const [xi, y] of mixedTuples) totals.set(xi, (totals.get(xi) ?? 0) + y)
  const sortedIndices = xCategories
    .map((_, i) => i)
    .toSorted((a, b) => {
      const diff = (totals.get(a) ?? 0) - (totals.get(b) ?? 0)
      return order === 'asc' ? diff : -diff
    })
  const indexToNew = new Map(sortedIndices.map((oldIdx, newIdx) => [oldIdx, newIdx]))
  return {
    xCategories: sortedIndices.map((i) => xCategories[i]!),
    mixedTuples: mixedTuples.map(([xi, y]) => [indexToNew.get(xi)!, y]),
  }
}

// Shared tail for the grouped/preserveRows builders: derives the x-axis value
// order, applies canonical ordering, sorts series, assembles the ChartData and
// dispatches the 3D render payload. Mutates nothing outside the returned chart.
export function finalizeChart(
  pieces: {
    statType: string
    statUnit?: string
    better?: Better
    title: string
    yAxisValues: string[]
    zAxisValues: string[]
    series: SeriesData[]
    points: Point3D[]
    axisLabels?: AxisLabels
    xSet: Set<string>
    mixedTuples?: [number, number][]
    xCategories?: string[]
  },
  ctx: BuildContext
): ChartData {
  const { canonical, threeD, preserveRows } = ctx
  // `best` becomes asc/desc from this chart's direction before any sorting.
  const sort = resolveSort(ctx.sort, pieces.better)
  let { yAxisValues, zAxisValues, mixedTuples, xCategories, series } = pieces

  let xAxisValues = mixedTuples
    ? (xCategories ?? [])
    : preserveRows
      ? series.map((s) => s.xAxis)
      : Array.from(pieces.xSet)

  if (!sort.enabled && canonical) {
    if (mixedTuples && xCategories && canonical.x?.length) {
      const ordered = applyCanonicalOrder(xCategories, canonical.x)
      const labelToNew = new Map(ordered.map((v, i) => [v, i]))
      const oldLabels = xCategories
      xCategories = ordered
      xAxisValues = ordered
      mixedTuples = mixedTuples.map(([xi, y]) => [labelToNew.get(oldLabels[xi]!) ?? xi, y])
    } else if (!preserveRows) {
      xAxisValues = applyCanonicalOrder(xAxisValues, canonical.x)
    }
    // Series values were built in the pre-reorder y order; permute them to
    // follow the canonical names or legends would label the wrong bars.
    const oldY = new Map(yAxisValues.map((y, i) => [y, i]))
    yAxisValues = applyCanonicalOrder(yAxisValues, canonical.y)
    series = series.map((s) => ({
      ...s,
      values: yAxisValues.map((y) => s.values[oldY.get(y)!] ?? null),
    }))
    zAxisValues = applyCanonicalOrder(zAxisValues, canonical.z)
  }

  if (sort.enabled && series.length) sortSeriesByTotal(series, sort.order)
  if (sort.enabled && mixedTuples?.length && xCategories?.length) {
    ;({ mixedTuples, xCategories } = sortMixedTuplesByCategoryTotal(
      mixedTuples,
      xCategories,
      sort.order
    ))
  }

  const chart: ChartData = {
    title: pieces.title,
    statType: pieces.statType,
    statUnit: pieces.statUnit,
    ...(pieces.better ? { better: pieces.better } : {}),
    yAxis: yAxisValues,
    zAxis: zAxisValues,
    series,
    points: pieces.points,
    axisLabels: pieces.axisLabels,
    ...(mixedTuples?.length ? { mixedTuples, xCategories } : {}),
  }

  if (chartIsGrouped3D(chart)) {
    chart.render3D = build3DRender(
      chart.points,
      chart.zAxis,
      sort,
      ctx.showLabels,
      ctx.scale,
      canonical,
      preserveRows
    )
    chart.render3D.mode = 'grouped'
  } else if (threeD && chartIsValue3DEligible(chart)) {
    chart.render3D = buildValue3DRender(chart, sort, ctx.showLabels, ctx.scale, canonical)
  }

  return chart
}

// Re-exported for builder query-method convenience.
export type { Point3D }
