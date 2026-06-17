// ui/src/workers/__test-utils__/workerHarness.ts
import { vi, type Mock } from 'vitest'

// Minimal Worker-shaped class for use inside `vi.mock('.../X.worker.ts?worker&inline', ...)`.
// The real `?worker&inline` import resolves to a Worker constructor; this mirrors the
// three members the bridges actually touch (onmessage, postMessage, terminate) and
// adds a test-only `__emit` helper that simulates a worker reply.
export class MockWorker {
  onmessage: ((e: MessageEvent) => void) | null = null
  postMessage: Mock
  terminate: Mock
  // Test helper: simulate a message coming back from the worker.
  __emit: (data: unknown) => void

  constructor() {
    this.postMessage = vi.fn()
    this.terminate = vi.fn()
    this.__emit = (data: unknown) => {
      this.onmessage?.({ data } as MessageEvent)
    }
  }
}

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
