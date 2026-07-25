// Helpers for testing worker modules and bridge composables that construct Workers.
import { vi } from 'vitest'

// Install a fake `self` on globalThis so a worker module's top-level
// `self.onmessage = ...` assignment has somewhere to land. Returns the
// postMessage spy and a getter for the captured handler. Caller must
// `await import()` the worker module AFTER calling this.
export function installMockSelf() {
  const postSpy = vi.fn()
  const selfObj = { onmessage: null as ((e: MessageEvent) => void) | null, postMessage: postSpy }
  ;(globalThis as unknown as { self: typeof selfObj }).self = selfObj
  return {
    postSpy,
    getHandler: () => selfObj.onmessage,
  }
}

// Cleanup helper: drop the fake self so the next test starts clean.
export function uninstallMockSelf() {
  delete (globalThis as unknown as { self?: unknown }).self
}

/**
 * Captures Worker instances constructed via `?worker&inline` mocks.
 * Used by bridge composable tests (useChartPipeline, useStatsWorker).
 *
 * ```ts
 * vi.mock('../workers/foo.worker.ts?worker&inline', () => ({
 *   default: TrackedMockWorker,
 * }))
 * ```
 */
export class TrackedMockWorker {
  static instances: TrackedMockWorker[] = []
  static ctorSpy = vi.fn()

  onmessage: ((e: MessageEvent) => void) | null = null
  postMessage = vi.fn()
  terminate = vi.fn()

  __emit = (data: unknown) => this.onmessage?.({ data } as MessageEvent)

  constructor() {
    TrackedMockWorker.ctorSpy()
    TrackedMockWorker.instances.push(this)
  }

  static reset() {
    TrackedMockWorker.instances.length = 0
    TrackedMockWorker.ctorSpy.mockClear()
  }

  static latest(): TrackedMockWorker {
    const w = TrackedMockWorker.instances[TrackedMockWorker.instances.length - 1]
    if (!w) throw new Error('TrackedMockWorker.latest(): no instances')
    return w
  }
}
