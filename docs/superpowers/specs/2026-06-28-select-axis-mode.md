# Design spec: `--select` axis mode

**Status:** Implemented (supersedes mixed-axis scatter `--axes` design)  
**Branch:** `feat/select-axis-mode`  
**Date:** 2026-06-28

## Summary

User-facing axis control for CSV/JSON tabular data uses **`--select` only**. The former `--axes` flag is not exposed. Solo `--select` (no explicit grouping) enters **select axis mode**: each repeatable flag value defines one 2–3 column view mapped to `x`, `y`, and optional `z`. Grouped `--select` (with `-g` / `-p` / `-r`) keeps the existing behaviour: pick numeric stat columns only.

## Supersedes

- Mixed-axis scatter design that introduced a separate `--axes` CLI flag.
- Docs that described `--select` only as a grouped stat-column filter or as a narrow override on auto-value.

## Mode at a glance

| Mode | Trigger | `--select` role | Auto-group | Auto-value | Chart notes |
|------|---------|-----------------|------------|------------|-------------|
| **Grouped** | `-g`, `-r`, or non-default `-p` | Numeric stat columns; optional `{label}` | Off (explicit) | Off | Standard grouped bar/line/scatter |
| **Solo value** | `--select` only; all selected cols numeric | 2–3 cols → `x,y[,z]` value axes | Off | Off | Bar/line/scatter value mode; 3 cols → auto-3D |
| **Solo mixed** | `--select` only; one categorical `x`, numeric `y[,z]` | 2–3 cols → category `x` + value `y[,z]` | Off | Off | **Scatter primary**; category X + continuous Y/Z |
| **Auto-group** | No flags; mixed cat + numeric file | — | On | Off | Picks highest-cardinality categorical column |
| **Auto-value** | No flags; all-numeric file | — | Off | On | First 2–3 numeric cols as coordinates |

## Solo `--select` rules

1. **Repeatable** — each `--select` produces one dataset view. Multiple views embed as an array in HTML; JSON emits an array when `N > 1`.
2. **2–3 columns per view** — positional assignment: 1st → `x`, 2nd → `y`, 3rd → `z`.
3. **Explicit syntax** — `x:col,y:col[,z:col]` for all columns or none; mixed implicit/explicit is fatal.
4. **Categorical constraint** — at most one categorical column; it must be on `x`. Use `x:region,y:latency` when inference is ambiguous.
5. **Disables inference** — solo `--select` turns off auto-group and auto-value for that run.
6. **Parsers** — `csv` and `json` only; other parsers warn and ignore.

## Grouped `--select` rules (unchanged)

- Requires explicit grouping (`-g`, `-r`, or non-default `-p`).
- Selected columns must be numeric (stat series).
- Optional `{label}` rename; quoted names for special characters.
- Cannot overlap with `--group` columns.

## Mixed vs value routing

After parsing column tokens, `ResolveAxesTypes` classifies each axis column:

- **x**: category if any cell is non-numeric; otherwise value.
- **y**, **z**: must be numeric (at least one numeric cell).

| Classification | Mode | Data shape |
|----------------|------|------------|
| All value | Value | `XAxis`, `YAxis`, `ZAxis` hold numeric strings; `Stats` empty |
| One category on x + value y[,z] | Mixed | `XAxis` categorical; `YAxis`/`ZAxis` numeric; `Stats` empty |

Mixed rendering is implemented in the scatter chart path (`buildMixedModeChart`). Bar and line accept mixed-axis data at parse time; scatter is the intended chart for category × value plots.

## Multi-dataset pipeline

```
--select view₁  --select view₂  …
        │                │
        ▼                ▼
  SelectViewData[]  (one slice per view)
        │
        ▼
  assembleDatasets → []*Dataset
        │
        ▼
  writeOutput: HTML embeds array; JSON array when N>1
```

- Dataset names auto-generated: `region × latency`, `region × sales`, …
- Per-view 3D auto-enable: a 3-column view enables 3D for that dataset only.
- `--id` suffixes: `my-id`, `my-id-2`, … for additional views.

## Examples

```bash
# Solo mixed scatter (category region, value latency)
vizb scatter examples/csv/region-metrics.csv --select region,latency -o mixed.html

# Solo value scatter (all numeric)
vizb scatter examples/csv/spiral-3d.csv --select x,y -o xy.html

# Multi-view HTML (two datasets, one file)
vizb scatter examples/csv/region-metrics.csv \
  --select region,latency \
  --select region,sales \
  -o dual.html

# Grouped stat pick (unchanged)
vizb bar examples/csv/sales.csv -g region,product -p x,y --select amount,total -o bars.html

# Explicit axis placement
vizb scatter data.csv --select x:region,y:latency,z:sales{Revenue} -o out.html
```

## Implementation map

| Component | Role |
|-----------|------|
| `cmd/cli/dataflags.go` | Repeatable `KindStringArray` for `--select` |
| `pkg/parser/select_view_spec.go` | `ParseSelectViewFlag`, `IsSelectAxisMode`, `SelectViewDatasetName` |
| `pkg/parser/axes_spec.go` | `ResolveAxesTypes`, `IsMixedMode`, `MixedAxes`, `ValueAxes` |
| `pkg/parser/csv/csv.go`, `json/json.go` | Route select axis / mixed / value; `ParseSelectViews` |
| `cmd/cli/pipeline.go` | `prepareDataViews`, `assembleDatasets`, multi-dataset output |
| `ui/src/lib/transform.ts` | `buildMixedModeChart` for scatter mixed rendering |

## Acceptance

- [x] Solo `--select region,latency` does not auto-group.
- [x] Grouped `-g … --select price` still picks numeric stats.
- [x] Multi `--select` → multiple datasets in one HTML.
- [x] Mixed scatter renders category X + value Y/Z.
- [x] Docs: mode-at-a-glance table; three-role `--select` in chart guides.