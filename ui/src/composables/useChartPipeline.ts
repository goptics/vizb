import { ref, watch, unref, markRaw, onScopeDispose, type MaybeRef, type Ref } from 'vue'
import type { AxisLabels, DataPoint, ChartData, Sort, ScaleType } from '../types'
import TransformWorker from '../workers/transform.worker.ts?worker&inline'
import type { WorkerResponse } from '../workers/transform.worker'

export type TriggerSwap = (currentKey: string, targetKey: string, newLabels?: AxisLabels) => void

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

// Runs the grouping/3D/sort transform in a Web Worker, one chart at a time. Two
// kinds of change drive it, and they cost very differently:
//   - data/labels change (dataset / group / swap): the rows are new, so we clone
//     them into the worker once (`init`) and recompute every chart.
//   - sort / scale / showLabels change: the rows are unchanged, so we DON'T
//     re-clone — the worker still holds the dataset, and we only enqueue compute
//     jobs against its cache. This is the hot path (a sort toggle on 100k points)
//     and avoids the multi-hundred-ms main-thread JSON clone of the reactive rows.
// Jobs drain serially; each chart reveals the moment its job returns (progressive)
// and each ChartCard drives its own skeleton off its `pending`.
export function useChartPipeline(
  results: Ref<DataPoint[]> | DataPoint[],
  axisLabels: MaybeRef<AxisLabels | undefined> | undefined,
  sort: Ref<Sort>,
  showLabels: Ref<boolean>,
  scale: Ref<ScaleType>
) {
  const charts = ref<ChartState[]>([])
  // True once any chart has data — gates the first-load full-page skeleton.
  const hasAny = ref(false)

  const worker = new TransformWorker()
  // dataEpoch: the worker's cached-dataset identity, bumped only on a data/labels
  // change. jobEpoch: the current compute batch, bumped on every flush. Replies
  // are dropped unless both match — so neither a job for a superseded dataset nor
  // one from a superseded sort/scale batch ever renders.
  let dataEpoch = 0
  let jobEpoch = 0
  // Signatures from the last `ready` — lets a params-only flush re-queue compute
  // jobs without waiting on a fresh init.
  let lastSignatures: { signature: string; title: string }[] = []
  // True from sending `init` until its `ready` lands. While set, a params flush
  // does nothing: the pending `ready` will queue the jobs, and each job reads the
  // live sort/scale/showLabels at send time, so it already picks up the change.
  let initInFlight = false
  // True while a data change is debounced but its re-init hasn't run yet. Suppresses
  // a params recompute that would otherwise fire a job against the soon-to-be-stale
  // cached dataset; the imminent re-init recomputes everything with live params.
  let dataPending = false
  // Set by triggerSwap before the caller mutates axisLabels/data. The data watcher
  // fires once from those mutations; this flag tells it to skip reinit (the worker
  // already received the swap message and will send `ready` directly).
  let swapPending = false
  // FIFO of signatures still to compute for the current batch.
  const queue: string[] = []
  let draining = false

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
      sort: currentSort(),
      showLabels: showLabels.value,
      scale: scale.value,
    })
  }

  worker.onmessage = (e: MessageEvent<WorkerResponse>) => {
    const msg = e.data

    if (msg.type === 'ready') {
      console.log('[pipe] ready dataEpoch=', msg.dataEpoch, 'local=', dataEpoch, 'sigs=', msg.signatures.length)
      if (msg.dataEpoch !== dataEpoch) return // superseded by a newer dataset
      initInFlight = false
      lastSignatures = msg.signatures
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
    console.log('[pipe] chart reply sig=', msg.signature, 'dE=', msg.dataEpoch, '/', dataEpoch, 'jE=', msg.jobEpoch, '/', jobEpoch, 'series0=', JSON.stringify(msg.chart.series?.[0]?.xAxis), 'nSeries=', msg.chart.series?.length)
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

  // Send a cheap swap to the worker (no re-clone). The worker applies the same O(n)
  // field-rename to its cached data and sends `ready`; the caller is expected to
  // apply the matching mutation to the main-thread store so the two copies stay in
  // sync. The data watcher fires from the axisLabels replacement that follows; the
  // swapPending flag tells it to skip reinit for that one fire.
  const triggerSwap: TriggerSwap = (currentKey, targetKey, newLabels) => {
    console.log('[pipe] triggerSwap', currentKey, '->', targetKey)
    swapPending = true
    startBatch()
    worker.postMessage({ type: 'swap', currentKey, targetKey, labels: newLabels ?? null })
  }

  // Recompute every chart against the worker's cached dataset — no re-clone. Used
  // when only sort / scale / showLabels changed.
  const recompute = () => {
    if (!lastSignatures.length || initInFlight || dataPending) return
    queue.length = 0
    queue.push(...lastSignatures.map((s) => s.signature))
    pumpQueue()
  }

  // Clone the current dataset into the worker and (re)establish its signatures.
  // Heavy (a structured clone of every row) — only runs on a real data change.
  const reinit = () => {
    dataPending = false
    const data = Array.isArray(results) ? results : results.value
    if (!data?.length) {
      charts.value = []
      lastSignatures = []
      hasAny.value = false
      initInFlight = false
      return
    }

    dataEpoch++
    initInFlight = true
    queue.length = 0
    draining = false

    // Send a clone-safe plain copy of the dataset. structured-clone (used by
    // postMessage) rejects Vue reactive Proxies — round-trip through JSON.
    worker.postMessage(
      JSON.parse(
        JSON.stringify({
          type: 'init',
          dataEpoch,
          data,
          labels: unref(axisLabels) ?? null,
        })
      )
    )
  }

  // Bump the batch and flip charts to pending synchronously (not after the
  // debounce) so the card skeletons rise the instant inputs change.
  const startBatch = () => {
    jobEpoch++
    const data = Array.isArray(results) ? results : results.value
    if (data?.length) for (const c of charts.value) c.pending = true
  }

  let dataDebounce: ReturnType<typeof setTimeout> | undefined
  let paramsDebounce: ReturnType<typeof setTimeout> | undefined

  // Data path — fires on dataset / group change (identity-based, no deep watch).
  // Also fires on swap (axisLabels replacement), but swapPending suppresses reinit
  // for that one fire — the worker already holds the swapped dataset via `swap` msg.
  watch(
    () => [Array.isArray(results) ? results : results.value, unref(axisLabels)] as const,
    () => {
      console.log('[pipe] data watcher fired, swapPending=', swapPending)
      if (swapPending) {
        swapPending = false
        return
      }
      dataPending = true
      startBatch()
      clearTimeout(dataDebounce)
      dataDebounce = setTimeout(reinit, 50)
    },
    { immediate: true }
  )

  // Params path — fires on sort / scale / showLabels. Recompute off the cached
  // dataset; no clone, no init.
  watch(
    () => [sort.value.enabled, sort.value.order, showLabels.value, scale.value] as const,
    () => {
      startBatch()
      clearTimeout(paramsDebounce)
      paramsDebounce = setTimeout(recompute, 50)
    }
  )

  onScopeDispose(() => {
    clearTimeout(dataDebounce)
    clearTimeout(paramsDebounce)
    queue.length = 0
    worker.terminate()
  })

  return { charts, hasAny, triggerSwap }
}
