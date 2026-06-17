// Helpers for testing the worker files in this directory. The bridge composables
// (`useStatsWorker`, `useChartPipeline`) declare their own `TrackedMockWorker`
// classes inside their test files because they need a `static instances` array
// to capture the singleton — the simpler shape here is enough for the worker
// file tests, which only need a place for `self.onmessage` to land.
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
