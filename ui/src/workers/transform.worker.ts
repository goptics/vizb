// Dedicated Web Worker that runs the heavy chart transform off the main thread,
// one chart at a time. Bundled and base64-inlined via `?worker&inline` so the
// single-file HTML output stays self-contained (no external worker asset).
//
// Protocol (stateful — the dataset is cached so a sort/scale toggle recomputes
// straight off the cache, without re-cloning the rows):
//   init    {dataEpoch,data,labels}                 -> ready {dataEpoch, signatures}
//   compute {dataEpoch,jobEpoch,signature,sort,...}  -> chart {dataEpoch, jobEpoch, signature, chart}
// Two counters guard staleness independently:
//   dataEpoch — the cached dataset's identity. Bumped only when the rows/labels
//     change (dataset / group / swap). `compute` is dropped unless it matches the
//     cached dataset, so a job for a superseded dataset never replies.
//   jobEpoch  — the compute batch's identity. Bumped on every recompute (sort,
//     scale, labels). Echoed back so the main thread can drop a reply from a
//     superseded batch (e.g. an in-flight sort=asc job after the user flipped to
//     desc) without it flashing on screen.
import { listChartSignatures, buildChartForSignature, type ChartSignature } from '../lib/transform'
import { translateAxisKey, swapAxisFields } from '../lib/swap'
import type { DataPoint, AxisLabels, Sort, ChartData, ScaleType } from '../types'

export type InitMessage = { type: 'init'; dataEpoch: number; data: DataPoint[]; labels?: AxisLabels }
export type ComputeMessage = { type: 'compute'; dataEpoch: number; jobEpoch: number; signature: string; sort: Sort; showLabels: boolean; scale: ScaleType }
// Cheaper than re-cloning: mutates the cached dataset in place and re-indexes.
export type SwapMessage = { type: 'swap'; currentKey: string; targetKey: string; labels?: AxisLabels | null }
export type WorkerRequest = InitMessage | ComputeMessage | SwapMessage

export type ReadyMessage = {
  type: 'ready'
  dataEpoch: number
  signatures: { signature: string; title: string }[]
}
export type ChartMessage = { type: 'chart'; dataEpoch: number; jobEpoch: number; signature: string; chart: ChartData }
export type WorkerResponse = ReadyMessage | ChartMessage

type State = {
  dataEpoch: number
  data: DataPoint[]
  labels?: AxisLabels
  bySignature: Map<string, ChartSignature>
}

let state: State | null = null

const post = (msg: WorkerResponse) => (self as unknown as Worker).postMessage(msg)

self.onmessage = (e: MessageEvent<WorkerRequest>) => {
  const msg = e.data

  if (msg.type === 'init') {
    const signatures = listChartSignatures(msg.data)
    state = {
      dataEpoch: msg.dataEpoch,
      data: msg.data,
      labels: msg.labels,
      bySignature: new Map(signatures.map((s) => [s.signature, s])),
    }
    post({
      type: 'ready',
      dataEpoch: msg.dataEpoch,
      signatures: signatures.map((s) => ({ signature: s.signature, title: s.statTemplate.type })),
    })
    return
  }

  if (msg.type === 'swap') {
    if (!state) return
    console.log('[worker] swap', msg.currentKey, '->', msg.targetKey, 'row0 before=', JSON.stringify({ ...state.data[0], stats: undefined }))
    const currentKeys = translateAxisKey(msg.currentKey)
    const targetKeys = translateAxisKey(msg.targetKey)
    swapAxisFields(state.data, currentKeys, targetKeys)
    console.log('[worker] swap row0 after=', JSON.stringify({ ...state.data[0], stats: undefined }))
    if (msg.labels !== undefined) state.labels = msg.labels ?? undefined
    const signatures = listChartSignatures(state.data)
    state.bySignature = new Map(signatures.map((s) => [s.signature, s]))
    post({
      type: 'ready',
      dataEpoch: state.dataEpoch,
      signatures: signatures.map((s) => ({ signature: s.signature, title: s.statTemplate.type })),
    })
    return
  }

  // compute: ignore jobs aimed at a superseded dataset (a newer init landed).
  if (!state || msg.dataEpoch !== state.dataEpoch) return
  const entry = state.bySignature.get(msg.signature)
  if (!entry) return

  const chart = buildChartForSignature(
    state.data,
    entry.signature,
    entry.statTemplate,
    state.labels,
    msg.sort,
    msg.showLabels,
    msg.scale
  )
  post({ type: 'chart', dataEpoch: msg.dataEpoch, jobEpoch: msg.jobEpoch, signature: msg.signature, chart })
}
