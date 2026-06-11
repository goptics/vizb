// Dedicated Web Worker for descriptive statistics, kept OFF the chart's critical
// path. The transform worker builds and replies the chart immediately; stats are
// computed here only when the user opens a chart's stats panel (lazy). Bundled and
// base64-inlined via `?worker&inline` so the single-file HTML output stays
// self-contained (no external worker asset).
//
// Stateless: each request carries one chart's point cloud (already plain /
// markRaw, clone-safe — never the full raw dataset, so there's no second
// full-dataset clone). The `id` lets the caller match replies to requests.
import { computeProfiles } from '../lib/stats'
import type { Point3D, SeriesProfile, CorrelationMatrix } from '../types'

export type StatsRequest = {
  type: 'compute'
  id: number
  points: Point3D[]
  yAxis: string[]
  seriesOrder: string[]
}
export type StatsResponse = {
  type: 'result'
  id: number
  seriesProfiles: SeriesProfile[]
  correlation?: CorrelationMatrix
}

self.onmessage = (e: MessageEvent<StatsRequest>) => {
  const { id, points, yAxis, seriesOrder } = e.data
  const { seriesProfiles, correlation } = computeProfiles(points, seriesOrder, yAxis)
  ;(self as unknown as Worker).postMessage({
    type: 'result',
    id,
    seriesProfiles,
    correlation,
  } satisfies StatsResponse)
}
