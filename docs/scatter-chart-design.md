# Scatter Chart Extraction — Design Summary

**Date:** 2026-06-20  
**Status:** Implemented (Tasks 1–8)

## Problem

Bar and line charts carried two distinct rendering modes: **grouped** (categorical x/y, stats as z) and **value** (`--axes`, raw numeric coordinates). Value mode complicated bar/line option builders, 3D detection, and settings UX. Scatter is the natural home for coordinate plotting.

## Solution

Extract value-mode and hybrid plotting into a first-class **scatter** chart type across Go CLI, parsers, UI renderers, and bundle pruning.

## Architecture

### Go backend

| Layer | Responsibility |
|-------|----------------|
| `config/charts/scatter` | Typed `Config` (mirrors line: scale, 3D flags, sort, labels) |
| `cmd/charts/scatter` | `vizb scatter` subcommand; **only** chart with `--axes` |
| `pkg/parser` | Three parse paths when axes present: **value** (2–3 numeric cols), **hybrid** (2 group + 1 axes col), standard group |
| `shared/chart_selection` | Scatter in `ValidChartTypes`; 3D-capable (z axis or `--3d` value-mode toggle) |

### CLI axes modes

1. **Pure value** — `--axes price,latency` (no `--group`): each row → `(x,y[,z])` on value axes.
2. **Hybrid** — `--group cat1,cat2 --axes z`: categorical x/y + numeric z in stats.
3. **Grouped** — standard `--group` path; scatter renders like line but with `scatter` series type.

`--axes` removed from bar/line; scatter validates arity, rejects overlap with `--group`/`--select`.

### UI

| Component | Role |
|-----------|------|
| `ChartScatter.vue` / `Chart3D` | Lazy 2D/3D renderers |
| `useScatterChartOptions` | Group + value-mode 2D (`valueTuples`) |
| `useScatter3DChartOptions` | Continuous, grouped, hybrid 3D (`scatter3D`) |
| `valueMode.ts` | Scatter-only tuple sort/scale/axis builders |
| `transform.worker` | Value/hybrid init+compute gated to `chartType === 'scatter'` |
| `useChartPipeline` | Forwards chart type; skeleton slots for value/hybrid |

Bar/line retain **category `--3d`** pseudo-3D only; continuous `--axes x,y,z` is scatter-exclusive.

### Bundle pruning

`ChartScatter-<hash>.js` mapped in `CHART_ROOT_PREFIX` → `VizbChartRoots["scatter"]`. `SelectChunks` gates scatter renderer like other chart roots; `needs3D` pulls echarts-gl stack.

## Data model

- **Value axes metadata:** `{ key, label, type: "value" }` on dataset `axes`.
- **Hybrid z:** stored in `DataPoint.Stats` (matched by z axis label), not `ZAxis` string.
- **Demos:** `sample.json` Value Mode + Spiral 3D use `type: "scatter"` with value axes preserved.

## Invariants

- `isValueMode` / `isHybridMode` / `isScatterTransformMode` in UI utils scope transforms.
- Settings panel hides sort/swap for scatter value/hybrid modes.
- `canOfferValue3D` offers grouped pseudo-3D for bar/line/scatter; excludes value/hybrid axes.
- Default `--charts` remains `bar,line,pie`; scatter is opt-in via `-c scatter` or `vizb scatter`.

## Test coverage

- Go: scatter config/CLI, parser value+hybrid paths, `chart_selection` 3D cases, chunk pruning.
- UI: scatter 2D/3D option builders, pipeline/worker routing, settings field registry.