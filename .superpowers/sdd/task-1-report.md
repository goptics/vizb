# Task 1 Report: Flag + inference gate (Phase A)

## Status: DONE

## Commit

Conventional commit on `feat/select-axis-mode` after `go test ./...` green.

## Summary

Implemented Phase A of the `--select` axis mode feature: repeatable flag wiring, config fields, inference gates, and ParseConfig routing between grouped numeric select vs solo axis views.

## Changes

### Flag infrastructure

- **`internal/flags/flag.go`**: Added `KindStringArray` for repeatable string flags (mirrors `--chart` pattern).
- **`cmd/cli/dataflags.go`**: Changed `--select` from `KindString` to `KindStringArray`; updated usage text.
- **`cmd/cli/flagbag.go`**:
  - Added `arrays` map backing `KindStringArray`.
  - `Bind`/`Reset`/`StringArray()` support.
  - **ParseConfig routing**:
    - `IsExplicitGrouping(cfg)` → merge all `--select` occurrences into `cfg.Select` (preserving order, cross-occurrence duplicate check).
    - `!IsExplicitGrouping` → each `--select` → one `cfg.SelectViews` entry via `ParseSelectViewFlag` (2–3 cols validated).

### Config & helpers

- **`pkg/parser/registry.go`**: Added `SelectViews [][]ColumnSpec`.
- **`pkg/parser/select_spec.go`**: Added `AxisKey` field to `ColumnSpec`.
- **`pkg/parser/select_view_spec.go`** (new):
  - `ParseSelectViewFlag` — 2–3 cols, positional x/y/z keys, optional `x:col` syntax; reuses `tokenizeSelectFlag`.
  - `HasSelect`, `IsSelectAxisMode`, `IsExplicitGrouping`.

### Inference gates

- **`pkg/parser/group_spec.go`**: `NoExplicitGrouping` returns false when `HasSelect(cfg)` (refactored to use `IsExplicitGrouping`).
- **`cmd/cli/pipeline.go`**: Skips `cfg.AutoGroup = true` when `HasSelect(cfg)`.
- **`pkg/parser/csv/csv.go`** / **`pkg/parser/json/json.go`**: Skip `AutoDetectTabularConfig` when `HasSelect(cfg)`.

### Tests

- **`pkg/parser/select_view_spec_test.go`** (new): ParseSelectViewFlag arity/syntax, HasSelect, IsExplicitGrouping, IsSelectAxisMode.
- **`pkg/parser/group_spec_test.go`**: NoExplicitGrouping false for grouped select and solo SelectViews.
- **`cmd/cli/flagbag_test.go`**: Grouped vs solo routing, repeatable merge, arity rejection.
- **`pkg/parser/csv/csv_test.go`** / **`pkg/parser/json/json_test.go`**: Renamed `TestSelectScopesAutoDetect` → `TestSelectSkipsAutoDetect` — verifies solo SelectViews disables auto-value (parser routing deferred to Task 2).

## Acceptance criteria

| Criterion | Result |
|-----------|--------|
| `vizb scatter sales.csv --select region,latency` does NOT set AutoGroup | ✅ `NoExplicitGrouping` false + pipeline `HasSelect` guard |
| `vizb bar sales.csv -g region --select price` uses grouped numeric select | ✅ `IsExplicitGrouping` → `cfg.Select` |
| Grouped non-numeric select still fatal | ✅ Unchanged `resolveExplicitChartColumns` gate |
| All existing tests pass + new tests green | ✅ `go test ./...` |

## Test run

```
go test ./...
ok  github.com/goptics/vizb/cmd
ok  github.com/goptics/vizb/cmd/cli
ok  github.com/goptics/vizb/pkg/parser
ok  github.com/goptics/vizb/pkg/parser/csv
ok  github.com/goptics/vizb/pkg/parser/json
(all packages green)
```

## Out of scope (Task 2+)

- Mixed/value parser routing for `SelectViews` (csv/json still use flat-series fallback).
- Multi-dataset pipeline (`prepareDataViews`).
- UI mixed-mode rendering.

## Concerns

None blocking. Solo `--select` currently populates `SelectViews` and disables inference but does not yet produce axis-placed data points — expected until Task 2.