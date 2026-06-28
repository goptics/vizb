# Task 4 Report: UI mixed-mode rendering

## Status

**Complete**

## Commits

- `8892f72` — `feat(ui): add mixed-axis scatter transform and rendering`

## Summary

Implemented the UI mixed-axis scatter path (category X + value Y/Z) per the approved design spec. Mixed mode runs as a third parallel transform alongside grouped and pure value modes.

### Changes

1. **`ui/src/lib/utils.ts`** — `isMixedMode`, `isMixedModeChart`, `isScatterTransformMode`, `mixedModeHasZAxis`; updated `is3D`, `canOfferValue3D`, `chartHasPlottableData`, `chartAxisBadgeCount`, `computeChartGrandTotal`.
2. **`ui/src/lib/transform.ts`** — `buildMixedModeChart` producing `mixedTuples` + `xCategories` (2D) or `render3D.mode: 'mixed'` (3D).
3. **`ui/src/workers/transform.worker.ts`** — `__mixed_mode__` synthetic signature routing (init/compute/setArrangement).
4. **`ui/src/composables/useChartPipeline.ts`** — mixed-mode skeleton slots on reinit.
5. **`ui/src/composables/charts/shared/mixedMode.ts`** — `buildMixedAxes2DOptions` (category X + value Y scatter).
6. **`ui/src/composables/charts/useScatterChartOptions.ts`** — routes to mixed 2D builder when `mixedTuples` present.
7. **`ui/src/composables/charts/useScatter3DChartOptions.ts`** + **`shared/3d.ts`** — mixed 3D scatter (`category` X, `value` Y/Z, `render3D.mode: 'mixed'` grid).
8. **`ui/src/components/SettingsPanel.vue`** — hides sort and swap in mixed mode (mirrors value mode; swap stays value-only).
9. **`ui/src/composables/useDataPoint.ts`** — exports `isMixedMode` / `isMixedModeDataset`.
10. **`ui/src/types/index.ts`** — `mixedTuples`, `xCategories`, `Render3D.mode: 'mixed'`.
11. **`pkg/template/vizb-ui.gen.go`** — regenerated via `task build:ui`.

### Tests

- `isMixedMode`, `buildMixedModeChart`, worker `__mixed_mode__` routing
- Scatter 2D mixed options (category X + value Y)
- Multi-dataset array normalization smoke test
- `cd ui && npm test` — **396 passed**
- `task build:ui` — **PASS**

## Acceptance

| Criterion | Result |
|-----------|--------|
| Mixed scatter renders category X + value Y/Z | ✅ |
| `isMixedMode` + `buildMixedModeChart` in UI | ✅ |
| Worker/pipeline `__mixed_mode__` routing | ✅ |
| Settings hide sort/swap in mixed mode | ✅ |
| UI tests pass | ✅ |
| `task build:ui` after UI change | ✅ |

## Notes

- Mixed 3D is intrinsic when axes include value `z` (`is3D` detects via `mixedModeHasZAxis`); no `--3d` toggle needed.
- `feat/mixed-axis-scatter` had no UI diff for these files — implementation follows `docs/superpowers/specs/2026-06-28-mixed-axis-scatter-design.md`.

## Out of scope (unchanged)

- Docs/examples (Task 5)
- Go backend mixed 3D auto-enable flag (value-only `autoEnableValueMode3D` unchanged)