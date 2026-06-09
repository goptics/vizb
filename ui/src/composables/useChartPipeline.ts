import { ref, watch, unref, onScopeDispose, type MaybeRef, type Ref } from 'vue'
import type { AxisLabels, DataPoint, ChartData, Sort } from '../types'
import TransformWorker from '../workers/transform.worker.ts?worker&inline'
import type { WorkerResponse } from '../workers/transform.worker'

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

// Runs the grouping/3D transform in a Web Worker, one chart at a time. On any
// input change we (re)init the worker with the dataset, then enqueue a compute
// job per chart and drain the queue serially — each chart reveals the moment its
// job returns (progressive), and each ChartCard drives its own skeleton off its
// `pending`. The dataset is cloned into the worker once per data change, not per
// sort toggle (the worker caches it).
export function useChartPipeline(
  results: Ref<DataPoint[]> | DataPoint[],
  axisLabels: MaybeRef<AxisLabels | undefined> | undefined,
  sort: Ref<Sort>,
  showLabels: Ref<boolean>
) {
  const charts = ref<ChartState[]>([])
  // True any chart already has data — gates the first-load full-page skeleton.
  const hasAny = ref(false)

  const worker = new TransformWorker()
  // Bumped on every input change. Replies tagged with a stale epoch are dropped,
  // so fast successive swaps/sorts never render an out-of-date result.
  let epoch = 0
  // FIFO of signatures still to compute for the current epoch.
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
    worker.postMessage({ type: 'compute', epoch, signature, sort: currentSort(), showLabels: showLabels.value })
  }

  worker.onmessage = (e: MessageEvent<WorkerResponse>) => {
    const msg = e.data
    if (msg.epoch !== epoch) return // stale dataset/job — ignore

    if (msg.type === 'ready') {
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

    // chart: store the result on its slot, then run the next job.
    const slot = charts.value.find((c) => c.key === msg.signature)
    if (slot) {
      slot.data = msg.chart
      slot.pending = false
      hasAny.value = true
    }
    draining = false
    pumpQueue()
  }

  const post = () => {
    const data = Array.isArray(results) ? results : results.value
    epoch++
    queue.length = 0
    draining = false

    if (!data?.length) {
      charts.value = []
      hasAny.value = false
      return
    }

    // Send a clone-safe plain copy of the dataset. structured-clone (used by
    // postMessage) rejects Vue reactive Proxies — round-trip through JSON.
    worker.postMessage(
      JSON.parse(
        JSON.stringify({
          type: 'init',
          epoch,
          data,
          labels: unref(axisLabels) ?? null,
        })
      )
    )
  }

  let debounce: ReturnType<typeof setTimeout> | undefined
  watch(
    () => [Array.isArray(results) ? results : results.value, unref(axisLabels), sort.value, showLabels.value] as const,
    () => {
      // Flip every chart to pending synchronously (not after the debounce) so the
      // card skeletons rise the instant inputs change, not 50ms + a round-trip
      // later. `post` then re-inits and re-queues.
      const data = Array.isArray(results) ? results : results.value
      if (data?.length) for (const c of charts.value) c.pending = true
      clearTimeout(debounce)
      debounce = setTimeout(post, 50)
    },
    { deep: true, immediate: true }
  )

  onScopeDispose(() => {
    clearTimeout(debounce)
    queue.length = 0
    worker.terminate()
  })

  return { charts, hasAny }
}
