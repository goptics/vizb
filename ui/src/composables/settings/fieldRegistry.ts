import type { Component } from 'vue'
import type { ChartConfig, ChartType } from '../../types'
import type { Dimension } from '../../lib/utils'
import SortControl from '../../components/settings/SortControl.vue'
import ScaleControl from '../../components/settings/ScaleControl.vue'
import ShowLabelsControl from '../../components/settings/ShowLabelsControl.vue'
import AutoRotateControl from '../../components/settings/AutoRotateControl.vue'
import SwapControl from '../../components/settings/SwapControl.vue'

// Field registry: maps a JSON field name to the control component that renders
// it. `SettingsPanel.vue` uses `getRenderableFields(activeConfig, ctx)` to
// discover which fields are AVAILABLE for the active chart — the result is
// filtered by:
//   - `appliesTo` : which chart types see the field (e.g. pie vs bar)
//   - `appliesOn` : which data shape the field makes sense for
//     (e.g. `autoRotate` is 3D-only — it writes `grid3D.viewControl.autoRotate`
//     and has no visual effect on a 2D bar/line chart).
//
// `appliesOn` is a list of dimensions on which the field is applicable.
// `undefined` (or the property being absent) means "no dimension constraint" —
// the field renders on every dimension. When the active data's dimension is
// unknown (empty dataset) the constraint is also skipped, so the panel still
// shows every field by default until data arrives.
//
// Adding a new field = one entry here + a Vue control file. Adding a new
// chart type = update the `appliesTo` matrix on existing fields. Adding a new
// dimension-specific field = set `appliesOn: ['3D']` (or whatever the
// constraint is).
type FieldMeta = {
  component: Component
  appliesTo: ChartType[]
  appliesOn?: Dimension[]
}

export const fieldRegistry: Record<string, FieldMeta> = {
  sort: {
    component: SortControl,
    appliesTo: ['bar', 'line', 'pie', 'heatmap', 'radar'],
  },
  scale: {
    component: ScaleControl,
    appliesTo: ['bar', 'line'],
  },
  showLabels: {
    component: ShowLabelsControl,
    appliesTo: ['bar', 'line', 'pie', 'heatmap', 'radar'],
  },
  autoRotate: {
    component: AutoRotateControl,
    appliesTo: ['bar', 'line'],
    appliesOn: ['3D'],
  },
  swap: {
    component: SwapControl,
    appliesTo: ['bar', 'line', 'pie', 'heatmap', 'radar'],
  },
}

// Returns the control component registered for `key`, or undefined if the key
// is the `type` discriminator or an unrecognised field.
export function getControl(key: string): Component | undefined {
  if (key === 'type') return undefined
  const meta = fieldRegistry[key]
  return meta?.component
}

// Renderable field metadata for the active chart's config: an ordered list of
// `{ key, component }` pairs for every registered field that applies to the
// chart's `type` AND (if constrained) the active data's dimension. Independent
// of which fields the config currently populates — the panel shows all
// available fields, each control displays in default/off state until the user
// opts in.
export type RenderableField = {
  key: string
  component: Component
}

export type RenderContext = {
  // Active data's dimensionality. `undefined` means unknown / not yet loaded;
  // the dimension constraint is skipped in that case so the panel still shows
  // every field by default.
  dimension?: Dimension
}

export function getRenderableFields(
  config: ChartConfig,
  ctx: RenderContext = {}
): RenderableField[] {
  const fields: RenderableField[] = []
  for (const [key, meta] of Object.entries(fieldRegistry)) {
    if (!meta.appliesTo.includes(config.type)) continue
    if (meta.appliesOn && ctx.dimension && !meta.appliesOn.includes(ctx.dimension)) {
      continue
    }
    fields.push({ key, component: meta.component })
  }
  return fields
}
