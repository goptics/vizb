import { ref, watch, unref, markRaw, onScopeDispose, type MaybeRef, type Ref } from 'vue'
import type { AxisLabels, DataPoint, ChartData, Sort, ScaleType, Axis } from '../types'
import TransformWorker from '../workers/transform.worker.ts?worker&inline'
import type { WorkerResponse } from '../workers/transform.worker'
import { listChartSignatures } from '../lib/transform'

// The arrangement the worker projects/groups under: present source axes in
// canonical order (identityString, e.g. "nx") and the selected target order
// (targetString, e.g. "yx"). Same length and value set.
export type Arrangement = { identityString: string; targetString: string }

// One chart's reactive slot. `key` is the stat signature — the chart's stable
// identity across swap/sort/group within a dataset — so the ChartCard keyed by
// it persists (echarts instance + double-buffer survive recomputes). `data` is
// null only until the chart's first job lands.
export type ChartState = {
  key: string
  title: string
  data: ChartData | null
  pending: boolean
}

// Runs the grouping/3D/sort transform in a Web Worker, one chart at a time. The
// worker owns the full raw dataset, so only one kind of change costs a clone:
//   - dataset change: the rows are new, so we clone them into the worker once
//     (`init`) and recompute every chart.
// Everything else is a param toggle against the worker's cache — NO clone, no
// main-thread data work:
//   - arrangement (swap): post `setArrangement`; the worker re-projects/re-groups
//     its cached raw data off-thread and replies `ready` with the new groupNames.
//   - group / sort / scale / showLabels: re-queue compute jobs against the cache.
// Jobs drain serially; each chart reveals the moment its job returns (progressive)
// and each ChartCard drives its own skeleton off its `pending`.
export function useChartPipeline(
  rawData: Ref<DataPoint[]> | DataPoint[],
  arrangement: Ref<Arrangement>,
  labels: MaybeRef<AxisLabels | undefined> | undefined,
  activeGroupId: Ref<number>,
  sort: Ref<Sort>,
  showLabels: Ref<boolean>,
  scale: Ref<ScaleType>,
  threeD: Ref<boolean>,
  axes?: MaybeRef<Axis[] | undefined>
) {
  const charts = ref<ChartState[]>([])
  // True once any chart has data — gates the first-load full-page skeleton.
  const hasAny = ref(false)
  // The worker's group list from the last `ready`. Drives the group selector and
  // the URL router; updated on init and on every arrangement change.
  const groupNames = ref<string[]>([])
  // Resolve the active group name synchronously from the pipeline's own groupNames.
  // Reading here (not from a Ref<string> input) means `ready` can call pumpQueue()
  // immediately and get the correct name without waiting for a downstream watcher to
  // copy groupNames into useDataPoint and transition activeGroupName '' → groupNames[0].
  const currentGroupName = () => groupNames.value[activeGroupId.value] ?? groupNames.value[0] ?? ''

  const worker = new TransformWorker()
  // dataEpoch: the worker's cached-dataset identity, bumped only on a dataset
  // change. jobEpoch: the current compute batch, bumped on every flush. Replies
  // are dropped unless both match — so neither a job for a superseded dataset nor
  // one from a superseded batch ever renders.
  let dataEpoch = 0
  let jobEpoch = 0
  // Signatures from the last `ready` — lets a params-only flush re-queue compute
  // jobs without waiting on a fresh init/arrangement.
  let lastSignatures: { signature: string; title: string }[] = []
  // True from sending `init`/`setArrangement` until its `ready` lands. While set,
  // a params flush does nothing: the pending `ready` will queue the jobs, and each
  // job reads the live params at send time, so it already picks up the change.
  let readyInFlight = false
  // True while a data change is debounced but its re-init hasn't run yet. Suppresses
  // a params recompute that would otherwise fire a job against the soon-to-be-stale
  // cached dataset; the imminent re-init recomputes everything with live params.
  let dataPending = false
  // FIFO of signatures still to compute for the current batch.
  const queue: string[] = []
  let draining = false

  const rows = () => (Array.isArray(rawData) ? rawData : rawData.value)

  // Plain (proxy-free) copy of the sort. postMessage structured-clone rejects
  // Vue reactive Proxies, and sort.value is one (it comes off the reactive
  // settings store) — so spread it into a fresh object before posting.
  const currentSort = (): Sort => ({ enabled: sort.value.enabled, order: sort.value.order })

  const pumpQueue = () => {
    if (draining) return
    const signature = queue.shift()
    if (signature === undefined) return
    draining = true
    worker.postMessage({
      type: 'compute',
      dataEpoch,
      jobEpoch,
      signature,
      groupName: currentGroupName(),
      sort: currentSort(),
      showLabels: showLabels.value,
      scale: scale.value,
      threeD: threeD.value,
    })
  }

  worker.onmessage = (e: MessageEvent<WorkerResponse>) => {
    const msg = e.data

    if (msg.type === 'ready') {
      if (msg.dataEpoch !== dataEpoch) return // superseded by a newer dataset
      readyInFlight = false
      lastSignatures = msg.signatures
      groupNames.value = msg.groupNames
      // Reconcile slots to the worker's signature list: keep existing rows (so
      // their old chart stays visible while the new one computes), add new ones,
      // drop gone ones. Mark everything pending and queue a job per chart.
      const prev = new Map(charts.value.map((c) => [c.key, c]))
      charts.value = msg.signatures.map(({ signature, title }) => {
        const existing = prev.get(signature)
        return {
          key: signature,
          title,
          data: existing?.data ?? null,
          pending: true,
        }
      })
      queue.length = 0
      queue.push(...msg.signatures.map((s) => s.signature))
      draining = false
      pumpQueue()
      return
    }

    // chart: free the drain and pump the next regardless, so a dropped (stale)
    // reply still advances the queue. Only store/reveal when the reply belongs to
    // the current dataset AND the current batch.
    draining = false
    if (msg.dataEpoch === dataEpoch && msg.jobEpoch === jobEpoch) {
      const slot = charts.value.find((c) => c.key === msg.signature)
      if (slot) {
        // Store the result raw — the ChartData holds the full point cloud / 3D
        // grid (up to 100k entries); deep-proxying it would tax every read echarts
        // and the option computeds make. Identity (key/title/pending) stays
        // reactive; the payload doesn't need to be.
        slot.data = markRaw(msg.chart)
        slot.pending = false
        hasAny.value = true
      }
    }
    pumpQueue()
  }

  // Post a new arrangement to the worker (no re-clone). The worker re-projects and
  // re-groups its cached raw data off-thread and replies `ready`, which re-queues
  // the compute jobs. Bumps the batch so in-flight replies from the old arrangement
  // are dropped.
  const setArrangement = () => {
    if (!lastSignatures.length || readyInFlight || dataPending) return
    readyInFlight = true
    // labels is a fresh plain object from swapAxisLabels (off now-plain axisLabels),
    // so postMessage clones it natively — no proxy stripping needed.
    worker.postMessage({
      type: 'setArrangement',
      identityString: arrangement.value.identityString,
      targetString: arrangement.value.targetString,
      labels: unref(labels) ?? null,
    })
  }

  // Recompute every chart against the worker's cached dataset — no re-clone. Used
  // when only group / sort / scale / showLabels changed.
  const recompute = () => {
    if (!lastSignatures.length || readyInFlight || dataPending) return
    queue.length = 0
    queue.push(...lastSignatures.map((s) => s.signature))
    pumpQueue()
  }

  // Clone the current dataset into the worker and (re)establish its signatures.
  // Heavy (a structured clone of every row) — only runs on a real dataset change.
  const reinit = () => {
    dataPending = false
    const data = rows()
    if (!data?.length) {
      charts.value = []
      lastSignatures = []
      groupNames.value = []
      hasAny.value = false
      readyInFlight = false
      return
    }

    dataEpoch++
    readyInFlight = true
    queue.length = 0
    draining = false

    // Pre-populate skeleton slots so ChartCards appear immediately while the worker
    // clones and groups the dataset. Signatures are arrangement-independent (raw
    // rows only), so they match what ready will return. The ready handler reconciles
    // via its own prev-map: same keys → same ChartCard instances, no flicker.
    const sigs = listChartSignatures(data)
    const prev = new Map(charts.value.map((c) => [c.key, c]))
    charts.value = sigs.map(({ signature, statTemplate }) => ({
      key: signature,
      title: statTemplate.type,
      data: prev.get(signature)?.data ?? null,
      pending: true,
    }))

    // The rows are kept non-reactive (shallowRef + markRaw in useDataPoint), so
    // postMessage's structured clone takes them directly — a single native pass,
    // no proxy stripping, no JSON stringify/parse round-trip. This was the 7s
    // bottleneck on large datasets.
    worker.postMessage({
      type: 'init',
      dataEpoch,
      data,
      identityString: arrangement.value.identityString,
      targetString: arrangement.value.targetString,
      labels: unref(labels) ?? null,
      axes: unref(axes) ?? undefined,
    })
  }

  // Bump the batch and flip charts to pending synchronously (not after the
  // debounce) so the card skeletons rise the instant inputs change.
  const startBatch = () => {
    jobEpoch++
    if (rows()?.length) for (const c of charts.value) c.pending = true
  }

  let dataDebounce: ReturnType<typeof setTimeout> | undefined
  let paramsDebounce: ReturnType<typeof setTimeout> | undefined
  let arrangeDebounce: ReturnType<typeof setTimeout> | undefined

  // Data path — fires only on a dataset change (identity-based, no deep watch).
  // This is the one path that re-clones the rows into the worker.
  watch(
    () => rows(),
    () => {
      dataPending = true
      startBatch()
      clearTimeout(dataDebounce)
      dataDebounce = setTimeout(reinit, 50)
    },
    { immediate: true }
  )

  // Arrangement path — fires on swap. The rows are unchanged, so we don't clone;
  // we tell the worker to re-project/re-group off-thread and re-queue on `ready`.
  watch(
    () => [arrangement.value.identityString, arrangement.value.targetString] as const,
    () => {
      startBatch()
      clearTimeout(arrangeDebounce)
      arrangeDebounce = setTimeout(setArrangement, 50)
    }
  )

  // Params path — fires on group / sort / scale / showLabels. Recompute off the
  // cached dataset; no clone, no init, no re-group.
  // NOTE: tracks activeGroupId (not a derived group-name string) so that the
  // groupNames [] → [...] transition on the first `ready` does NOT trip this watch
  // (activeGroupId stays 0). pumpQueue() inside `ready` already reads the live
  // currentGroupName() synchronously, so the first batch is correct without this
  // watch firing a redundant second recompute.
  watch(
    () =>
      [
        activeGroupId.value,
        sort.value.enabled,
        sort.value.order,
        showLabels.value,
        scale.value,
        threeD.value,
      ] as const,
    () => {
      startBatch()
      clearTimeout(paramsDebounce)
      paramsDebounce = setTimeout(recompute, 50)
    }
  )

  onScopeDispose(() => {
    clearTimeout(dataDebounce)
    clearTimeout(paramsDebounce)
    clearTimeout(arrangeDebounce)
    queue.length = 0
    worker.terminate()
  })

  return { charts, hasAny, groupNames }
}
