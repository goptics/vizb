// ui/src/workers/stats.worker.test.ts
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { installMockSelf, uninstallMockSelf } from './__test-utils__/workerHarness'
import type { StatsRequest, StatsResponse } from './stats.worker'
import type { Point3D } from '../types'

// Tiny deterministic point cloud so the lib/stats output is predictable.
// Real Point3D shape: { xAxis, yAxis, zAxis, value } — one numeric value per point.
const points: Point3D[] = [
  { xAxis: 'a', yAxis: 'y1', zAxis: 'z1', value: 1 },
  { xAxis: 'a', yAxis: 'y2', zAxis: 'z1', value: 2 },
  { xAxis: 'b', yAxis: 'y1', zAxis: 'z1', value: 3 },
  { xAxis: 'b', yAxis: 'y2', zAxis: 'z1', value: 4 },
]
const yAxis = ['y1', 'y2']
const zAxis = ['z1']
const seriesOrder = ['a', 'b']

async function loadWorker() {
  const harness = installMockSelf()
  await import('./stats.worker.ts')
  const handler = harness.getHandler()!
  return { postSpy: harness.postSpy, handler }
}

describe('stats.worker', () => {
  let postSpy: ReturnType<typeof vi.fn>
  let handler: (e: MessageEvent<StatsRequest>) => void

  beforeEach(async () => {
    ;({ postSpy, handler } = await loadWorker())
  })

  afterEach(() => {
    uninstallMockSelf()
    vi.resetModules()
  })

  it("kind: 'descriptive' returns seriesProfiles only", () => {
    const req: StatsRequest = {
      type: 'compute',
      id: 1,
      kind: 'descriptive',
      points,
      yAxis,
      zAxis,
      seriesOrder,
    }
    handler({ data: req } as MessageEvent<StatsRequest>)

    expect(postSpy).toHaveBeenCalledTimes(1)
    const reply = postSpy.mock.calls[0]![0] as StatsResponse
    expect(reply.type).toBe('result')
    expect(reply.id).toBe(1)
    expect(reply.seriesProfiles).toBeDefined()
    expect(reply.seriesProfiles!.length).toBe(2)
    expect(reply.seriesProfiles!.some((p) => p.stats.count > 0)).toBe(true)
    expect(reply.correlation).toBeUndefined()
  })

  it("kind: 'correlation' returns correlation only", () => {
    const req: StatsRequest = {
      type: 'compute',
      id: 2,
      kind: 'correlation',
      points,
      yAxis,
      zAxis,
      seriesOrder,
      axis: 'x',
    }
    handler({ data: req } as MessageEvent<StatsRequest>)

    expect(postSpy).toHaveBeenCalledTimes(1)
    const reply = postSpy.mock.calls[0]![0] as StatsResponse
    expect(reply.id).toBe(2)
    expect(reply.correlation).toBeDefined()
    expect(reply.seriesProfiles).toBeUndefined()
  })

  it('replies with the same id as the request', () => {
    handler({
      data: { type: 'compute', id: 42, kind: 'descriptive', points, yAxis, zAxis, seriesOrder },
    } as MessageEvent<StatsRequest>)
    expect((postSpy.mock.calls[0]![0] as StatsResponse).id).toBe(42)
  })

  it('handles two back-to-back requests with different ids', () => {
    handler({
      data: { type: 'compute', id: 1, kind: 'descriptive', points, yAxis, zAxis, seriesOrder },
    } as MessageEvent<StatsRequest>)
    handler({
      data: { type: 'compute', id: 2, kind: 'descriptive', points, yAxis, zAxis, seriesOrder },
    } as MessageEvent<StatsRequest>)

    expect(postSpy).toHaveBeenCalledTimes(2)
    expect((postSpy.mock.calls[0]![0] as StatsResponse).id).toBe(1)
    expect((postSpy.mock.calls[1]![0] as StatsResponse).id).toBe(2)
  })
})
