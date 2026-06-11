# Requirement: Move all chart computation off the main thread (swap-axis fix + async grouping)

> Spec for subagent-driven development. Self-contained — assume no prior conversation.
> Scope: `ui/` Vue app only. Build with `task build:ui && task build:cli` (gen.go is
> compiled into the Go binary). Debug logs: `console.log/info/warn/debug/trace` are
> STRIPPED in prod builds (`ui/vite.config.ts` `esbuild.pure`) — use `console.error`
> for temporary instrumentation, or `task dev:ui`.

## 1. Problem & root cause (confirmed)

Vizb renders benchmark/CSV data as charts. A Web Worker (`ui/src/workers/transform.worker.ts`,
driven by `ui/src/composables/useChartPipeline.ts`) computes charts off-thread. Sort/scale
already work via the worker.

**Swap-axis is broken and laggy.** Repro:
`./bin/vizb examples/csv/sales.csv -g order_date,category -p x/n -o /tmp/swap.html`.
Each row is `{ name=category, xAxis=order_date, stats[] }`.

- **Grouping runs on the main thread** (`useDataPoint.ts` `grouped` computed): it groups
  rows by `name` and STRIPS `name` out, then hands one group's `name`-stripped rows to
  the worker. So the worker never sees `name`.
- A swap such as `nx → yx` ("move the value currently on `name` to the `y` axis") cannot
  be satisfied by the worker — `name` isn't in its copy — so the rebuilt chart is byte
  identical (verified: `chart reply … nSeries=731 series0="2024-01-01"` unchanged before
  and after; epochs matched, so nothing was dropped). Result: **"nothing changes."**
- **The lag** is the synchronous `swapAxisFields(activeDataSet.value.data, …)` in
  `AxisSwapper.vue` mutating ~500k **reactive Vue proxies** (delete+assign), which also
  forces the `grouped` computed to rebuild. Sort never touches main data, so sort isn't laggy.

`ui/src/lib/swap.ts` (helpers `translateAxisKey`, `swapAxisFields`, `swapAxisLabels`) was
missing and has been created this session; the build now compiles. The worker `swap`
message path it feeds is the broken design being removed.

## 2. Goal / invariant

**After initial load, the main thread performs NO data computation for any setting change.**
Swap, group-select, sort, scale, and show-labels are each a small message to the worker;
ALL work — projection, **grouping**, sorting, 3D build — runs in the worker. The main
thread only posts a param and renders the returned `ChartData`. The one unavoidable
main-thread data cost is the **one-time clone** of the raw dataset into the worker at
`init` (and again only when switching to a different dataset).

`useDataPoint.grouped` / `dataPointGroups` / `activeGroup` (main-thread 500k grouping) are
**removed**. The group list comes back from the worker as a plain string array.

## 3. Design

Rows never change on swap/group/sort/scale — only *how they're read*. So the worker owns
the full raw dataset and treats **arrangement + group + sort + scale + showLabels all as
compute params** (exactly how `sort` already works). Only a **dataset switch** re-clones.

The worker caches a `Map<groupName, DataPoint[]>`: rebuilt once on `init` and once per
**arrangement** change (off-thread, one pass); `compute` jobs read the cache.

### Arrangement model
- `identityString` = present source axes in canonical `n,x,y,z` order (e.g. `"nx"`).
- `targetString` = the selected arrangement (e.g. `"yx"`); same length, same value set.
- `identityKeys = translateAxisKey(identityString)` → source fields `['name','xAxis']`.
- `targetKeys  = translateAxisKey(targetString)`  → display axes `['yAxis','xAxis']`.
- Projection (non-mutating): `displayRow[targetKeys[i]] = rawRow[identityKeys[i]]`, carry `stats`.
- Group-by field = `identityKeys[i]` where `targetKeys[i] === 'name'`; if none → single
  `"Default"` group.
- Labels: `swapAxisLabels(identityString, targetString, baseLabels)` (cheap, 4 keys) →
  display labels, sent to the worker to attach to each `ChartData`.

Reuse `translateAxisKey` / `swapAxisLabels` from `ui/src/lib/swap.ts`. The mutating
`swapAxisFields` is replaced by the pure projection. Reuse `listChartSignatures` and
`buildChartForSignature` in `ui/src/lib/transform.ts` unchanged (stat signatures are
arrangement-independent).

### Worker protocol (`transform.worker.ts`)
State: `{ dataEpoch, raw: DataPoint[], grouped: Map<string,DataPoint[]>, groupNames: string[], labels }`.
- `init {dataEpoch, data}` → store raw; project+group under identity; reply
  `ready {dataEpoch, signatures, groupNames}`.
- `setArrangement {identityString, targetString, labels}` → re-project + re-group raw;
  reply `ready {dataEpoch, signatures, groupNames}`. No clone.
- `compute {dataEpoch, jobEpoch, signature, groupName, sort, showLabels, scale}` → build
  from `grouped.get(groupName)` via `buildChartForSignature`; reply
  `chart {dataEpoch, jobEpoch, signature, chart}`.
- Remove old `swap` message + `swapAxisFields` import.

### `lib/transform.ts`
- Add pure `projectAndGroup(raw, identityKeys, targetKeys): { grouped: Map<string,DataPoint[]>, groupNames: string[] }`
  (one pass; arrangement-aware replacement for `useDataPoint.grouped`).

### `useChartPipeline.ts`
- Inputs: `rawData` (`activeDataSet.data`), arrangement refs (identity+target strings) +
  display labels, `groupName` ref, plus existing `sort/showLabels/scale`.
- `dataEpoch` reinit (the one clone) fires only on **rawData/dataset change**.
- Params path (arrangement / group / sort / scale / showLabels) bumps `jobEpoch`; when the
  arrangement changed it posts `setArrangement` and re-queues on the resulting `ready`;
  otherwise it re-queues computes against the cache (existing `recompute`).
- Expose a reactive `groupNames` ref from `ready` (drives selector + URL router).
- Remove `triggerSwap` / `swap` message / `swapPending`.

### `useDataPoint.ts`
- Remove `grouped` / `dataPointGroups` / `activeGroup` array building.
- Add per-dataset `arrangementKey` state (default = identity) + setter (used by AxisSwapper);
  expose `activeArrangement`.
- `resultGroups` / group count now sourced from the pipeline's `groupNames`. Keep
  `activeGroupId` / `selectGroup` (index into `groupNames`); map index ↔ name for the
  worker `groupName` param.

### `AxisSwapper.vue`
- Keep `swapOptions` / `presentKeys` / identity (cheap `.some()` reads).
- `handleSwapSelect`: set the arrangement key (→ pipeline posts `setArrangement`), persist
  `selectedSwapIndex`, reset group to 0. **Delete** the `swapAxisFields(activeDataSet.value.data…)`
  call and the `activeDataSet.value.axisLabels =` mutation (labels now derived from arrangement).

### `Dashboard.vue`
- Pass `activeDataSet.data` + arrangement + `groupName` to `useChartPipeline`; feed
  `DataSetHeader` `resultGroups` from the pipeline's `groupNames`.

### `useUrlRouter.ts` / `useDashboardInit.ts`
- `applyParams` reads group count from `groupNames` (now worker-provided / async). Apply the
  `?g=` param after the first `ready` (gate on `groupNames.length`); `?d=`/settings as-is.

`DataSetHeader.vue` unchanged (still receives `resultGroups` / `activeGroupId`).

### Cleanup (current working tree has temporary debug state)
- Remove temporary `console.log`s in `transform.worker.ts` and `useChartPipeline.ts`.
- Restore `ui/vite.config.ts` `esbuild.pure` (the user removed it to see logs) — verify
  against `git diff`.

## 4. Task breakdown (ordered for subagent-driven development)

Each task: implement, typecheck (`ui/node_modules/.bin/vue-tsc -b ui`), keep the app
compiling. Land in order (later tasks depend on earlier types/signatures).

1. **transform helpers** — add `projectAndGroup` to `lib/transform.ts` (+ small node test
   of groupNames + projected fields on `/tmp/sales_small.csv`-shaped input).
   *AC:* identity arrangement reproduces today's grouping; `nx→yx` yields single
   `"Default"` group with `yAxis=category`.
2. **worker protocol** — rewrite `transform.worker.ts` to own raw data, `init` /
   `setArrangement` / `compute`, returning `groupNames`. Remove `swap`.
   *AC:* unit-style check via a small node harness importing the worker logic functions.
3. **pipeline** — rework `useChartPipeline.ts`: new inputs, reinit-only-on-dataset,
   params path posts `setArrangement`/`recompute`, expose `groupNames`. Remove
   `triggerSwap`/`swapPending`.
   *AC:* types compile; sort/scale path unchanged in behavior.
4. **data store** — `useDataPoint.ts`: drop main-thread grouping; add `arrangementKey`
   state; source groups from pipeline. Wire `Dashboard.vue`.
   *AC:* app renders; group selector populated from worker `groupNames`.
5. **swap UI** — `AxisSwapper.vue`: set arrangement, drop mutations.
   *AC:* swapping changes the chart; no main-thread row mutation.
6. **URL router** — `useUrlRouter.ts`/`useDashboardInit.ts`: async-safe group apply.
   *AC:* `?g=`/`?d=` deep-links still resolve.
7. **cleanup + verify** — remove debug logs, restore `vite.config.ts`; full build; manual
   repro per §5.

## 5. Verification

1. `ui/node_modules/.bin/vue-tsc -b ui` → exit 0.
2. `task build:ui && task build:cli`.
3. `./bin/vizb examples/csv/sales.csv -g order_date,category -p x/n -o /tmp/swap.html`,
   open in a browser:
   - swap `nx → yx`: chart actually changes (category becomes a y-series);
   - swap/group/sort/scale clicks are responsive — no multi-second freeze;
   - group selector lists groups and switches; `?g=` / `?d=` deep-links work;
     sort/scale/labels still work and compose with a swap.
4. **Invariant check** (DevTools Performance): a swap/group/sort/scale click shows no long
   main-thread task over the row data — only postMessage + worker task + render.
5. `task test`.

## 6. Key files
- `ui/src/lib/transform.ts`, `ui/src/lib/swap.ts`
- `ui/src/workers/transform.worker.ts`
- `ui/src/composables/useChartPipeline.ts`, `useDataPoint.ts`, `useUrlRouter.ts`, `useDashboardInit.ts`
- `ui/src/components/AxisSwapper.vue`, `DataSetHeader.vue`
- `ui/src/views/Dashboard.vue`
- `ui/vite.config.ts` (restore), `ui/src/types/index.ts` (types as needed)
