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
import type { DataPoint, ScaleType, ChartData, Axis } from '../types'
import { dp, noSort, VALUE_AXES } from '@/test-utils'

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

  it('applies non-null labels on grouped setArrangement', () => {
    send(buildInit())
    postSpy.mockClear()
    send({
      type: 'setArrangement',
      identityString: 'xy',
      targetString: 'xy',
      labels: { x: 'region', y: 'cat' },
    })
    expect(ready()).toBeDefined()
  })

  it('compute uses the named group when present', () => {
    send(buildInit())
    const sig = ready()!.signatures[0]!.signature
    const group = ready()!.groupNames[0]!
    postSpy.mockClear()
    send(buildCompute({ signature: sig, groupName: group }))
    expect(charts()).toHaveLength(1)
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

function valueDp(xAxis: string, yAxis: string): DataPoint {
  return { xAxis, yAxis, stats: [] }
}

describe('transform.worker — value mode init', () => {
  it('replies with one synthetic signature when scatter + value-mode axes', () => {
    send(
      buildInit({
        data: [valueDp('100', '12'), valueDp('200', '8')],
        axes: VALUE_AXES,
        chartType: 'scatter',
      })
    )

    const r = ready()
    expect(r).toBeDefined()
    expect(r!.signatures).toHaveLength(1)
    expect(r!.signatures[0]!.signature).toBe('__value_mode__')
    expect(r!.groupNames).toEqual([])
  })

  it('uses normal stat signatures when value axes but chart is not in eligible set', () => {
    send(
      buildInit({
        data: [dp('x1', 'y1', '', 'val', 10)],
        axes: VALUE_AXES,
        chartType: 'pie',
      })
    )

    const r = ready()
    expect(r!.signatures[0]!.signature).not.toBe('__value_mode__')
    expect(r!.signatures.length).toBeGreaterThan(0)
  })
})

describe('transform.worker — value mode compute', () => {
  it('returns a ChartData with valueTuples for __value_mode__ signature', () => {
    send(
      buildInit({
        data: [valueDp('100', '12'), valueDp('200', '8')],
        axes: VALUE_AXES,
        chartType: 'scatter',
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
        chartType: 'scatter',
      })
    )
    postSpy.mockClear()

    send(buildCompute({ signature: '__value_mode__', groupName: '' }))

    const chart = charts()[0]!.chart as ChartData
    expect(chart.valueTuples).toHaveLength(1)
    expect(chart.valueTuples![0]).toEqual([1, 2])
  })

  it('setArrangement xy→yx flips valueTuples on 2-col value mode', () => {
    send(
      buildInit({
        data: [valueDp('100', '12')],
        axes: VALUE_AXES,
        chartType: 'scatter',
      })
    )
    postSpy.mockClear()

    send({ type: 'setArrangement', identityString: 'xy', targetString: 'yx' })
    ready()

    send(buildCompute({ signature: '__value_mode__', groupName: '' }))
    const chart = charts()[0]!.chart as ChartData
    expect(chart.valueTuples).toEqual([[12, 100]])
  })

  it('3-col value mode compute with xyz swap yields continuous render3D', () => {
    const VALUE_AXES_3 = [
      { key: 'x' as const, label: 'x', type: 'value' as const },
      { key: 'y' as const, label: 'y', type: 'value' as const },
      { key: 'z' as const, label: 'z', type: 'value' as const },
    ]
    send(
      buildInit({
        data: [{ xAxis: '1', yAxis: '2', zAxis: '3', stats: [] }],
        axes: VALUE_AXES_3,
        identityString: 'xyz',
        targetString: 'xyz',
        chartType: 'scatter',
      })
    )
    postSpy.mockClear()
    send(buildCompute({ signature: '__value_mode__', groupName: '', threeD: true }))
    const chart = charts()[0]!.chart as ChartData
    expect(chart.render3D?.mode).toBe('continuous')
    expect(chart.valuePoints3D).toEqual([[1, 2, 3]])
  })

  it('replies with __mixed_mode__ signature for bar + mixed axes', () => {
    const MIXED_AXES: Axis[] = [
      { key: 'x', label: 'region' },
      { key: 'y', label: 'latency', type: 'value' },
    ]
    send(
      buildInit({
        data: [
          { xAxis: 'Asia', yAxis: '12', stats: [] },
          { xAxis: 'EU', yAxis: '11', stats: [] },
        ],
        axes: MIXED_AXES,
        chartType: 'bar',
      })
    )

    const r = ready()
    expect(r!.signatures).toHaveLength(1)
    expect(r!.signatures[0]!.signature).toBe('__mixed_mode__')
    expect(r!.groupNames).toEqual([])
  })

  it('replies with __mixed_mode__ signature for scatter + mixed axes', () => {
    const MIXED_AXES: Axis[] = [
      { key: 'x', label: 'region' },
      { key: 'y', label: 'latency', type: 'value' },
    ]
    send(
      buildInit({
        data: [
          { xAxis: 'Asia', yAxis: '12', stats: [] },
          { xAxis: 'EU', yAxis: '11', stats: [] },
        ],
        axes: MIXED_AXES,
        chartType: 'scatter',
      })
    )

    const r = ready()
    expect(r!.signatures).toHaveLength(1)
    expect(r!.signatures[0]!.signature).toBe('__mixed_mode__')
    expect(r!.groupNames).toEqual([])
  })

  it('returns mixedTuples for __mixed_mode__ compute', () => {
    const MIXED_AXES: Axis[] = [
      { key: 'x', label: 'region' },
      { key: 'y', label: 'latency', type: 'value' },
    ]
    send(
      buildInit({
        data: [
          { xAxis: 'Asia', yAxis: '12', stats: [] },
          { xAxis: 'EU', yAxis: '11', stats: [] },
        ],
        axes: MIXED_AXES,
        chartType: 'scatter',
      })
    )
    postSpy.mockClear()

    send(buildCompute({ signature: '__mixed_mode__', groupName: '' }))

    const chart = charts()[0]!.chart as ChartData
    expect(chart.statType).toBe('mixed')
    expect(chart.xCategories).toEqual(['Asia', 'EU'])
    expect(chart.mixedTuples).toEqual([
      [0, 12],
      [1, 11],
    ])
  })

  it('preserveRows expands collapsed stats[] on one DataPoint', () => {
    send(
      buildInit({
        data: [
          {
            xAxis: 'West',
            yAxis: '',
            stats: [
              { type: 'tax', value: 10 },
              { type: 'amount', value: 100 },
              { type: 'tax', value: 20 },
              { type: 'amount', value: 200 },
            ],
          },
        ],
        preserveRows: true,
      })
    )
    const taxSig = ready()!.signatures.find((s) => s.title === 'tax')!.signature
    postSpy.mockClear()

    send(buildCompute({ signature: taxSig, groupName: '' }))

    const chart = charts()[0]!.chart as ChartData
    expect(chart.mixedTuples).toEqual([
      [0, 10],
      [0, 20],
    ])
    expect(chart.xCategories).toEqual(['West'])
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

describe('transform.worker — labels on setArrangement and mixed 3D', () => {
  it('grouped setArrangement accepts labels null and unknown group falls back', () => {
    send(buildInit({ labels: { x: 'X' } }))
    postSpy.mockClear()
    send({
      type: 'setArrangement',
      identityString: 'xy',
      targetString: 'xy',
      labels: null as unknown as undefined,
    })
    expect(ready()).toBeDefined()

    // compute with empty grouped map path: re-init with empty data
    send(buildInit({ data: [] }))
    const sig = ready()?.signatures[0]?.signature
    postSpy.mockClear()
    if (sig) {
      send(buildCompute({ signature: sig, groupName: 'ghost' }))
      // no signatures from empty data → may be undefined
    }
  })

  it('value mode setArrangement updates labels', () => {
    send(
      buildInit({
        data: [valueDp('1', '2')],
        axes: VALUE_AXES,
        chartType: 'scatter',
        labels: { x: 'old' },
      })
    )
    postSpy.mockClear()
    send({
      type: 'setArrangement',
      identityString: 'xy',
      targetString: 'xy',
      labels: { x: 'price', y: 'lat' },
    })
    const r = ready()
    expect(r!.signatures[0]!.title).toContain('price')
  })

  it('value mode compute ignores wrong signature', () => {
    send(
      buildInit({
        data: [valueDp('1', '2')],
        axes: VALUE_AXES,
        chartType: 'scatter',
      })
    )
    postSpy.mockClear()
    send(buildCompute({ signature: 'nope' }))
    expect(charts()).toHaveLength(0)
  })

  it('mixed mode setArrangement and compute with 3D axes', () => {
    const MIXED3: Axis[] = [
      { key: 'x', label: 'region' },
      { key: 'y', label: 'lat', type: 'value' },
      { key: 'z', label: 'z', type: 'value' },
    ]
    send(
      buildInit({
        data: [{ xAxis: 'Asia', yAxis: '12', zAxis: '3', stats: [] }],
        axes: MIXED3,
        chartType: 'scatter',
      })
    )
    expect(ready()!.signatures[0]!.signature).toBe('__mixed_mode__')
    postSpy.mockClear()
    send({
      type: 'setArrangement',
      identityString: 'xyz',
      targetString: 'xyz',
      labels: { x: 'region' },
    })
    expect(ready()).toBeDefined()
    postSpy.mockClear()
    send(buildCompute({ signature: '__mixed_mode__' }))
    const chart = charts()[0]!.chart as ChartData
    expect(chart.render3D?.mode).toBe('mixed')

    postSpy.mockClear()
    send(buildCompute({ signature: 'wrong' }))
    expect(charts()).toHaveLength(0)
  })
})

describe('transform.worker remaining branch edges', () => {
  it('value setArrangement with and without labels', () => {
    send(
      buildInit({
        data: [valueDp('1', '2')],
        axes: VALUE_AXES,
        chartType: 'scatter',
        labels: { x: 'a' },
      })
    )
    postSpy.mockClear()
    send({
      type: 'setArrangement',
      identityString: 'xy',
      targetString: 'yx',
      labels: undefined,
    })
    expect(ready()!.signatures[0]!.title).toBeTruthy()

    postSpy.mockClear()
    send({ type: 'setArrangement', identityString: 'xy', targetString: 'xy' })
    expect(ready()).toBeDefined()

    postSpy.mockClear()
    send({
      type: 'setArrangement',
      identityString: 'xy',
      targetString: 'xy',
      labels: null as unknown as undefined,
    })
    expect(ready()).toBeDefined()
  })

  it('mixed setArrangement with null labels and without labels key', () => {
    const MIXED: Axis[] = [
      { key: 'x', label: 'region' },
      { key: 'y', label: 'lat', type: 'value' },
    ]
    send(
      buildInit({
        data: [{ xAxis: 'A', yAxis: '1', stats: [] }],
        axes: MIXED,
        chartType: 'scatter',
        labels: { x: 'keep' },
      })
    )
    expect(ready()!.signatures[0]!.title).toContain('region')
    postSpy.mockClear()
    send({
      type: 'setArrangement',
      identityString: 'xy',
      targetString: 'xy',
      labels: null as unknown as undefined,
    })
    expect(ready()).toBeDefined()

    postSpy.mockClear()
    send({ type: 'setArrangement', identityString: 'xy', targetString: 'xy' })
    expect(ready()).toBeDefined()
  })

  it('value mode compute succeeds for __value_mode__', () => {
    send(
      buildInit({
        data: [valueDp('1', '2')],
        axes: VALUE_AXES,
        chartType: 'scatter',
      })
    )
    postSpy.mockClear()
    send(buildCompute({ signature: '__value_mode__' }))
    expect(charts()).toHaveLength(1)
  })

  it('empty init has no signatures to compute', () => {
    send(
      buildInit({
        data: [],
        identityString: 'xy',
        targetString: 'xy',
      })
    )
    postSpy.mockClear()
    send(buildCompute({ signature: 'val-', groupName: 'x' }))
    expect(charts()).toHaveLength(0)
  })

  it('value and mixed ready titles when axes present', () => {
    send(
      buildInit({
        data: [valueDp('1', '2')],
        axes: VALUE_AXES,
        chartType: 'scatter',
      })
    )
    expect(ready()!.signatures[0]!.title).toBe('price vs latency')

    postSpy.mockClear()
    const MIXED: Axis[] = [
      { key: 'x', label: 'region' },
      { key: 'y', label: 'lat', type: 'value' },
    ]
    send(
      buildInit({
        data: [{ xAxis: 'A', yAxis: '1', stats: [] }],
        axes: MIXED,
        chartType: 'bar',
      })
    )
    expect(ready()!.signatures[0]!.title).toBe('region vs lat')
  })
})
