# Task 3 Report: Multi-dataset pipeline for `--select` axis mode

## Status

**Complete**

## Commits

- `0ab3c51` — `feat(cli): multi-dataset pipeline for repeatable --select views`

## Summary

Implemented N-view parsing and assembly when `len(SelectViews) > 1`: one file read, one dataset per `--select`, shared meta/settings, per-view axes and 3D auto-enable.

### Changes

1. **`select_view_spec.go`** — `SelectViewData`, `SelectViewDatasetName` (auto-names like `region × latency`).
2. **`csv.go` / `json.go`** — `ParseSelectViews` reads input once and parses each `SelectViews[i]` through existing mixed/value paths.
3. **`pipeline.go`**
   - `prepareDataViews` — multi-view path via `ParseSelectViews`; single-view delegates to `prepareData`.
   - `assembleDatasets` / `buildDataset` — one `*shared.Dataset` per view; `cloneChartConfigs` isolates per-view 3D flags.
   - `RunLinear` — branches on `IsSelectAxisMode && len(SelectViews) > 1`; swap/rules run per dataset.
   - `writeOutput([]*shared.Dataset)` — JSON array when N>1, single object when N=1; HTML embeds marshaled payload (array when N>1).
4. **Tests** — `prepareDataViews`, `assembleDatasets` per-view 3D, multi-dataset JSON, end-to-end HTML with two embedded datasets.

### Tests

- `go test ./...` — **PASS**
- `TestRunLinearMultiSelectProducesTwoDatasetsInHTML` — `--select region,latency --select region,sales` → 2 entries in `window.VIZB_DATA`
- `TestAssembleDatasetsMultiSelectPerView3D` — value xyz view enables 3D; mixed view does not

## Acceptance

| Criterion | Result |
|-----------|--------|
| `--select region,latency --select region,sales` → 2 datasets in one HTML | ✅ |
| Single file read for N views | ✅ |
| Per-view 3D auto-enable | ✅ |
| JSON backward compat (single object when N=1) | ✅ |

## Concerns

1. **Multi `--id`** — secondary datasets get suffixed ids (`id-2`, `id-3`) when `--id` is set; first keeps the raw value.
2. **Non-csv/json multi `--select`** — warns and falls back to first view only.
3. **Dataset naming** — multi-view uses auto names; single-view keeps `--name` when provided.

## Out of scope (unchanged)

- UI mixed-mode rendering (Task 4)
- Docs (Task 5)