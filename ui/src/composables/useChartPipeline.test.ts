// ui/src/composables/useChartPipeline.test.ts
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { effectScope, ref, type Ref } from 'vue'
import type { WorkerResponse, ReadyMessage, ChartMessage } from '../workers/transform.worker'
import type { DataPoint, AxisLabels, Sort, ScaleType, ChartData } from '../types'

const ctorSpy = vi.fn()
class TrackedMockWorker {
  static instances: TrackedMockWorker[] = []
  onmessage: ((e: MessageEvent) => void) | null = null
  postMessage = vi.fn()
  terminate = vi.fn()
  __emit = (data: unknown) => this.onmessage?.({ data } as MessageEvent)
  constructor() {
    ctorSpy()
    TrackedMockWorker.instances.push(this)
  }
}

vi.mock('../workers/transform.worker.ts?worker&inline', () => ({
  default: TrackedMockWorker,
}))

const { useChartPipeline } = await import('./useChartPipeline')

const noSort: Sort = { enabled: false, order: 'asc' }

function dp(xAxis: string, yAxis = '', zAxis = '', type = 'val', value = 1): DataPoint {
  return { xAxis, yAxis, zAxis, stats: [{ type, value }] }
}

const defaultLabels: AxisLabels = { x: 'X', y: 'Y', z: 'Z' }

let scope: ReturnType<typeof effectScope>
let worker: TrackedMockWorker
let result: ReturnType<typeof useChartPipeline>
let rawData: Ref<DataPoint[]>
let arrangement: Ref<{ identityString: string; targetString: string }>
let activeGroupId: Ref<number>
let sort: Ref<Sort>
let showLabels: Ref<boolean>
let scale: Ref<ScaleType>
let threeD: Ref<boolean>

beforeEach(async () => {
  vi.useFakeTimers()
  TrackedMockWorker.instances.length = 0
  ctorSpy.mockClear()

  rawData = ref([dp('x1', 'y1'), dp('x2', 'y1'), dp('x1', 'y2')])
  arrangement = ref({ identityString: 'xy', targetString: 'xy' })
  activeGroupId = ref(0)
  sort = ref(noSort)
  showLabels = ref(false)
  scale = ref('linear' as ScaleType)
  threeD = ref(false)

  scope = effectScope()
  result = scope.run(() =>
    useChartPipeline(
      rawData,
      arrangement,
      ref(defaultLabels),
      activeGroupId,
      sort,
      showLabels,
      scale,
      threeD
    )
  )!
  // Flush the immediate watch + 50 ms debounce.
  await vi.advanceTimersByTimeAsync(50)
  worker = TrackedMockWorker.instances[0]!
  expect(worker).toBeDefined()
})

afterEach(() => {
  scope.stop()
  vi.useRealTimers()
})

function replyReady(dataEpoch: number) {
  const r: ReadyMessage = {
    type: 'ready',
    dataEpoch,
    signatures: [
      { signature: 'sig-val', title: 'val' },
      { signature: 'sig-other', title: 'other' },
    ],
    groupNames: [''],
  }
  worker.__emit(r)
}

function replyChart(dataEpoch: number, jobEpoch: number, signature: string) {
  const chart: ChartData = {
    points: [],
    yAxis: [],
    zAxis: [],
    series: [],
  } as unknown as ChartData
  const c: ChartMessage = { type: 'chart', dataEpoch, jobEpoch, signature, chart }
  worker.__emit(c)
}

describe('useChartPipeline — init', () => {
  it('posts init on first run with non-empty data', () => {
    const initCall = worker.postMessage.mock.calls.find((c) => c[0].type === 'init')
    expect(initCall).toBeDefined()
    expect(initCall![0].dataEpoch).toBe(1)
    expect(initCall![0].identityString).toBe('xy')
  })

  it('populates charts with skeleton slots after ready', () => {
    replyReady(1)
    expect(result.charts.value.length).toBe(2)
    expect(result.charts.value.every((c) => c.pending)).toBe(true)
  })
})

describe('useChartPipeline — compute drain', () => {
  it('posts the first compute on ready (serial drain)', () => {
    replyReady(1)
    // After ready, exactly one compute is in flight; subsequent computes wait for replies.
    // Filter for `compute` (not `init`) so the pre-ready init call doesn't confuse the count.
    const computeCalls = worker.postMessage.mock.calls.filter((c) => c[0].type === 'compute')
    expect(computeCalls.length).toBe(1)
    expect(computeCalls[0]![0].signature).toBe('sig-val')
  })

  it('unblocks the next compute when a chart reply lands', () => {
    replyReady(1)
    worker.postMessage.mockClear()
    replyChart(1, 1, 'sig-val')
    expect(worker.postMessage.mock.calls.length).toBe(1)
    expect(worker.postMessage.mock.calls[0]![0].signature).toBe('sig-other')
  })

  it('drops stale ready replies (mismatched dataEpoch)', () => {
    replyReady(1)
    worker.postMessage.mockClear()
    worker.__emit({ type: 'ready', dataEpoch: 0, signatures: [], groupNames: [] } as WorkerResponse)
    expect(worker.postMessage.mock.calls.length).toBe(0)
  })

  it('drops stale chart replies (mismatched jobEpoch) but still drains', () => {
    replyReady(1)
    worker.postMessage.mockClear()
    replyChart(1, 999, 'sig-val')
    expect(result.charts.value.find((c) => c.key === 'sig-val')!.data).toBeNull()
    expect(worker.postMessage.mock.calls.length).toBe(1)
  })
})

describe('useChartPipeline — setArrangement', () => {
  it('posts setArrangement on swap and does not re-clone data', async () => {
    replyReady(1)
    worker.postMessage.mockClear()
    arrangement.value = { identityString: 'yx', targetString: 'yx' }
    await vi.advanceTimersByTimeAsync(50)
    const setCall = worker.postMessage.mock.calls.find((c) => c[0].type === 'setArrangement')
    expect(setCall).toBeDefined()
    const initCall = worker.postMessage.mock.calls.find((c) => c[0].type === 'init')
    expect(initCall).toBeUndefined()
  })
})

describe('useChartPipeline — param changes', () => {
  it('posts compute (not init) when sort changes', async () => {
    replyReady(1)
    // Reply to BOTH in-flight computes to fully free the drain. The chart
    // handler pumps the next compute on every reply, so a single reply still
    // leaves draining=true; recompute re-uses that state and pumpQueue returns
    // early. Two replies exhaust the queue and leave draining=false.
    replyChart(1, 1, 'sig-val')
    replyChart(1, 1, 'sig-other')
    worker.postMessage.mockClear()
    sort.value = { enabled: true, order: 'desc' }
    await vi.advanceTimersByTimeAsync(50)
    const initCall = worker.postMessage.mock.calls.find((c) => c[0].type === 'init')
    expect(initCall).toBeUndefined()
    expect(worker.postMessage.mock.calls.some((c) => c[0].type === 'compute')).toBe(true)
  })
})

describe('useChartPipeline — data change', () => {
  it('bumps dataEpoch and posts init on a new dataset', async () => {
    replyReady(1)
    worker.postMessage.mockClear()
    rawData.value = [dp('new1', 'y1'), dp('new2', 'y1')]
    await vi.advanceTimersByTimeAsync(50)
    const initCall = worker.postMessage.mock.calls.find((c) => c[0].type === 'init')
    expect(initCall).toBeDefined()
    expect(initCall![0].dataEpoch).toBe(2)
  })
})

describe('useChartPipeline — dispose', () => {
  it('terminates the worker on scope dispose', () => {
    replyReady(1)
    scope.stop()
    expect(worker.terminate).toHaveBeenCalled()
  })
})

describe('useChartPipeline — empty data', () => {
  it('does not post init and clears charts when data is empty', async () => {
    worker.postMessage.mockClear()
    rawData.value = []
    await vi.advanceTimersByTimeAsync(50)
    const initCall = worker.postMessage.mock.calls.find((c) => c[0].type === 'init')
    expect(initCall).toBeUndefined()
    expect(result.charts.value).toEqual([])
    expect(result.hasAny.value).toBe(false)
  })
})
