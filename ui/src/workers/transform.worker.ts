// Dedicated Web Worker that runs the heavy chart transform off the main thread,
// one chart at a time. Bundled and base64-inlined via `?worker&inline` so the
// single-file HTML output stays self-contained (no external worker asset).
//
// Protocol (stateful — the dataset is cached so a sort toggle doesn't re-clone
// it):
//   init    {epoch,data,labels}        -> ready {epoch, signatures}
//   compute {epoch,signature,sort}     -> chart {epoch, signature, chart}
// Every message carries an `epoch`; the worker ignores anything that doesn't
// match the last init, so stale jobs from a superseded input never reply.
import { listChartSignatures, buildChartForSignature, type ChartSignature } from '../lib/transform'
import type { DataPoint, AxisLabels, Sort, ChartData } from '../types'

export type InitMessage = { type: 'init'; epoch: number; data: DataPoint[]; labels?: AxisLabels }
export type ComputeMessage = { type: 'compute'; epoch: number; signature: string; sort: Sort }
export type WorkerRequest = InitMessage | ComputeMessage

export type ReadyMessage = {
  type: 'ready'
  epoch: number
  signatures: { signature: string; title: string }[]
}
export type ChartMessage = { type: 'chart'; epoch: number; signature: string; chart: ChartData }
export type WorkerResponse = ReadyMessage | ChartMessage

type State = {
  epoch: number
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
      epoch: msg.epoch,
      data: msg.data,
      labels: msg.labels,
      bySignature: new Map(signatures.map((s) => [s.signature, s])),
    }
    post({
      type: 'ready',
      epoch: msg.epoch,
      signatures: signatures.map((s) => ({ signature: s.signature, title: s.statTemplate.type })),
    })
    return
  }

  // compute: ignore jobs from a superseded dataset.
  if (!state || msg.epoch !== state.epoch) return
  const entry = state.bySignature.get(msg.signature)
  if (!entry) return

  const chart = buildChartForSignature(
    state.data,
    entry.signature,
    entry.statTemplate,
    state.labels,
    msg.sort
  )
  post({ type: 'chart', epoch: msg.epoch, signature: msg.signature, chart })
}
