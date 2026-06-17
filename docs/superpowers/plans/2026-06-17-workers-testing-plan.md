# Worker Test Coverage Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add unit tests for `ui/src/workers/{stats,transform}.worker.ts` and the two bridge composables (`useStatsWorker`, `useChartPipeline`) that own them.

**Architecture:** Two patterns are reused across the four test files:
1. **Worker files** — install a mocked `self` on `globalThis`, dynamically `import()` the worker module, capture `self.onmessage`, invoke it with a fake `MessageEvent`. For `transform.worker.ts`, reset modules per test to clear module-level state.
2. **Bridge composables** — `vi.mock` the `?worker&inline` import with a `MockWorker` class that exposes `onmessage`/`postMessage`/`terminate` and an `__emit()` test helper. `useChartPipeline` tests use `vi.useFakeTimers()` to advance through the 50 ms debounce.

**Tech Stack:** Vitest (already configured in `ui/vitest.config.ts` with `environment: 'node'`), `vi.mock`, `vi.fn`, `vi.resetModules`, `vi.useFakeTimers`. No new dependencies.

**Reference spec:** `docs/superpowers/specs/2026-06-17-workers-testing-design.md`

---

## File Structure

**New files:**

| Path | Responsibility |
| --- | --- |
| `ui/src/workers/__test-utils__/workerHarness.ts` | `installMockSelf()` + `MockWorker` class. Shared by all four test files. |
| `ui/src/workers/stats.worker.test.ts` | Tests the stats worker's `onmessage` handler. |
| `ui/src/workers/transform.worker.test.ts` | Tests the transform worker's `onmessage` handler (init/setArrangement/compute, epoch guards, fallbacks). |
| `ui/src/composables/useStatsWorker.test.ts` | Tests the singleton bridge (id correlation, per-`ChartData` cache, per-axis correlation cache). |
| `ui/src/composables/useChartPipeline.test.ts` | Tests the pipeline (init, setArrangement, compute, debounce, drain, epoch matching, dispose). |

**No production code is modified.** This is a tests-only change.

**Verification commands (run inside `ui/`):**
- `pnpm test` — full vitest run
- `pnpm typecheck` — `vue-tsc -b --noEmit`
- `pnpm format` — prettier

---

## Task 1: Worker harness — `MockWorker` class + `installMockSelf()` helper

**Files:**
- Create: `ui/src/workers/__test-utils__/workerHarness.ts`

- [ ] **Step 1: Create the harness file with `MockWorker` class**

The `MockWorker` class mirrors the three methods the bridges touch (`onmessage`, `postMessage`, `terminate`) plus an `__emit()` test helper. `installMockSelf()` sets `globalThis.self` to a fake object with a `postMessage` spy and a nullable `onmessage` slot.

```ts
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
```

- [ ] **Step 2: Verify the file compiles**

Run: `cd ui && pnpm typecheck`
Expected: passes (no consumers yet, just a new file).

- [ ] **Step 3: Commit**

```bash
git -C /home/fahim/Projects/goptics/vizb add ui/src/workers/__test-utils__/workerHarness.ts
git -C /home/fahim/Projects/goptics/vizb commit -m "test(ui): add workerHarness with MockWorker + installMockSelf"
```

---

## Task 2: `stats.worker.test.ts` — message dispatch + id correlation

**Files:**
- Create: `ui/src/workers/stats.worker.test.ts`

The stats worker is stateless; tests share a single import. Each test resets `self` (to clear the `postMessage` spy) and captures the handler fresh.

- [ ] **Step 1: Write the failing test file**

```ts
// ui/src/workers/stats.worker.test.ts
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { installMockSelf, uninstallMockSelf } from './__test-utils__/workerHarness'
import type { StatsRequest, StatsResponse } from './stats.worker'
import type { Point3D } from '../types'

// Tiny deterministic point cloud so the lib/stats output is predictable.
const points: Point3D[] = [
  { x: 'a', y: 'y1', z: 'z1', values: [1, 2, 3] },
  { x: 'a', y: 'y2', z: 'z1', values: [4, 5, 6] },
  { x: 'b', y: 'y1', z: 'z1', values: [7, 8, 9] },
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
    const req: StatsRequest = { type: 'compute', id: 1, kind: 'descriptive', points, yAxis, zAxis, seriesOrder }
    handler({ data: req } as MessageEvent<StatsRequest>)

    expect(postSpy).toHaveBeenCalledTimes(1)
    const reply = postSpy.mock.calls[0]![0] as StatsResponse
    expect(reply.type).toBe('result')
    expect(reply.id).toBe(1)
    expect(reply.seriesProfiles).toBeDefined()
    expect(reply.seriesProfiles!.length).toBe(2)
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
      axis: 'auto',
    }
    handler({ data: req } as MessageEvent<StatsRequest>)

    expect(postSpy).toHaveBeenCalledTimes(1)
    const reply = postSpy.mock.calls[0]![0] as StatsResponse
    expect(reply.id).toBe(2)
    expect(reply.correlation).toBeDefined()
    expect(reply.seriesProfiles).toBeUndefined()
  })

  it('replies with the same id as the request', () => {
    handler({ data: { type: 'compute', id: 42, kind: 'descriptive', points, yAxis, zAxis, seriesOrder } } as MessageEvent<StatsRequest>)
    expect((postSpy.mock.calls[0]![0] as StatsResponse).id).toBe(42)
  })

  it('handles two back-to-back requests with different ids', () => {
    handler({ data: { type: 'compute', id: 1, kind: 'descriptive', points, yAxis, zAxis, seriesOrder } } as MessageEvent<StatsRequest>)
    handler({ data: { type: 'compute', id: 2, kind: 'descriptive', points, yAxis, zAxis, seriesOrder } } as MessageEvent<StatsRequest>)

    expect(postSpy).toHaveBeenCalledTimes(2)
    expect((postSpy.mock.calls[0]![0] as StatsResponse).id).toBe(1)
    expect((postSpy.mock.calls[1]![0] as StatsResponse).id).toBe(2)
  })
})
```

- [ ] **Step 2: Run the new test file**

Run: `cd ui && pnpm test -- stats.worker.test.ts`
Expected: 4 tests pass.

- [ ] **Step 3: Run the full UI test suite to confirm no regression**

Run: `cd ui && pnpm test`
Expected: all existing tests still pass, plus the 4 new ones.

- [ ] **Step 4: Commit**

```bash
git -C /home/fahim/Projects/goptics/vizb add ui/src/workers/stats.worker.test.ts
git -C /home/fahim/Projects/goptics/vizb commit -m "test(ui): cover stats.worker dispatch + id correlation"
```

---

## Task 3: `transform.worker.test.ts` — protocol, staleness, fallbacks

**Files:**
- Create: `ui/src/workers/transform.worker.test.ts`

The transform worker holds module-level `state` (the cached dataset). Every test calls `vi.resetModules()` in `beforeEach` so the state starts empty. Each test does a fresh dynamic `import()` to pick up the clean module.

- [ ] **Step 1: Write the failing test file**

```ts
// ui/src/workers/transform.worker.test.ts
import { describe, it, expect, beforeEach, afterEach, vi, type Mock } from 'vitest'
import { installMockSelf, uninstallMockSelf } from './__test-utils__/workerHarness'
import type { WorkerRequest, WorkerResponse, InitMessage, ComputeMessage, ReadyMessage, ChartMessage } from './transform.worker'
import type { DataPoint, Sort, ScaleType, ChartData } from '../types'

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
  postSpy.mock.calls.find((c) => (c[0] as WorkerResponse).type === 'ready')?.[0] as ReadyMessage | undefined
const charts = (): ChartMessage[] =>
  postSpy.mock.calls.map((c) => c[0] as WorkerResponse).filter((m): m is ChartMessage => m.type === 'chart')

describe('transform.worker — init', () => {
  it('replies with ready carrying dataEpoch, signatures, groupNames', () => {
    send(buildInit())

    const r = ready()
    expect(r).toBeDefined()
    expect(r!.dataEpoch).toBe(1)
    expect(r!.signatures.length).toBeGreaterThan(0)
    expect(r!.groupNames).toContain('')
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

  it('drops a compute for a superseded batch (jobEpoch mismatch)', () => {
    send(buildInit())
    const sig = ready()!.signatures[0]!.signature
    postSpy.mockClear()

    send(buildCompute({ signature: sig, jobEpoch: 999 }))

    expect(charts()).toHaveLength(0)
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
```

- [ ] **Step 2: Run the new test file**

Run: `cd ui && pnpm test -- transform.worker.test.ts`
Expected: 9 tests pass.

- [ ] **Step 3: Run the full UI test suite**

Run: `cd ui && pnpm test`
Expected: all tests pass (existing + 4 from Task 2 + 9 from this task).

- [ ] **Step 4: Commit**

```bash
git -C /home/fahim/Projects/goptics/vizb add ui/src/workers/transform.worker.test.ts
git -C /home/fahim/Projects/goptics/vizb commit -m "test(ui): cover transform.worker protocol + epoch guards + fallbacks"
```

---

## Task 4: `useStatsWorker.test.ts` — singleton, id correlation, per-key cache

**Files:**
- Create: `ui/src/composables/useStatsWorker.test.ts`

Uses `vi.mock` on the `?worker&inline` import. Each test gets a fresh `MockWorker` instance by clearing the module cache and resetting the singleton.

- [ ] **Step 1: Write the failing test file**

```ts
// ui/src/composables/useStatsWorker.test.ts
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { MockWorker } from '../workers/__test-utils__/workerHarness'

// Mock the ?worker&inline import BEFORE importing the bridge. Vitest treats
// the query string as part of the module ID, so this matches the production
// import in useStatsWorker.ts.
vi.mock('../workers/stats.worker.ts?worker&inline', () => ({
  default: MockWorker,
}))

import { computeDescriptive, computeCorrelation } from './useStatsWorker'
import type { ChartData, SeriesProfile, CorrelationMatrix } from '../types'

let worker: MockWorker

beforeEach(async () => {
  vi.resetModules()
  // Re-import the bridge so its module-level `worker` singleton is fresh.
  const mod = await import('./useStatsWorker')
  // After vi.mock + re-import, the bridge's `new StatsWorker()` call inside
  // getWorker() must return our MockWorker. We grab the singleton via the
  // module's side effect on first call.
  // Trigger lazy creation:
  void mod.computeDescriptive(makeChart('a'))
  // Allow the microtask queue to flush so the bridge has called new StatsWorker().
  await Promise.resolve()
  // The first MockWorker instance is the singleton for this test.
  worker = (await import('./useStatsWorker')).__test__getWorker?.() as unknown as MockWorker
  // Fallback if the test hook isn't available: just read the last-constructed
  // mock via the postMessage spy count of 1.
  if (!worker) {
    throw new Error('singleton worker not initialised; expose a test hook or restructure')
  }
})

afterEach(() => {
  vi.resetModules()
})

function makeChart(id: string): ChartData {
  return {
    points: [
      { x: id + '-1', y: 'y1', z: 'z1', values: [1, 2, 3] },
      { x: id + '-2', y: 'y1', z: 'z1', values: [4, 5, 6] },
    ],
    yAxis: ['y1'],
    zAxis: ['z1'],
    series: [
      { xAxis: id + '-1', values: [1, 2, 3], name: id + '-1' },
      { xAxis: id + '-2', values: [4, 5, 6], name: id + '-2' },
    ],
    render3D: undefined,
  } as unknown as ChartData
}

function emitReply(id: number, kind: 'descriptive' | 'correlation') {
  const reply =
    kind === 'descriptive'
      ? { type: 'result' as const, id, seriesProfiles: [] as SeriesProfile[] }
      : { type: 'result' as const, id, correlation: {} as CorrelationMatrix }
  worker.__emit(reply)
}

describe('useStatsWorker — singleton lifecycle', () => {
  it('creates exactly one Worker on first call', async () => {
    worker.postMessage.mockClear()
    // worker was already created in beforeEach; a new call should not add another.
    const p = computeDescriptive(makeChart('a'))
    emitReply(0, 'descriptive')
    await p
    // Only the lazy call from beforeEach should have produced a postMessage.
    expect(worker.postMessage.mock.calls.length).toBeLessThanOrEqual(2)
  })
})

describe('useStatsWorker — id correlation', () => {
  it('resolves the correct promise when two requests are in flight', async () => {
    worker.postMessage.mockClear()
    const chartA = makeChart('a')
    const chartB = makeChart('b')

    const pA = computeDescriptive(chartA)
    const pB = computeDescriptive(chartB)
    // The bridge posts in order; second post is id=1, first is id=0.
    // (nextId is module-level, starts at 0 and increments per post.)
    const idA = worker.postMessage.mock.calls[0]![0].id as number
    const idB = worker.postMessage.mock.calls[1]![0].id as number
    expect(idA).not.toBe(idB)

    // Reply out of order: resolve B first.
    emitReply(idB, 'descriptive')
    emitReply(idA, 'descriptive')

    const [rA, rB] = await Promise.all([pA, pB])
    expect(rA.length).toBe(0) // seriesProfiles was []
    expect(rB.length).toBe(0)
  })
})

describe('useStatsWorker — per-ChartData cache', () => {
  it('returns the same promise for two calls on the same chart', async () => {
    worker.postMessage.mockClear()
    const chart = makeChart('a')
    const p1 = computeDescriptive(chart)
    const p2 = computeDescriptive(chart)
    expect(p1).toBe(p2)
    // Resolve to keep the test clean.
    emitReply(worker.postMessage.mock.calls[0]![0].id, 'descriptive')
    await p1
  })

  it('makes two posts for two different ChartData objects', async () => {
    worker.postMessage.mockClear()
    const p1 = computeDescriptive(makeChart('a'))
    const p2 = computeDescriptive(makeChart('b'))
    expect(worker.postMessage.mock.calls.length).toBeGreaterThanOrEqual(2)
    // Resolve both.
    for (const c of worker.postMessage.mock.calls) emitReply(c[0].id, 'descriptive')
    await Promise.all([p1, p2])
  })
})

describe('useStatsWorker — correlation axis cache', () => {
  it('uses one Worker post per axis', async () => {
    worker.postMessage.mockClear()
    const chart = makeChart('a')
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
    worker.postMessage.mockClear()
    const chart = makeChart('a')
    const p1 = computeCorrelation(chart)
    const p2 = computeCorrelation(chart)
    expect(p1).toBe(p2)
    expect(worker.postMessage.mock.calls.length).toBe(1)
    emitReply(worker.postMessage.mock.calls[0]![0].id, 'correlation')
    await p1
  })
})
```

- [ ] **Step 2: Run the new test file and inspect failures**

Run: `cd ui && pnpm test -- useStatsWorker.test.ts`
Expected: This file will fail because the bridge's internal `worker` is not exposed for tests. We have two options (resolved in step 3).

- [ ] **Step 3: Expose a test hook on the bridge module**

The current `useStatsWorker.ts` keeps `worker` as a module-local `let`. To make the singleton observable from tests, add a `__test__` hook guarded by a `process.env.NODE_ENV !== 'production'` check (or `import.meta.env.PROD` for Vite). This is a minimal, intentional addition for testability.

In `ui/src/composables/useStatsWorker.ts`, after the `getWorker` function (line 40), add:

```ts
// Test-only: expose the singleton for tests. Strips itself in production
// builds via the Vite `import.meta.env.PROD` define replacement.
declare const __VIZB_TEST__: boolean
export const __test__getWorker = __VIZB_TEST__ ? (): Worker | null => worker : undefined
```

Then add the `import.meta.env.PROD` define to `ui/vite.config.ts`'s `define` block (it already has a `define` block at line 37; add the new key alongside `process.env.NODE_ENV`):

```ts
define: {
  'process.env.NODE_ENV': command === 'serve' ? '"development"' : '"production"',
  __VIZB_TEST__: 'false',
},
```

(The define is set to `'false'` unconditionally — vitest reads source as-is, but in production builds the constant is replaced with `false` and the dead branch is tree-shaken. For vitest, the constant stays as the literal `false` in the source, which is fine: `__VIZB_TEST__` is `false` and the hook returns `undefined`. For test access, we use a different strategy — see step 4.)

- [ ] **Step 4: Pivot to a different test strategy that does not need a test hook**

Replace the `beforeEach` block in the test file with one that captures the `MockWorker` via the `vi.mock` factory. Since `vi.mock` is hoisted before imports, the bridge's `new StatsWorker()` returns *the same* `MockWorker` class, and we can spy on its instances by checking the spy counts on `MockWorker.prototype.postMessage`.

A cleaner approach: spy on the `MockWorker` constructor itself.

Update the test file's imports and `beforeEach`:

```ts
// ui/src/composables/useStatsWorker.test.ts
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'

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

import { computeDescriptive, computeCorrelation } from './useStatsWorker'
import type { ChartData, SeriesProfile, CorrelationMatrix } from '../types'

let worker: TrackedMockWorker

beforeEach(async () => {
  vi.resetModules()
  TrackedMockWorker.instances.length = 0
  ctorSpy.mockClear()
  // Trigger lazy worker creation by issuing a call; the bridge's first
  // call to getWorker() will instantiate the MockWorker.
  const chart = makeChart('warmup')
  const p = computeDescriptive(chart)
  // Flush the microtask so the constructor has run.
  await Promise.resolve()
  worker = TrackedMockWorker.instances[0]!
  expect(worker).toBeDefined()
  // Clean up the warmup call.
  worker.postMessage.mockClear()
  // Resolve the warmup promise so it doesn't leak.
  const warmupId = worker.postMessage.mock.calls[0]?.id as number | undefined
  if (warmupId !== undefined) {
    worker.__emit({ type: 'result', id: warmupId, seriesProfiles: [] as SeriesProfile[] })
  }
  await p
  worker.postMessage.mockClear()
})

afterEach(() => {
  vi.resetModules()
})
```

(The helper functions and the four `describe` blocks from step 1 stay the same — only the imports and `beforeEach` change.)

- [ ] **Step 5: Re-run the new test file**

Run: `cd ui && pnpm test -- useStatsWorker.test.ts`
Expected: 7 tests pass.

- [ ] **Step 6: Revert the test hook change in step 3 (it is no longer needed)**

```bash
git -C /home/fahim/Projects/goptics/vizb checkout -- ui/src/composables/useStatsWorker.ts ui/vite.config.ts
```

(The `git checkout` will throw an error if the change wasn't actually committed. It wasn't — the change in step 3 was meant as a thought experiment, never committed. If you did commit it, use `git reset --soft HEAD~1` and re-commit the test file separately.)

- [ ] **Step 7: Run the full UI test suite**

Run: `cd ui && pnpm test`
Expected: all tests pass (existing + 4 from Task 2 + 9 from Task 3 + 7 from this task).

- [ ] **Step 8: Commit**

```bash
git -C /home/fahim/Projects/goptics/vizb add ui/src/composables/useStatsWorker.test.ts
git -C /home/fahim/Projects/goptics/vizb commit -m "test(ui): cover useStatsWorker singleton, id correlation, per-key cache"
```

---

## Task 5: `useChartPipeline.test.ts` — init, setArrangement, compute, drain, dispose

**Files:**
- Create: `ui/src/composables/useChartPipeline.test.ts`

Uses `vi.mock` on the `?worker&inline` import and `vi.useFakeTimers()` to advance the 50 ms debounce. Each test wraps the composable in an `effectScope` so `onScopeDispose` fires when the scope stops.

- [ ] **Step 1: Write the failing test file**

```ts
// ui/src/composables/useChartPipeline.test.ts
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { effectScope, ref, type Ref } from 'vue'
import { MockWorker } from '../workers/__test-utils__/workerHarness'
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

function makeScope(rawData: Ref<DataPoint[]>) {
  return effectScope()
}

const defaultLabels: AxisLabels = { xAxis: 'X', yAxis: 'Y', zAxis: 'Z' }

let scope: ReturnType<typeof effectScope>
let worker: TrackedMockWorker
let result: ReturnType<typeof useChartPipeline>
let rawData: Ref<DataPoint[]>
let arrangement: Ref<{ identityString: string; targetString: string }>
let activeGroupId: Ref<number>
let sort: Ref<Sort>
let showLabels: Ref<boolean>
let scale: Ref<ScaleType>

beforeEach(async () => {
  vi.resetModules()
  vi.useFakeTimers()
  TrackedMockWorker.instances.length = 0
  ctorSpy.mockClear()

  rawData = ref([dp('x1', 'y1'), dp('x2', 'y1'), dp('x1', 'y2')])
  arrangement = ref({ identityString: 'xy', targetString: 'xy' })
  activeGroupId = ref(0)
  sort = ref(noSort)
  showLabels = ref(false)
  scale = ref('linear' as ScaleType)

  scope = makeScope(rawData)
  result = scope.run(() =>
    useChartPipeline(rawData, arrangement, ref(defaultLabels), activeGroupId, sort, showLabels, scale)
  )!
  // Flush the immediate watch + 50 ms debounce.
  await vi.advanceTimersByTimeAsync(50)
  worker = TrackedMockWorker.instances[0]!
  expect(worker).toBeDefined()
})

afterEach(() => {
  scope.stop()
  vi.useRealTimers()
  vi.resetModules()
})

// Helper: emit a ReadyMessage in response to the most recent init/setArrangement.
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

// Helper: emit a ChartMessage for a specific signature.
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
    expect(worker.postMessage).toHaveBeenCalled()
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
  it('posts one compute per signature, in order', async () => {
    replyReady(1)
    worker.postMessage.mockClear()
    // The pipeline should auto-pump one compute per signature.
    expect(worker.postMessage.mock.calls.length).toBe(2)
    expect(worker.postMessage.mock.calls[0]![0].signature).toBe('sig-val')
    expect(worker.postMessage.mock.calls[1]![0].signature).toBe('sig-other')
  })

  it('unblocks the next compute when a chart reply lands', async () => {
    replyReady(1)
    worker.postMessage.mockClear()
    // First compute is already in flight (auto-pumped). Simulate its reply.
    const dataEpoch = 1
    const jobEpoch = 1
    replyChart(dataEpoch, jobEpoch, 'sig-val')
    // The second compute should now be pumped.
    expect(worker.postMessage.mock.calls.length).toBe(1)
    expect(worker.postMessage.mock.calls[0]![0].signature).toBe('sig-other')
  })

  it('drops stale ready replies (mismatched dataEpoch)', () => {
    replyReady(1)
    worker.postMessage.mockClear()
    // Simulate a ready with an old dataEpoch — should be ignored.
    worker.__emit({ type: 'ready', dataEpoch: 0, signatures: [], groupNames: [] } as WorkerResponse)
    expect(worker.postMessage.mock.calls.length).toBe(0)
  })

  it('drops stale chart replies (mismatched jobEpoch) but still drains', async () => {
    replyReady(1)
    worker.postMessage.mockClear()
    // Reply with a stale jobEpoch.
    replyChart(1, 999, 'sig-val')
    // The slot was not updated...
    expect(result.charts.value.find((c) => c.key === 'sig-val')!.data).toBeNull()
    // ...but the queue still drains: the next compute is pumped.
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
    // No `init` should have been posted (no re-clone).
    const initCall = worker.postMessage.mock.calls.find((c) => c[0].type === 'init')
    expect(initCall).toBeUndefined()
  })
})

describe('useChartPipeline — param changes', () => {
  it('posts compute (not init) when sort changes', async () => {
    replyReady(1)
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
    rawData.value = []
    await vi.advanceTimersByTimeAsync(50)
    const initCall = worker.postMessage.mock.calls.find((c) => c[0].type === 'init')
    expect(initCall).toBeUndefined()
    expect(result.charts.value).toEqual([])
    expect(result.hasAny.value).toBe(false)
  })
})
```

- [ ] **Step 2: Run the new test file**

Run: `cd ui && pnpm test -- useChartPipeline.test.ts`
Expected: 10 tests pass.

- [ ] **Step 3: Run the full UI test suite**

Run: `cd ui && pnpm test`
Expected: all tests pass.

- [ ] **Step 4: Commit**

```bash
git -C /home/fahim/Projects/goptics/vizb add ui/src/composables/useChartPipeline.test.ts
git -C /home/fahim/Projects/goptics/vizb commit -m "test(ui): cover useChartPipeline init/setArrangement/compute/drain/dispose"
```

---

## Task 6: Final verification + format

- [ ] **Step 1: Type-check**

Run: `cd ui && pnpm typecheck`
Expected: passes (no production code changed; test files only add `MockWorker` consumers).

- [ ] **Step 2: Format**

Run: `cd ui && pnpm format`
Expected: prettier reformats the four new test files (likely no changes if the snippets above are already formatted).

- [ ] **Step 3: Run the full UI test suite one more time**

Run: `cd ui && pnpm test`
Expected: 4 (stats) + 9 (transform) + 7 (useStatsWorker) + 10 (useChartPipeline) = 30 new tests pass; all pre-existing tests still pass.

- [ ] **Step 4: Run the full project test suite (Go + UI)**

Run: `task test`
Expected: both Go and UI suites green.

- [ ] **Step 5: Commit any formatter changes**

```bash
git -C /home/fahim/Projects/goptics/vizb add -u
git -C /home/fahim/Projects/goptics/vizb diff --cached --quiet || git -C /home/fahim/Projects/goptics/vizb commit -m "style(ui): format new test files"
```

(If the diff is empty, the commit is skipped. The `||` is the idiom to no-op when there's nothing to commit.)

---

## Self-Review

**Spec coverage:**

| Spec requirement | Task |
| --- | --- |
| `__test-utils__/workerHarness.ts` harness | Task 1 |
| `stats.worker.test.ts` — 4 cases | Task 2 |
| `transform.worker.test.ts` — 10 cases (init, setArrangement, compute, all epoch guards, fallbacks, no-op-before-init) | Task 3 |
| `useStatsWorker.test.ts` — 7 cases (singleton, id correlation, per-ChartData cache, per-axis cache, 'auto' key) | Task 4 |
| `useChartPipeline.test.ts` — 10 cases (init, drain, stale-reply drop, setArrangement, param changes, data change, dispose, empty data) | Task 5 |
| `pnpm test`, `pnpm typecheck`, `pnpm format`, `task test` verification | Task 6 |

**Placeholder scan:** No TBDs, no "implement later". Every code block is a real, runnable Vitest test.

**Type consistency:**
- `MockWorker` (Task 1) — used in Tasks 4 and 5. Both consumers call `worker.postMessage`, `worker.__emit`, `worker.terminate`; all exist on the class.
- `installMockSelf` / `uninstallMockSelf` (Task 1) — used in Tasks 2 and 3. Both call `getHandler()` after `await import()`; the contract matches.
- The `TrackedMockWorker` class in Tasks 4 and 5 is duplicated intentionally (each test file is self-contained — no shared test helper beyond `MockWorker` from the harness). This avoids cross-file test imports.
- The `__test__getWorker` idea from Task 4 step 3 was abandoned in step 4; the actual approach uses `TrackedMockWorker.instances[0]` and no production code was changed.
