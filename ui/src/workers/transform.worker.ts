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
  buildMixedModeChart,
  canonicalAxisOrdersFromStrings,
  projectAndGroup,
  type ChartSignature,
} from '../lib/transform'
import { isValueChartType, isValueMode, isMixedMode } from '../lib/utils'
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
  preserveRows?: boolean
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
  axes?: Axis[]
  chartType?: ChartType
  preserveRows: boolean
  handler: ModeHandler
}

let state: State | null = null

const post = (msg: WorkerResponse) => (self as unknown as Worker).postMessage(msg)

// A mode handler owns the ready/compute behaviour for one chart pipeline shape.
// `init` seeds state + returns the first ready reply; `setArrangement` reacts to
// a swap/arrangement change; `compute` builds one chart for a posted job. Picking
// the handler once at init collapses the prior valueMode/mixedMode/normal branch
// triplication into a single polymorphic dispatch.
interface ModeHandler {
  init(s: State, msg: InitMessage): ReadyMessage
  setArrangement(s: State, msg: SetArrangementMessage): ReadyMessage
  compute(s: State, msg: ComputeMessage): void
  readyReply(s: State): ReadyMessage
}

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

const postChart = (s: State, msg: ComputeMessage, chart: ChartData) =>
  post({
    type: 'chart',
    dataEpoch: s.dataEpoch,
    jobEpoch: msg.jobEpoch,
    signature: msg.signature,
    chart,
  })

// Grouped (default) pipeline: project+group rows, one chart per stat signature.
const GroupedHandler: ModeHandler = {
  init(s, msg) {
    applyArrangement(s, msg.identityString, msg.targetString)
    return readyReply(s)
  },
  setArrangement(s, msg) {
    if (msg.labels !== undefined) s.labels = msg.labels ?? undefined
    applyArrangement(s, msg.identityString, msg.targetString)
    return readyReply(s)
  },
  readyReply,
  compute(s, msg) {
    const entry = s.bySignature.get(msg.signature)
    if (!entry) return

    // Read the selected group's rows; fall back to the first group (or empty) so a
    // stale/unknown groupName still produces a renderable chart instead of dropping.
    const rows = s.grouped.get(msg.groupName) ?? s.grouped.values().next().value ?? []

    const canonical = canonicalAxisOrdersFromStrings(s.raw, s.identityString, s.targetString)

    const chart = buildChartForSignature(
      rows,
      entry.signature,
      entry.statTemplate,
      s.labels,
      msg.sort,
      msg.showLabels,
      msg.scale,
      canonical,
      msg.threeD,
      s.preserveRows
    )
    postChart(s, msg, chart)
  },
}

// Value-mode pipeline: continuous numeric axes, one synthetic __value_mode__
// chart. Swap/arrangement is a no-op on the rows (coordinates are fixed) but the
// identity/target strings still drive the value-mode title + 3D projection.
const ValueHandler: ModeHandler = {
  init(s) {
    s.bySignature.set('__value_mode__', {
      signature: '__value_mode__',
      statTemplate: { type: 'value' },
    })
    return this.readyReply(s)
  },
  setArrangement(s, msg) {
    if (msg.labels !== undefined) s.labels = msg.labels ?? undefined
    s.identityString = msg.identityString
    s.targetString = msg.targetString
    return this.readyReply(s)
  },
  compute(s, msg) {
    if (msg.signature !== '__value_mode__' || !s.axes) return
    const chart = buildValueModeChart(s.raw, s.axes, s.identityString, s.targetString, {
      scale: msg.scale,
      showLabels: msg.showLabels,
      threeD: msg.threeD,
    })
    postChart(s, msg, chart)
  },
  readyReply(s: State): ReadyMessage {
    return {
      type: 'ready',
      dataEpoch: s.dataEpoch,
      signatures: [
        {
          signature: '__value_mode__',
          title: buildValueModeChart([], s.axes ?? [], s.identityString, s.targetString).title,
        },
      ],
      groupNames: [],
    }
  },
}

// Mixed-axis pipeline: category x + value y[,z], one synthetic __mixed_mode__
// chart. No grouping/stat pipeline; arrangement is a no-op.
const MixedHandler: ModeHandler = {
  init(s) {
    s.bySignature.set('__mixed_mode__', {
      signature: '__mixed_mode__',
      statTemplate: { type: 'mixed' },
    })
    return this.readyReply(s)
  },
  setArrangement(s, msg) {
    if (msg.labels !== undefined) s.labels = msg.labels ?? undefined
    return this.readyReply(s)
  },
  compute(s, msg) {
    if (msg.signature !== '__mixed_mode__' || !s.axes) return
    const chart = buildMixedModeChart(s.raw, s.axes, {
      scale: msg.scale,
      showLabels: msg.showLabels,
    })
    postChart(s, msg, chart)
  },
  readyReply(s: State): ReadyMessage {
    return {
      type: 'ready',
      dataEpoch: s.dataEpoch,
      signatures: [
        {
          signature: '__mixed_mode__',
          title: buildMixedModeChart([], s.axes ?? []).title,
        },
      ],
      groupNames: [],
    }
  },
}

// Pick the handler once at init based on the dataset's axis shape.
const pickHandler = (axes: Axis[] | undefined, chartType?: ChartType): ModeHandler => {
  if (isValueChartType(chartType) && isValueMode(axes)) return ValueHandler
  if (isValueChartType(chartType) && isMixedMode(axes)) return MixedHandler
  return GroupedHandler
}

self.onmessage = (e: MessageEvent<WorkerRequest>) => {
  const msg = e.data

  switch (msg.type) {
    case 'init': {
      const handler = pickHandler(msg.axes, msg.chartType)
      state = {
        dataEpoch: msg.dataEpoch,
        raw: msg.data,
        identityString: msg.identityString,
        targetString: msg.targetString,
        grouped: new Map(),
        groupNames: [],
        labels: msg.labels,
        bySignature: new Map(),
        axes: msg.axes,
        chartType: msg.chartType,
        preserveRows: msg.preserveRows === true,
        handler,
      }
      post(handler.init(state, msg))
      return
    }
    case 'setArrangement': {
      if (!state) return
      post(state.handler.setArrangement(state, msg))
      return
    }
  }

  // compute: ignore jobs aimed at a superseded dataset (a newer init landed).
  if (!state || msg.dataEpoch !== state.dataEpoch) return
  state.handler.compute(state, msg)
}
