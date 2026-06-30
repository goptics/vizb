// Singleton bridge to the dedicated stats worker. The worker is created lazily on
// first use (so charts that never open a panel never spin one up) and shared by
// every StatsPanel. Requests are correlated by an incrementing id; each piece
// (descriptive / correlation) is cached per ChartData object so reopening the
// *same* chart — or re-clicking a tab — is instant, while a recompute (a fresh
// ChartData) recomputes.
import StatsWorker from '../workers/stats.worker.ts?worker&inline'
import type { StatsResponse, StatsKind } from '../workers/stats.worker'
import type { CorrelationAxis } from '../lib/stats'
import { chartSeriesLabels } from '../lib/utils'
import type { ChartData, SeriesProfile, CorrelationMatrix } from '../types'

let worker: Worker | null = null
let nextId = 0
// id → resolve. `unknown` because each request resolves a different shape; the
// caller's typed wrapper narrows it.
const pending = new Map<number, (r: StatsResponse) => void>()

// Per-ChartData cache of the in-flight/settled promise for each piece. WeakMap so
// cached results don't pin ChartData objects in memory after the chart is replaced
// by a recompute — they're collected with their key.
type PieceCache = {
  descriptive?: Promise<SeriesProfile[]>
  // keyed by axis ('x'|'y'|'z') or 'auto' for the default pick
  correlation: Map<string, Promise<CorrelationMatrix | undefined>>
}
const cache = new WeakMap<ChartData, PieceCache>()

function getWorker(): Worker {
  if (!worker) {
    worker = new StatsWorker()
    worker.onmessage = (e: MessageEvent<StatsResponse>) => {
      const resolve = pending.get(e.data.id)
      if (resolve) {
        pending.delete(e.data.id)
        resolve(e.data)
      }
    }
  }
  return worker
}

// Post one kinded request and resolve when its matching reply lands.
function request(
  chartData: ChartData,
  kind: StatsKind,
  axis?: CorrelationAxis
): Promise<StatsResponse> {
  return new Promise<StatsResponse>((resolve) => {
    const id = nextId++
    pending.set(id, resolve)
    // points/yAxis/series are plain (markRaw) → structured-clone takes them directly.
    getWorker().postMessage({
      type: 'compute',
      id,
      kind,
      points: chartData.points,
      yAxis: chartData.yAxis,
      zAxis: chartData.zAxis,
      seriesOrder: chartSeriesLabels(chartData),
      axis,
    })
  })
}

function cacheFor(chartData: ChartData): PieceCache {
  let c = cache.get(chartData)
  if (!c) {
    c = { correlation: new Map() }
    cache.set(chartData, c)
  }
  return c
}

export function computeDescriptive(chartData: ChartData): Promise<SeriesProfile[]> {
  const c = cacheFor(chartData)
  return (c.descriptive ??= request(chartData, 'descriptive').then((r) => r.seriesProfiles ?? []))
}

export function computeCorrelation(
  chartData: ChartData,
  axis?: CorrelationAxis
): Promise<CorrelationMatrix | undefined> {
  const c = cacheFor(chartData)
  const key = axis ?? 'auto'
  let p = c.correlation.get(key)
  if (!p) {
    p = request(chartData, 'correlation', axis).then((r) => r.correlation)
    c.correlation.set(key, p)
  }
  return p
}
