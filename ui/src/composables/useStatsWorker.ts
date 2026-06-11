// Singleton bridge to the dedicated stats worker. The worker is created lazily on
// first use (so charts that never open a panel never spin one up) and shared by
// every StatsPanel. Requests are correlated by an incrementing id; results are
// cached per ChartData object so reopening the *same* chart is instant, while a
// recompute (a fresh ChartData) recomputes.
import StatsWorker from '../workers/stats.worker.ts?worker&inline'
import type { StatsResponse } from '../workers/stats.worker'
import type { ChartData, SeriesProfile, CorrelationMatrix } from '../types'

export type StatsResult = {
  seriesProfiles: SeriesProfile[]
  correlation?: CorrelationMatrix
}

let worker: Worker | null = null
let nextId = 0
const pending = new Map<number, (r: StatsResult) => void>()
// WeakMap so cached results don't pin ChartData objects in memory after the chart
// is replaced by a recompute — they're collected with their key.
const cache = new WeakMap<ChartData, Promise<StatsResult>>()

function getWorker(): Worker {
  if (!worker) {
    worker = new StatsWorker()
    worker.onmessage = (e: MessageEvent<StatsResponse>) => {
      const { id, seriesProfiles, correlation } = e.data
      const resolve = pending.get(id)
      if (resolve) {
        pending.delete(id)
        resolve({ seriesProfiles, correlation })
      }
    }
  }
  return worker
}

export function computeStats(chartData: ChartData): Promise<StatsResult> {
  const cached = cache.get(chartData)
  if (cached) return cached

  const promise = new Promise<StatsResult>((resolve) => {
    const id = nextId++
    pending.set(id, resolve)
    // points/yAxis/series are plain (markRaw) → structured-clone takes them directly.
    getWorker().postMessage({
      type: 'compute',
      id,
      points: chartData.points,
      yAxis: chartData.yAxis,
      seriesOrder: chartData.series.map((s) => s.xAxis),
    })
  })

  cache.set(chartData, promise)
  return promise
}
