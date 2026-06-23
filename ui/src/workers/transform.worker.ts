// Dedicated Web Worker that runs the heavy chart transform off the main thread,
// one chart at a time. Bundled and base64-inlined via `?worker&inline` so the
// embedded HTML template stays self-contained (no external worker asset).
//
// The worker owns the FULL raw dataset and treats arrangement + group + sort +
// scale + showLabels all as compute params — the rows never change on a setting
// toggle, only how they're read. So the main thread never recomputes: it posts a
// param and renders the returned ChartData. The only data clone is the one-time
// `init` (and again only on a dataset switch).
//
// Protocol (stateful):
//   init           {dataEpoch,data,identityString,targetString,labels}
//                    -> ready {dataEpoch, signatures, groupNames}
//   setArrangement {identityString,targetString,labels}
//                    -> ready {dataEpoch, signatures, groupNames}   (no re-clone)
//   compute        {dataEpoch,jobEpoch,signature,groupName,sort,showLabels,scale}
//                    -> chart {dataEpoch, jobEpoch, signature, chart}
//
// Two counters guard staleness independently:
//   dataEpoch — the cached dataset's identity. Bumped only when the rows change
//     (a dataset switch). `compute`/`setArrangement` are dropped unless they match
//     the cached dataset, so work for a superseded dataset never replies.
//   jobEpoch  — the compute batch's identity. Bumped on every recompute (arrangement,
//     group, sort, scale, labels). Echoed back so the main thread can drop a reply
//     from a superseded batch without it flashing on screen.
import {
  listChartSignatures,
  buildChartForSignature,
  buildValueModeChart,
  canonicalAxisOrdersFromStrings,
  projectAndGroup,
  type ChartSignature,
} from '../lib/transform'
import { isValueChartType, isValueMode } from '../lib/utils'
import { translateAxisKey } from '../lib/swap'
import type { DataPoint, AxisLabels, Sort, ChartData, ScaleType, Axis, ChartType } from '../types'

export type InitMessage = {
  type: 'init'
  dataEpoch: number
  data: DataPoint[]
  identityString: string
  targetString: string
  labels?: AxisLabels
  axes?: Axis[] // present when axes metadata is available from the dataset
  chartType?: ChartType
}
export type SetArrangementMessage = {
  type: 'setArrangement'
  identityString: string
  targetString: string
  labels?: AxisLabels | null
}
export type ComputeMessage = {
  type: 'compute'
  dataEpoch: number
  jobEpoch: number
  signature: string
  groupName: string
  sort: Sort
  showLabels: boolean
  scale: ScaleType
  threeD: boolean
}
export type WorkerRequest = InitMessage | SetArrangementMessage | ComputeMessage

export type ReadyMessage = {
  type: 'ready'
  dataEpoch: number
  signatures: { signature: string; title: string }[]
  groupNames: string[]
}
export type ChartMessage = {
  type: 'chart'
  dataEpoch: number
  jobEpoch: number
  signature: string
  chart: ChartData
}
export type WorkerResponse = ReadyMessage | ChartMessage

type State = {
  dataEpoch: number
  raw: DataPoint[]
  identityString: string
  targetString: string
  grouped: Map<string, DataPoint[]>
  groupNames: string[]
  labels?: AxisLabels
  bySignature: Map<string, ChartSignature>
  valueMode: boolean
  axes?: Axis[]
  chartType?: ChartType
}

let state: State | null = null

const post = (msg: WorkerResponse) => (self as unknown as Worker).postMessage(msg)

// Project + group the raw dataset under an arrangement, and re-derive the stat
// signatures (signatures are arrangement-independent, computed off the raw rows).
const applyArrangement = (s: State, identityString: string, targetString: string) => {
  s.identityString = identityString
  s.targetString = targetString
  const { grouped, groupNames } = projectAndGroup(
    s.raw,
    translateAxisKey(identityString),
    translateAxisKey(targetString)
  )
  s.grouped = grouped
  s.groupNames = groupNames
  const signatures = listChartSignatures(s.raw)
  s.bySignature = new Map(signatures.map((sig) => [sig.signature, sig]))
}

const readyReply = (s: State): ReadyMessage => ({
  type: 'ready',
  dataEpoch: s.dataEpoch,
  signatures: Array.from(s.bySignature.values()).map((sig) => ({
    signature: sig.signature,
    title: sig.statTemplate.type,
  })),
  groupNames: s.groupNames,
})

const valueModeReadyReply = (s: State): ReadyMessage => ({
  type: 'ready',
  dataEpoch: s.dataEpoch,
  signatures: [
    {
      signature: '__value_mode__',
      title: buildValueModeChart([], s.axes ?? [], s.identityString, s.targetString).title,
    },
  ],
  groupNames: [],
})

self.onmessage = (e: MessageEvent<WorkerRequest>) => {
  const msg = e.data

  switch (msg.type) {
    case 'init': {
      const axes = msg.axes
      const chartType = msg.chartType
      const valueModeDetected = isValueChartType(chartType) && isValueMode(axes)

      state = {
        dataEpoch: msg.dataEpoch,
        raw: msg.data,
        identityString: msg.identityString,
        targetString: msg.targetString,
        grouped: new Map(),
        groupNames: [],
        labels: msg.labels,
        bySignature: new Map(),
        valueMode: valueModeDetected,
        axes,
        chartType,
      }

      if (valueModeDetected) {
        state.bySignature.set('__value_mode__', {
          signature: '__value_mode__',
          statTemplate: { type: 'value' },
        })
        post(valueModeReadyReply(state))
        return
      }

      applyArrangement(state, msg.identityString, msg.targetString)
      post(readyReply(state))
      return
    }
    case 'setArrangement': {
      if (!state) return
      if (msg.labels !== undefined) state.labels = msg.labels ?? undefined
      // Swap/arrangement is a no-op in value mode (coordinates are fixed).
      // applyArrangement would call listChartSignatures on stats-less rows and
      // wipe the synthetic __value_mode__ signature.
      if (state.valueMode) {
        state.identityString = msg.identityString
        state.targetString = msg.targetString
        post(valueModeReadyReply(state))
        return
      }
      applyArrangement(state, msg.identityString, msg.targetString)
      post(readyReply(state))
      return
    }
  }

  // compute: ignore jobs aimed at a superseded dataset (a newer init landed).
  if (!state || msg.dataEpoch !== state.dataEpoch) return

  // Value-mode compute: bypass grouping/signature/stat pipeline entirely.
  if (state.valueMode && msg.signature === '__value_mode__' && state.axes) {
    const chart = buildValueModeChart(
      state.raw,
      state.axes,
      state.identityString,
      state.targetString,
      { scale: msg.scale, showLabels: msg.showLabels, threeD: msg.threeD }
    )
    post({
      type: 'chart',
      dataEpoch: msg.dataEpoch,
      jobEpoch: msg.jobEpoch,
      signature: msg.signature,
      chart,
    })
    return
  }

  const entry = state.bySignature.get(msg.signature)
  if (!entry) return

  // Read the selected group's rows; fall back to the first group (or empty) so a
  // stale/unknown groupName still produces a renderable chart instead of dropping.
  const rows = state.grouped.get(msg.groupName) ?? state.grouped.values().next().value ?? []

  const canonical = canonicalAxisOrdersFromStrings(
    state.raw,
    state.identityString,
    state.targetString
  )

  const chart = buildChartForSignature(
    rows,
    entry.signature,
    entry.statTemplate,
    state.labels,
    msg.sort,
    msg.showLabels,
    msg.scale,
    canonical,
    msg.threeD
  )
  post({
    type: 'chart',
    dataEpoch: msg.dataEpoch,
    jobEpoch: msg.jobEpoch,
    signature: msg.signature,
    chart,
  })
}
