# Task 2 Report: Parser repoint for `--select` axis mode

## Status

**Complete**

## Commits

- `2201cc1` — `feat(parser): route solo --select through mixed/value axis parsers`

## Summary

Ported mixed/value axis parsing from `feat/mixed-axis-scatter` (`a3ecd8f`) and wired it to solo `--select` (`SelectViews[0]`) instead of `--axes`.

### Changes

1. **`ColumnSpec`** — added `AxisType` (`category` | `value`).
2. **`axes_spec.go`** — `ResolveAxesTypes`, `IsMixedMode`, `MixedAxes`; updated `ValueAxes` to honor `AxisKey`; added `SelectViewAxesCfg`, `DatasetAxesForSelectView`.
3. **`csv.go` / `json.go`** — when `IsSelectAxisMode(cfg)`, copy `SelectViews[0]` → `Axes`, resolve types, route to mixed or value parse path. Auto-value `cfg.Axes` path unchanged.
4. **`pipeline.go`** — `assembleDataset` uses `DatasetAxesForSelectView` for select axis mode; `autoEnableValueMode3D` for value xyz.
5. **`select_view_spec.go`** — fixed `IsExplicitGrouping` so empty `GroupPattern` is not treated as explicit grouping (required for `IsSelectAxisMode` with default config).

### Tests

- `go test ./...` — **PASS**
- CSV/JSON: select mixed (`region,latency`) and value (`x,y`) modes
- Pipeline: select view mixed/value axis assembly + 3D auto-enable
- Grouped `--select` path unchanged (existing grouped tests still pass)

## Acceptance

| Criterion | Result |
|-----------|--------|
| Solo `--select region,latency` → mixed-axis points (category x, value y) | ✅ |
| Solo `--select x,y` all-numeric → value mode | ✅ |
| Grouped path unchanged | ✅ |

## Concerns

1. **`DatasetAxesForSelectView` infers mixed vs value from parsed results** (non-numeric x → mixed) because `ResolveAxesTypes` runs on a local cfg copy inside parsers and does not mutate caller `SelectViews`. Works for normal parse→assemble flow; edge case: empty results default to value axes.
2. **Error messages** use `--select` vs `--axes` via `AxisColumnLabel` depending on routing path.
3. **`IsExplicitGrouping` fix** was necessary for Task 2; empty `GroupPattern` previously blocked `IsSelectAxisMode` — worth noting in Task 1 review.

## Out of scope (unchanged)

- Multi `SelectViews` pipeline (Task 3)
- UI changes (Task 4)