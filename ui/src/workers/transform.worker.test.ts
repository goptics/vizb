// ui/src/workers/transform.worker.test.ts
import { describe, it, expect, beforeEach, afterEach, vi, type Mock } from 'vitest'
import { installMockSelf, uninstallMockSelf } from './__test-utils__/workerHarness'
import type {
  WorkerRequest,
  WorkerResponse,
  InitMessage,
  ComputeMessage,
  ReadyMessage,
  ChartMessage,
} from './transform.worker'
import type { DataPoint, Sort, ScaleType, ChartData, Axis } from '../types'

const noSort: Sort = { enabled: false, order: 'asc' }

function dp(xAxis: string, yAxis = '', zAxis = '', type = 'val', value = 1): DataPoint {
  return { xAxis, yAxis, zAxis, stats: [{ type, value }] }
}

function buildInit(overrides: Partial<InitMessage> = {}): InitMessage {
  return {
    type: 'init',
    dataEpoch: 1,
    data: [dp('x1', 'y1'), dp('x2', 'y1'), dp('x1', 'y2')],
    identityString: 'xy',
    targetString: 'xy',
    ...overrides,
  }
}

function buildCompute(overrides: Partial<ComputeMessage> = {}): ComputeMessage {
  return {
    type: 'compute',
    dataEpoch: 1,
    jobEpoch: 1,
    signature: '',
    groupName: '',
    sort: noSort,
    showLabels: false,
    scale: 'linear' as ScaleType,
    threeD: false,
    ...overrides,
  }
}

let postSpy: Mock
let handler: (e: MessageEvent<WorkerRequest>) => void

beforeEach(async () => {
  vi.resetModules()
  const harness = installMockSelf()
  await import('./transform.worker.ts')
  postSpy = harness.postSpy
  handler = harness.getHandler()!
})

afterEach(() => {
  uninstallMockSelf()
  vi.resetModules()
})

function send(msg: WorkerRequest) {
  handler({ data: msg } as MessageEvent<WorkerRequest>)
}

const ready = (): ReadyMessage | undefined =>
  postSpy.mock.calls.find((c) => (c[0] as WorkerResponse).type === 'ready')?.[0] as
    | ReadyMessage
    | undefined
const charts = (): ChartMessage[] =>
  postSpy.mock.calls
    .map((c) => c[0] as WorkerResponse)
    .filter((m): m is ChartMessage => m.type === 'chart')

describe('transform.worker — init', () => {
  it('replies with ready carrying dataEpoch, signatures, groupNames', () => {
    send(buildInit())

    const r = ready()
    expect(r).toBeDefined()
    expect(r!.dataEpoch).toBe(1)
    expect(r!.signatures.length).toBeGreaterThan(0)
    expect(r!.groupNames).toContain('Default')
  })
})

describe('transform.worker — setArrangement', () => {
  it('re-projects and re-replies with ready when called after init', () => {
    send(buildInit())
    postSpy.mockClear()

    send({ type: 'setArrangement', identityString: 'yx', targetString: 'yx' })

    const r = ready()
    expect(r).toBeDefined()
    expect(r!.dataEpoch).toBe(1)
    expect(r!.groupNames).toBeDefined()
  })

  it('is a no-op when called before init', () => {
    send({ type: 'setArrangement', identityString: 'yx', targetString: 'yx' })
    expect(postSpy).not.toHaveBeenCalled()
  })
})

describe('transform.worker — compute', () => {
  it('replies with a chart for a valid compute', () => {
    send(buildInit())
    const sig = ready()!.signatures[0]!.signature
    postSpy.mockClear()

    send(buildCompute({ signature: sig, groupName: '' }))

    const out = charts()
    expect(out).toHaveLength(1)
    expect(out[0]!.dataEpoch).toBe(1)
    expect(out[0]!.jobEpoch).toBe(1)
    expect(out[0]!.signature).toBe(sig)
    expect(out[0]!.chart).toBeDefined()
  })

  it('drops a compute for a superseded dataset (dataEpoch mismatch)', () => {
    send(buildInit())
    const sig = ready()!.signatures[0]!.signature
    postSpy.mockClear()

    send(buildCompute({ signature: sig, dataEpoch: 999 }))

    expect(charts()).toHaveLength(0)
  })

  it('echoes the request jobEpoch on chart replies (drop is the main-thread consumer’s job)', () => {
    send(buildInit())
    const sig = ready()!.signatures[0]!.signature
    postSpy.mockClear()

    send(buildCompute({ signature: sig, jobEpoch: 999 }))

    const out = charts()
    expect(out).toHaveLength(1)
    expect(out[0]!.jobEpoch).toBe(999)
  })

  it('drops a compute for an unknown signature', () => {
    send(buildInit())
    postSpy.mockClear()

    send(buildCompute({ signature: 'does-not-exist' }))

    expect(charts()).toHaveLength(0)
  })

  it('falls back to the first group for an unknown groupName', () => {
    send(buildInit())
    const sig = ready()!.signatures[0]!.signature
    postSpy.mockClear()

    send(buildCompute({ signature: sig, groupName: 'no-such-group' }))

    expect(charts()).toHaveLength(1)
    expect((charts()[0]!.chart as ChartData).series.length).toBeGreaterThan(0)
  })

  it('is a no-op when called before init', () => {
    send(buildCompute({ signature: 'anything' }))
    expect(postSpy).not.toHaveBeenCalled()
  })
})

const VALUE_AXES: Axis[] = [
  { key: 'x', label: 'price', type: 'value' },
  { key: 'y', label: 'latency', type: 'value' },
]

function valueDp(xAxis: string, yAxis: string): DataPoint {
  return { xAxis, yAxis, stats: [] }
}

describe('transform.worker — value mode init', () => {
  it('replies with one synthetic signature when axes are value-mode', () => {
    send(
      buildInit({
        data: [valueDp('100', '12'), valueDp('200', '8')],
        axes: VALUE_AXES,
      })
    )

    const r = ready()
    expect(r).toBeDefined()
    expect(r!.signatures).toHaveLength(1)
    expect(r!.signatures[0]!.signature).toBe('__value_mode__')
    expect(r!.groupNames).toEqual([])
  })
})

describe('transform.worker — value mode compute', () => {
  it('returns a ChartData with valueTuples for __value_mode__ signature', () => {
    send(
      buildInit({
        data: [valueDp('100', '12'), valueDp('200', '8')],
        axes: VALUE_AXES,
      })
    )
    postSpy.mockClear()

    send(buildCompute({ signature: '__value_mode__', groupName: '' }))

    const out = charts()
    expect(out).toHaveLength(1)
    const chart = out[0]!.chart as ChartData
    expect(chart.valueTuples).toEqual([
      [100, 12],
      [200, 8],
    ])
    expect(chart.series).toEqual([])
  })

  it('drops non-finite rows in value mode', () => {
    send(
      buildInit({
        data: [valueDp('1', '2'), valueDp('bad', '3')],
        axes: VALUE_AXES,
      })
    )
    postSpy.mockClear()

    send(buildCompute({ signature: '__value_mode__', groupName: '' }))

    const chart = charts()[0]!.chart as ChartData
    expect(chart.valueTuples).toHaveLength(1)
    expect(chart.valueTuples![0]).toEqual([1, 2])
  })

  it('setArrangement after value mode init preserves __value_mode__ signature', () => {
    send(
      buildInit({
        data: [valueDp('100', '12')],
        axes: VALUE_AXES,
      })
    )
    postSpy.mockClear()

    send({ type: 'setArrangement', identityString: 'xy', targetString: 'yx' })

    const r = ready()
    expect(r!.signatures).toHaveLength(1)
    expect(r!.signatures[0]!.signature).toBe('__value_mode__')
  })

  it('value mode init still allows normal category compute after re-init', () => {
    // Re-init with category data on the same worker instance
    send(buildInit())
    const r = ready()
    const sig = r!.signatures[0]!.signature
    postSpy.mockClear()

    send(buildCompute({ signature: sig, groupName: '' }))
    expect(charts()).toHaveLength(1)
    expect((charts()[0]!.chart as ChartData).valueTuples).toBeUndefined()
  })
})
