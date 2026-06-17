// ui/src/composables/useStatsWorker.test.ts
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import type { ChartData, SeriesProfile, CorrelationMatrix, Point3D, SeriesData } from '../types'

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

vi.mock('../workers/stats.worker.ts?worker&inline', () => ({
  default: TrackedMockWorker,
}))

const { computeDescriptive, computeCorrelation } = await import('./useStatsWorker')

function makeChart(id: string): ChartData {
  return {
    points: [
      { xAxis: id + '-1', yAxis: 'y1', zAxis: 'z1', value: 1 },
      { xAxis: id + '-2', yAxis: 'y1', zAxis: 'z1', value: 2 },
    ] as Point3D[],
    yAxis: ['y1'],
    zAxis: ['z1'],
    series: [
      { xAxis: id + '-1', values: [1], benchmarkId: id + '-1' },
      { xAxis: id + '-2', values: [2], benchmarkId: id + '-2' },
    ] as SeriesData[],
    render3D: undefined,
  } as unknown as ChartData
}

let worker: TrackedMockWorker

beforeEach(async () => {
  // We deliberately do NOT call vi.resetModules() — the bridge has a module-level
  // worker singleton and a module-level pending map. Re-importing would replace the
  // bridge functions, but our top-level binding is to the original module's closure,
  // so re-importing would also break the binding. The singleton is reused across tests
  // (a TrackedMockWorker instance is created on the FIRST test's warmup, then reused).
  // We clean the spy history and the per-ChartData cache (by using unique chart IDs).
  ctorSpy.mockClear()
  TrackedMockWorker.instances.length = 0
  // If the singleton doesn't exist yet (first test), trigger its creation.
  if (!worker) {
    const warmupChart = makeChart('__warmup_first__')
    const warmupPromise = computeDescriptive(warmupChart)
    await Promise.resolve()
    worker = TrackedMockWorker.instances[0]!
    // Resolve the warmup so it doesn't leak.
    if (worker.postMessage.mock.calls.length > 0) {
      const id = worker.postMessage.mock.calls[0]![0].id as number
      worker.__emit({ type: 'result', id, seriesProfiles: [] as SeriesProfile[] })
    }
    await warmupPromise
  }
  worker.postMessage.mockClear()
})

afterEach(() => {
  vi.clearAllMocks()
})

function emitReply(id: number, kind: 'descriptive' | 'correlation') {
  const reply =
    kind === 'descriptive'
      ? { type: 'result' as const, id, seriesProfiles: [] as SeriesProfile[] }
      : { type: 'result' as const, id, correlation: {} as CorrelationMatrix }
  worker.__emit(reply)
}

describe('useStatsWorker — singleton lifecycle', () => {
  it('creates exactly one Worker across all tests (singleton)', () => {
    expect(ctorSpy).toHaveBeenCalledTimes(1)
  })

  it('does not create a new Worker on subsequent calls', () => {
    const ctorCountBefore = ctorSpy.mock.calls.length
    void computeDescriptive(makeChart('subsequent-call'))
    expect(ctorSpy.mock.calls.length).toBe(ctorCountBefore)
  })
})

describe('useStatsWorker — id correlation', () => {
  it('resolves the correct promise when two requests are in flight', async () => {
    const chartA = makeChart('a')
    const chartB = makeChart('b')

    const pA = computeDescriptive(chartA)
    const pB = computeDescriptive(chartB)
    // Two posts, with two distinct ids.
    expect(worker.postMessage.mock.calls.length).toBe(2)
    const idA = worker.postMessage.mock.calls[0]![0].id as number
    const idB = worker.postMessage.mock.calls[1]![0].id as number
    expect(idA).not.toBe(idB)

    // Reply out of order: B first, then A.
    emitReply(idB, 'descriptive')
    emitReply(idA, 'descriptive')

    const [rA, rB] = await Promise.all([pA, pB])
    expect(rA).toEqual([])
    expect(rB).toEqual([])
  })
})

describe('useStatsWorker — per-ChartData cache', () => {
  it('returns the same promise for two calls on the same chart', async () => {
    const chart = makeChart('cache-same')
    const p1 = computeDescriptive(chart)
    const p2 = computeDescriptive(chart)
    expect(p1).toBe(p2)
    // Only one post was made (the second call returned the cached promise).
    expect(worker.postMessage.mock.calls.length).toBe(1)
    emitReply(worker.postMessage.mock.calls[0]![0].id, 'descriptive')
    await p1
  })

  it('makes two posts for two different ChartData objects', async () => {
    const p1 = computeDescriptive(makeChart('diff-1'))
    const p2 = computeDescriptive(makeChart('diff-2'))
    expect(worker.postMessage.mock.calls.length).toBe(2)
    // Resolve both.
    for (const c of worker.postMessage.mock.calls) emitReply(c[0].id, 'descriptive')
    await Promise.all([p1, p2])
  })
})

describe('useStatsWorker — correlation axis cache', () => {
  it('uses one Worker post per axis', async () => {
    const chart = makeChart('axis')
    const pX = computeCorrelation(chart, 'x')
    const pY = computeCorrelation(chart, 'y')
    expect(worker.postMessage.mock.calls.length).toBe(2)
    const idX = worker.postMessage.mock.calls[0]![0].id as number
    const idY = worker.postMessage.mock.calls[1]![0].id as number
    emitReply(idX, 'correlation')
    emitReply(idY, 'correlation')
    await Promise.all([pX, pY])
  })

  it('caches by the "auto" key when no axis is given', async () => {
    const chart = makeChart('auto-axis')
    const p1 = computeCorrelation(chart)
    const p2 = computeCorrelation(chart)
    expect(p1).toBe(p2)
    expect(worker.postMessage.mock.calls.length).toBe(1)
    emitReply(worker.postMessage.mock.calls[0]![0].id, 'correlation')
    await p1
  })
})
