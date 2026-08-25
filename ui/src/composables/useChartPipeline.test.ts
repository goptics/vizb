// ui/src/composables/useChartPipeline.test.ts
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { effectScope, reactive, ref, type Ref } from 'vue'
import type { WorkerResponse, ReadyMessage, ChartMessage } from '../workers/transform.worker'
import type { DataPoint, AxisLabels, Sort, ScaleType, ScaleInput, ChartData, Axis } from '../types'
import { dp, noSort, VALUE_AXES, MIXED_AXES } from '@/test-utils'
import { TrackedMockWorker } from '../workers/__test-utils__/workerHarness'

vi.mock('../workers/transform.worker.ts?worker&inline', () => ({
  default: TrackedMockWorker,
}))

const { useChartPipeline } = await import('./useChartPipeline')

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
  TrackedMockWorker.reset()

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
  worker = TrackedMockWorker.latest()
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

describe('useChartPipeline — object scale clone', () => {
  it('posts a plain cloneable object scale from a Vue reactive Proxy', async () => {
    scope.stop()
    TrackedMockWorker.reset()
    const objectScale = reactive<ScaleInput>({ type: 'log', axes: ['x'] })
    const objectScaleRef = ref<ScaleInput>(objectScale)
    // Vue reactive Proxies cannot be structured-cloned (the original bug).
    expect(() => structuredClone(objectScaleRef.value)).toThrow()

    scope = effectScope()
    result = scope.run(() =>
      useChartPipeline(
        rawData,
        arrangement,
        ref(defaultLabels),
        activeGroupId,
        sort,
        showLabels,
        objectScaleRef,
        threeD
      )
    )!
    await vi.advanceTimersByTimeAsync(50)
    worker = TrackedMockWorker.latest()
    replyReady(1)

    const computeCall = worker.postMessage.mock.calls.find((c) => c[0].type === 'compute')
    expect(computeCall).toBeDefined()
    const posted = computeCall![0] as { scale: ScaleInput }
    expect(posted.scale).toEqual({ type: 'log', axes: ['x'] })
    expect(Object.getPrototypeOf(posted.scale)).toBe(Object.prototype)
    expect(Array.isArray((posted.scale as { axes?: unknown }).axes)).toBe(true)
    expect(posted.scale).not.toBe(objectScale)
    expect(() => structuredClone(posted)).not.toThrow()
  })

  it('copies optional log bases and still posts string scale as a string', async () => {
    scope.stop()
    TrackedMockWorker.reset()
    const objectScale = reactive<ScaleInput>({
      type: 'log',
      axes: ['x', 'y'],
      base: 2,
      baseX: 4,
      baseY: 8,
      baseZ: 16,
    })
    scope = effectScope()
    result = scope.run(() =>
      useChartPipeline(
        rawData,
        arrangement,
        ref(defaultLabels),
        activeGroupId,
        sort,
        showLabels,
        ref(objectScale),
        threeD
      )
    )!
    await vi.advanceTimersByTimeAsync(50)
    worker = TrackedMockWorker.latest()
    replyReady(1)

    const computeCall = worker.postMessage.mock.calls.find((c) => c[0].type === 'compute')
    const posted = computeCall![0] as { scale: ScaleInput }
    expect(posted.scale).toEqual({
      type: 'log',
      axes: ['x', 'y'],
      base: 2,
      baseX: 4,
      baseY: 8,
      baseZ: 16,
    })
    expect(() => structuredClone(posted.scale)).not.toThrow()

    scope.stop()
    TrackedMockWorker.reset()
    scope = effectScope()
    result = scope.run(() =>
      useChartPipeline(
        rawData,
        arrangement,
        ref(defaultLabels),
        activeGroupId,
        sort,
        showLabels,
        ref('log' as ScaleType),
        threeD
      )
    )!
    await vi.advanceTimersByTimeAsync(50)
    worker = TrackedMockWorker.latest()
    replyReady(1)
    const stringCompute = worker.postMessage.mock.calls.find((c) => c[0].type === 'compute')
    expect(stringCompute![0].scale).toBe('log')
    expect(() => structuredClone(stringCompute![0])).not.toThrow()
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

describe('useChartPipeline — value mode setArrangement', () => {
  it('posts setArrangement when arrangement changes with value-mode axes', async () => {
    scope.stop()
    TrackedMockWorker.reset()
    scope = effectScope()
    scope.run(() =>
      useChartPipeline(
        rawData,
        arrangement,
        ref(defaultLabels),
        activeGroupId,
        sort,
        showLabels,
        scale,
        threeD,
        ref(VALUE_AXES),
        ref('scatter')
      )
    )
    await vi.advanceTimersByTimeAsync(50)
    const w = TrackedMockWorker.latest()
    w.postMessage.mock.calls.find((c) => c[0].type === 'init')
    w.__emit({
      type: 'ready',
      dataEpoch: 1,
      signatures: [{ signature: '__value_mode__', statTemplate: { type: 'value' } }],
      groupNames: [],
    })
    w.postMessage.mockClear()

    arrangement.value = { identityString: 'xy', targetString: 'yx' }
    await vi.advanceTimersByTimeAsync(50)

    const setCall = w.postMessage.mock.calls.find((c) => c[0].type === 'setArrangement')
    expect(setCall).toBeDefined()
    expect(setCall![0]).toMatchObject({ targetString: 'yx', identityString: 'xy' })
  })
})

describe('useChartPipeline — value mode axes forwarding', () => {
  it('includes axes in the init postMessage when provided', async () => {
    // Re-mount pipeline with axes ref
    scope.stop()
    TrackedMockWorker.reset()
    scope = effectScope()
    scope.run(() =>
      useChartPipeline(
        rawData,
        arrangement,
        ref(defaultLabels),
        activeGroupId,
        sort,
        showLabels,
        scale,
        threeD,
        ref(VALUE_AXES) // new axes param
      )
    )
    await vi.advanceTimersByTimeAsync(50)
    const w = TrackedMockWorker.latest()

    const initCall = w.postMessage.mock.calls.find(
      (c: unknown[]) => (c[0] as { type: string }).type === 'init'
    )
    expect(initCall).toBeDefined()
    expect((initCall![0] as { axes: Axis[] }).axes).toEqual(VALUE_AXES)
  })

  it('posts plain (non-proxy) axes so structured clone succeeds', async () => {
    const axes = reactive<Axis[]>([
      { key: 'x', label: 'price', type: 'value' },
      { key: 'y', label: 'latency', type: 'value' },
    ])

    scope.stop()
    TrackedMockWorker.reset()
    scope = effectScope()
    scope.run(() =>
      useChartPipeline(
        rawData,
        arrangement,
        ref(defaultLabels),
        activeGroupId,
        sort,
        showLabels,
        scale,
        threeD,
        ref(axes)
      )
    )
    await vi.advanceTimersByTimeAsync(50)
    const w = TrackedMockWorker.latest()
    const initCall = w.postMessage.mock.calls.find(
      (c: unknown[]) => (c[0] as { type: string }).type === 'init'
    )
    const posted = (initCall![0] as { axes: Axis[] }).axes
    expect(posted).toEqual([
      { key: 'x', label: 'price', type: 'value' },
      { key: 'y', label: 'latency', type: 'value' },
    ])
    // structuredClone throws on Vue reactive proxies — must be plain objects.
    expect(() => structuredClone(posted)).not.toThrow()
  })
})

describe('useChartPipeline — mixed mode skeleton', () => {
  it('pre-populates mixed-mode skeleton with 2D title when z is categorical', async () => {
    scope.stop()
    TrackedMockWorker.reset()
    scope = effectScope()
    result = scope.run(() =>
      useChartPipeline(
        rawData,
        arrangement,
        ref({ name: 'N', x: 'X', y: 'Y', z: 'Z', metric: 'M' }),
        activeGroupId,
        sort,
        showLabels,
        scale,
        threeD,
        ref(MIXED_AXES),
        ref('scatter'),
        ref(true),
        ref('Dataset Title')
      )
    )!
    await vi.advanceTimersByTimeAsync(50)
    worker = TrackedMockWorker.latest()

    expect(result.charts.value).toEqual([
      {
        key: '__mixed_mode__',
        title: 'region vs latency',
        data: null,
        pending: true,
      },
    ])

    const initCall = worker.postMessage.mock.calls.find((c) => c[0].type === 'init')
    expect(initCall?.[0]).toMatchObject({
      labels: { name: 'N', x: 'X', y: 'Y', z: 'Z', metric: 'M' },
      preserveRows: true,
      chartType: 'scatter',
    })
  })

  it('uses 3D mixed title when z axis is value-typed', async () => {
    scope.stop()
    TrackedMockWorker.reset()
    const mixed3D = [
      { key: 'x' as const, label: 'region' },
      { key: 'y' as const, label: 'latency', type: 'value' as const },
      { key: 'z' as const, label: 'depth', type: 'value' as const },
    ]
    scope = effectScope()
    result = scope.run(() =>
      useChartPipeline(
        rawData,
        arrangement,
        undefined,
        activeGroupId,
        sort,
        showLabels,
        scale,
        threeD,
        ref(mixed3D),
        ref('bar')
      )
    )!
    await vi.advanceTimersByTimeAsync(50)

    expect(result.charts.value[0]).toMatchObject({
      key: '__mixed_mode__',
      title: 'region · latency · depth',
      pending: true,
    })
  })

  it('falls back to axis key labels when mixed axes omit labels', async () => {
    scope.stop()
    TrackedMockWorker.reset()
    scope = effectScope()
    result = scope.run(() =>
      useChartPipeline(
        rawData,
        arrangement,
        ref({}),
        activeGroupId,
        sort,
        showLabels,
        scale,
        threeD,
        ref([
          { key: 'x', type: 'category' as never },
          { key: 'y', type: 'value' },
        ]),
        ref('line')
      )
    )!
    await vi.advanceTimersByTimeAsync(50)

    expect(result.charts.value[0]?.title).toBe('x vs y')
  })
})

describe('useChartPipeline — chart title fallback and plain array input', () => {
  it('fills blank chart titles from datasetName and accepts plain array rawData', async () => {
    scope.stop()
    TrackedMockWorker.reset()
    const plainRows = [dp('x1', 'y1'), dp('x2', 'y1')]
    scope = effectScope()
    result = scope.run(() =>
      useChartPipeline(
        plainRows,
        arrangement,
        ref(undefined),
        activeGroupId,
        sort,
        showLabels,
        scale,
        threeD,
        undefined,
        undefined,
        undefined,
        ref('Fallback Name')
      )
    )!
    await vi.advanceTimersByTimeAsync(50)
    worker = TrackedMockWorker.latest()

    replyReady(1)
    worker.postMessage.mockClear()

    const chart: ChartData = {
      points: [],
      yAxis: [],
      zAxis: [],
      series: [],
      title: '',
    } as unknown as ChartData
    worker.__emit({
      type: 'chart',
      dataEpoch: 1,
      jobEpoch: 1,
      signature: 'sig-val',
      chart,
    } as ChartMessage)

    expect(result.charts.value.find((c) => c.key === 'sig-val')?.data?.title).toBe('Fallback Name')
    expect(result.hasAny.value).toBe(true)
  })

  it('skips setArrangement and recompute while ready is in flight or data is pending', async () => {
    replyReady(1)
    replyChart(1, 1, 'sig-val')
    replyChart(1, 1, 'sig-other')

    // Trigger data pending — params recompute should no-op until reinit.
    worker.postMessage.mockClear()
    rawData.value = [dp('pending', 'y1')]
    sort.value = { enabled: true, order: 'asc' }
    await vi.advanceTimersByTimeAsync(10)
    expect(worker.postMessage.mock.calls.some((c) => c[0].type === 'compute')).toBe(false)

    await vi.advanceTimersByTimeAsync(50)
    const initCall = worker.postMessage.mock.calls.find((c) => c[0].type === 'init')
    expect(initCall).toBeDefined()

    // While waiting for ready, arrangement should also no-op.
    worker.postMessage.mockClear()
    arrangement.value = { identityString: 'xy', targetString: 'yx' }
    await vi.advanceTimersByTimeAsync(50)
    expect(
      worker.postMessage.mock.calls.find((c) => c[0].type === 'setArrangement')
    ).toBeUndefined()
  })
})

describe('useChartPipeline — drain guard and value label fallbacks', () => {
  it('ignores pump while already draining and drops unknown chart slots', async () => {
    replyReady(1)
    // A second ready while a compute is draining should still process, but
    // pumpQueue's draining early-return is hit by recompute while a compute is open.
    worker.postMessage.mockClear()
    sort.value = { enabled: true, order: 'asc' }
    await vi.advanceTimersByTimeAsync(50)
    // draining is true from the ready-started compute; recompute queued but pump returns early.
    // Emit a chart for an unknown signature — slot branch is skipped.
    replyChart(1, 2, 'missing-sig')
    expect(result.charts.value.every((c) => c.key !== 'missing-sig')).toBe(true)
  })

  it('uses default x/y labels in value mode when axes omit labels', async () => {
    scope.stop()
    TrackedMockWorker.reset()
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
        threeD,
        ref([
          { key: 'x', type: 'value' },
          { key: 'y', type: 'value' },
        ]),
        ref('scatter')
      )
    )!
    await vi.advanceTimersByTimeAsync(50)
    expect(result.charts.value[0]).toMatchObject({
      key: '__value_mode__',
      title: 'x vs y',
    })
  })

  it('keeps existing chart title when worker already filled it', async () => {
    replyReady(1)
    const chart: ChartData = {
      points: [],
      yAxis: [],
      zAxis: [],
      series: [],
      title: 'Worker Title',
    } as unknown as ChartData
    worker.__emit({
      type: 'chart',
      dataEpoch: 1,
      jobEpoch: 1,
      signature: 'sig-val',
      chart,
    } as ChartMessage)
    expect(result.charts.value.find((c) => c.key === 'sig-val')?.data?.title).toBe('Worker Title')
  })
})
