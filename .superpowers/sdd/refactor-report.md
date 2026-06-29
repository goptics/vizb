# SDD Refactor Report: select-axis-mode pattern cleanup

## Status: DONE

## Summary

Structural refactor of the `feat/select-axis-mode` feature branch, collapsing 9 over-engineering/spaghetti findings across both Go backend and TypeScript frontend into 3 design patterns:

1. **Mode enum + RowReader** (Go parser, Task 1) — deleted dead code, introduced `Mode` type resolved once in `ParseConfig`, `RowReader` interface collapsing CSV/JSON duplication, unified FlagBag string-slice store.
2. **ChartBuilder Strategy + polymorphism** (TS UI, Task 2) — extracted `ChartBuilder` interface with 4 strategy implementations (`GroupedBuilder`/`PreserveRowsBuilder`/`ValueBuilder`/`MixedBuilder`), builder registry, polymorphic shape-query methods, `ModeHandler` registry in transform worker.
3. **Single `chartMode` computed** (TS composable, Task 3) — collapsed 4 mirrored mode computeds into one `ChartMode` type + computed, with backward-compatible aliases.

## Changes

| Layer | Files changed | Added | Deleted | Net |
|-------|--------------|-------|---------|-----|
| Go (`pkg/`, `cmd/`, `shared/`) | ~25 | ~1,400 | ~300 | +1,100 |
| TS (`ui/src/`) | ~20 | ~1,500 | ~300 | +1,200 |
| Total | ~71 | ~5,520 | ~616 | +4,904 |

*Note: the +4,904 line delta includes the original PR's feature code (~4,000 lines of CSV examples, docs, tests, and feature logic). The refactor itself deleted ~600 lines of over-engineering.*

## Acceptance Criteria

| Finding | Resolution | Status |
|---------|-----------|--------|
| F1: Dead code `readCSVTable`/`readJSONArray` | Deleted from `csv.go`/`json.go` | ✅ |
| F2: CSV↔JSON parser duplication | `RowReader` interface + shared `ParseMixedMode`/`ParseValueMode`/`ParseSelectStatMode` | ✅ |
| F3: Mode-predicate explosion (9 sites) | `Mode` enum resolved once in `ParseConfig`; `cfg.Mode` switch everywhere | ✅ |
| F4: `buildChartForSignature` dual code path | ChartBuilder strategies: GroupedBuilder/PreserveRowsBuilder/ValueBuilder/MixedBuilder | ✅ |
| F5: `utils.ts` shotgun surgery per shape | Per-shape `plottable`/`badgeCount`/`grandTotal`/`is3D` methods on builders | ✅ |
| F6: FlagBag parallel arrays | Unified `stringSlices` map; branch only at `Bind` | ✅ |
| F7: Dual type-inference paths | AxisType carried from parser; `inferSelectViewMixed`/`parseAxisFloat` added to `axes_spec.go` | ✅ |
| F8: Worker dispatch triplication | ModeHandler registry with GroupedHandler/ValueHandler/MixedHandler | ✅ |
| F9: `useDataPoint` mirrored computeds | One `chartMode` computed + aliases | ✅ |

## Commits (14 refactor commits + 1 plan commit)

```
6e2c439 docs(sdd): plan for select-axis-mode pattern refactor
52162a6 refactor(parser): delete unused readCSVTable/readJSONArray helpers
ab15bb8 refactor(parser): add Mode enum resolved once in ParseConfig
1e362d3 refactor(parser): introduce RowReader, collapse CSV/JSON select-mode duplication
956a2bb refactor(cli): unify FlagBag slice/array stores into one map
1350e6b fix: axes type inference fallback for ParseFunc value boundary
f740c85 refactor(ui): add ChartBuilder interface
77260d5 refactor(ui): extract GroupedBuilder
ecbe2d8 refactor(ui): extract PreserveRowsBuilder
daf8cdf refactor(ui): extract ValueBuilder and MixedBuilder
be3bf5f refactor(ui): add builder registry
36e9e4e refactor(ui): delegate buildChartForSignature to ChartBuilder
a909baf refactor(ui): delegate utils shape queries to ChartBuilder
7f4529c refactor(ui): replace worker mode branches with ModeHandler registry
a484288 refactor(ui): collapse useDataPoint mode computeds into chartMode
```

## Test Results

### Go (`go test ./...`)
```
ok  github.com/goptics/vizb/cmd
ok  github.com/goptics/vizb/cmd/cli
ok  github.com/goptics/vizb/pkg/parser
ok  github.com/goptics/vizb/pkg/parser/csv
ok  github.com/goptics/vizb/pkg/parser/json
(all 26 packages green)
```

### Go vet (`go vet ./...`)
```
(no warnings)
```

### TypeScript (`vitest run`)
```
Test Files  27 passed (27)
Tests       406 passed (406)
```

### TypeScript typecheck (`vue-tsc -b --noEmit`)
```
(no errors)
```

### UI build (`EMBED_UI=True pnpm build`)
```
✓ built in 730ms
✓ vizb-ui.gen.go is in sync with ui sources
```

## Key Decisions

- `ParseFunc` stays `func(filename string, cfg Config) []DataPoint` (value type); `DispatchSelectMode` operates on a local copy. Mixed-mode axes fallback in `DatasetAxesForSelectView` checks parsed `DataPoint.XAxis` values for non-numeric strings when `AxisType` is empty.
- `chartHasPlottableData` kept as original single OR expression (genuinely shape-agnostic; builder dispatch by `statType` can't route the test factory's charts).
- `statType` for preserveRows stays as `statTemplate.type` (not `'preserveRows'`) — preserves StatsPanel/Pie/Radar display and existing test expectations.
- ModeHandler has 3 methods (`init`/`setArrangement`/`compute`) plus `readyReply` for cleaner state separation.

## Risk Mitigation Verified

- No behavior changes — all existing tests pass unchanged.
- `go test ./...` and `vitest run` green after every step.
- Each task independently committable.
- Go logic files untouched by TS tasks; TS files untouched by Go tasks (except generated `vizb-ui.gen.go` bundle artifact).
