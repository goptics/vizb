// Dedicated Web Worker for descriptive statistics, kept OFF the chart's critical
// path. The transform worker builds and replies the chart immediately; stats are
// computed here only when the user opens a chart's stats panel (lazy). Bundled and
// base64-inlined via `?worker&inline` so the embedded HTML template stays
// self-contained (no external worker asset).
//
// Stateless: each request carries one chart's point cloud (already plain /
// markRaw, clone-safe — never the full raw dataset, so there's no second
// full-dataset clone). The `id` lets the caller match replies to requests.
import { computeDescriptive, computeCorrelation } from '../lib/stats'
import type { CorrelationAxis } from '../lib/stats'
import type { Point3D, SeriesProfile, CorrelationMatrix } from '../types'

// `kind` selects which (potentially expensive) piece to compute, so the panel can
// pull descriptive eagerly and defer correlation until its tab opens.
export type StatsKind = 'descriptive' | 'correlation'

export type StatsRequest = {
  type: 'compute'
  id: number
  kind: StatsKind
  points: Point3D[]
  yAxis: string[]
  zAxis: string[]
  seriesOrder: string[]
  axis?: CorrelationAxis
}
// Only the field matching the request `kind` is populated; the caller knows which
// one it asked for by `id`.
export type StatsResponse = {
  type: 'result'
  id: number
  seriesProfiles?: SeriesProfile[]
  correlation?: CorrelationMatrix
}

self.onmessage = (e: MessageEvent<StatsRequest>) => {
  const { id, kind, points, yAxis, zAxis, seriesOrder, axis } = e.data
  const res: StatsResponse = { type: 'result', id }
  switch (kind) {
    case 'descriptive':
      res.seriesProfiles = computeDescriptive(points, seriesOrder, yAxis)
      break
    case 'correlation':
      res.correlation = computeCorrelation(points, seriesOrder, yAxis, zAxis, axis)
      break
  }
  ;(self as unknown as Worker).postMessage(res satisfies StatsResponse)
}
