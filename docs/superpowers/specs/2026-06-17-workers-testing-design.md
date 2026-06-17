# Worker Test Coverage — Design

**Date:** 2026-06-17
**Status:** Approved (pre-implementation)
**Owner:** QA

## Goal

Add unit tests for the two Web Worker files in `ui/src/workers/` and the two
composable bridges that own them (`useStatsWorker`, `useChartPipeline`). The
pure logic the workers call (`lib/stats.ts`, `lib/transform.ts`) is already
covered; this spec covers the **protocol / dispatch / glue** that surrounds it.

## Scope

Four test files, one shared harness:

| File | Purpose |
| --- | --- |
| `ui/src/workers/__test-utils__/workerHarness.ts` | Shared mock-`self` + mock-`Worker` helpers |
| `ui/src/workers/stats.worker.test.ts` | Stats worker protocol & dispatch |
| `ui/src/workers/transform.worker.test.ts` | Transform worker protocol, dispatch, staleness guards |
| `ui/src/composables/useStatsWorker.test.ts` | Singleton bridge, id correlation, per-`ChartData` cache |
| `ui/src/composables/useChartPipeline.test.ts` | Init / setArrangement / compute flow, debounce, epoch matching, drain |

## Out of scope

- `lib/stats.ts`, `lib/transform.ts` — already tested
- Vite's `?worker&inline` build transformation
- Vue reactivity surface of `useChartPipeline` beyond the protocol
- Browser/jsdom environment — stays in node per `vitest.config.ts`

## Architecture

### Worker test pattern — "mocked self"

Both worker files set `self.onmessage = ...` at import time. To exercise them
in a node test environment:

```ts
// 1. install a fake self on globalThis BEFORE importing the worker
const postSpy = vi.fn()
;(globalThis as any).self = { onmessage: null, postMessage: postSpy }

// 2. import the worker — this runs the module top-level, which assigns
//    self.onmessage = handler
const mod = await import('./transform.worker.ts')

// 3. invoke the captured handler directly
const handler = (globalThis as any).self.onmessage as (e: MessageEvent) => void
handler({ data: msg } as MessageEvent)

// 4. assert on the spy
expect(postSpy).toHaveBeenCalledWith(expect.objectContaining({ type: 'ready' }))
```

This works because:
- The worker module's only top-level side effect is assigning `self.onmessage`.
- The handler is a plain function — no `this` binding issues.
- The same handler is reused for every message, so we can fire multiple
  messages per test.

For `transform.worker.ts` the module also holds a `state` variable. Each test
uses `vi.resetModules()` + a fresh `import()` to get a clean state.

### Bridge test pattern — mocked `?worker&inline` import

The production code does:

```ts
import TransformWorker from '../workers/transform.worker.ts?worker&inline'
const worker = new TransformWorker()
```

Vitest understands the `?worker&inline` query string in `vi.mock` paths, so
we mock the constructor with a class that mirrors the three methods the
bridges actually touch (`onmessage`, `postMessage`, `terminate`):

```ts
vi.mock('../workers/transform.worker.ts?worker&inline', () => ({
  default: class MockTransformWorker {
    onmessage: ((e: MessageEvent) => void) | null = null
    postMessage = vi.fn()
    terminate = vi.fn()
    __emit(msg: WorkerResponse) { this.onmessage?.({ data: msg } as MessageEvent) }
  }
}))
```

The `__emit` test helper is the inverse of `postMessage`: it lets the test
simulate a reply from the worker.

For `useChartPipeline` we also need `vi.useFakeTimers()` so we can advance
through the 50 ms debounce without real wall-clock waits.

## Test cases

### `stats.worker.test.ts`

| Case | Asserts |
| --- | --- |
| `kind: 'descriptive'` request | `postMessage` carries `seriesProfiles`, no `correlation` |
| `kind: 'correlation'` request | `postMessage` carries `correlation`, no `seriesProfiles` |
| Reply `id` matches request `id` | correlation sanity |
| Two back-to-back requests with different `id`s | two replies, each with the right `id` |

### `transform.worker.test.ts`

Each test uses `vi.resetModules()` + fresh `import()` to clear module state.

| Case | Asserts |
| --- | --- |
| `init` with non-empty data | `ready` reply: correct `dataEpoch`, `signatures`, `groupNames` |
| `init` then `setArrangement` | second `ready` reply reflects new `groupNames`, same `dataEpoch` |
| `init` then `compute` | `chart` reply: matching `dataEpoch` + `jobEpoch` + `signature` |
| `compute` with mismatched `dataEpoch` (stale dataset) | no reply |
| `compute` with mismatched `jobEpoch` (stale batch) | no reply |
| `compute` with unknown `signature` | no reply |
| `compute` with unknown `groupName` | still produces a chart (falls back to first group) |
| `setArrangement` before any `init` | no reply, no crash |
| `compute` before any `init` | no reply, no crash |
| `setArrangement` updates `state.labels` when included | re-`ready` carries new labels path |

### `useStatsWorker.test.ts`

| Case | Asserts |
| --- | --- |
| First call to `computeDescriptive` | creates Worker; `postMessage` called once |
| Second call to `computeDescriptive` | no new Worker created, no new `postMessage` for the same key |
| Out-of-order replies | each promise resolves with its own reply's `id` |
| Per-`ChartData` cache | two calls on the same `chartData` → one Worker post |
| Two different `ChartData` objects | two Worker posts (cache is per-key) |
| `computeCorrelation(chart, 'x')` vs `computeCorrelation(chart, 'y')` | two Worker posts (axis is part of the key) |
| `computeCorrelation(chart)` no axis → second call | one Worker post (`'auto'` key cached) |

### `useChartPipeline.test.ts`

Uses `vi.useFakeTimers()`. Each test instantiates the composable inside an
`effectScope` so `onScopeDispose` runs in cleanup.

| Case | Asserts |
| --- | --- |
| Non-empty data passed at call time | `init` posted synchronously (the `watch(..., { immediate: true })` fires), `ready` populates `charts` with skeleton slots |
| `ready` triggers one `compute` per signature | order matches signature list |
| Multiple charts | chart-1 reply unblocks chart-2's `compute` (queue pump) |
| Stale `ready` (mismatched `dataEpoch`) | no chart updates |
| Stale `chart` reply (mismatched `jobEpoch`) | slot not updated, but queue still drains |
| `setArrangement` after init | posts `setArrangement`, no re-clone |
| Param change (sort/scale/showLabels/activeGroupId) | debounced `compute` posts, no `init` |
| Data change | debounced `init` posts, `dataEpoch` bumps |
| `onScopeDispose` | `worker.terminate()` called, debounces cleared |
| Empty data | no `init` post, `charts` cleared, `hasAny` false |

## Shared harness — `ui/src/workers/__test-utils__/workerHarness.ts`

Exports:

- `installMockSelf()` — sets `globalThis.self = { onmessage: null, postMessage: vi.fn() }`,
  returns `{ postSpy, getHandler() }`. `getHandler()` reads `self.onmessage` after
  the worker module is imported.
- `MockWorker` class — for `vi.mock('.../X.worker.ts?worker&inline', ...)`.
  Mirrors a real Worker's surface: `onmessage`, `postMessage` (a `vi.fn()`),
  `terminate` (a `vi.fn()`), plus `__emit(msg)` test helper.

Both test files use it; the harness is ~30 lines, no abstractions over
vitest's primitives.

## Verification

- `pnpm test` (runs `vitest run`)
- `pnpm typecheck` (mock class declarations must satisfy `Worker`-shape typing)
- `pnpm format` (prettier)
- Final: `task test` (also runs the Go suite, but only the UI portion is
  relevant here)

## Risks

- **Module-level state in `transform.worker.ts`** — handled by
  `vi.resetModules()` per test.
- **`globalThis.self` pollution between tests** — handled by `beforeEach` /
  `afterEach` that deletes the property.
- **`?worker&inline` query in `vi.mock`** — vitest does treat the query as
  part of the module ID; if it ever stops, the bridge tests fail loudly
  (mock isn't picked up, real worker code path runs). Acceptable failure mode.

## Estimated test count

- stats.worker: 4
- transform.worker: 10
- useStatsWorker: 7
- useChartPipeline: 10

Total: **31 new tests** across 4 files + 1 harness.
