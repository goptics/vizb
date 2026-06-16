import type { Component } from 'vue'
import type { ChartConfig, ChartType } from '../../types'
import SortControl from '../../components/settings/SortControl.vue'
import ScaleControl from '../../components/settings/ScaleControl.vue'
import ShowLabelsControl from '../../components/settings/ShowLabelsControl.vue'
import AutoRotateControl from '../../components/settings/AutoRotateControl.vue'
import SwapControl from '../../components/settings/SwapControl.vue'

// Field registry: maps a JSON field name to the control component that renders
// it. `SettingsPanel.vue` uses `getRenderableFields(activeConfig)` to discover
// which fields are AVAILABLE for the active chart type — independent of which
// fields are currently populated in the config. A control renders in its
// default/off state when the field is absent; the user can opt in interactively.
// Adding a new field = one entry here + a Vue control file. Adding a new chart
// type = update the `appliesTo` matrix on existing fields.
type FieldMeta = {
  component: Component
  appliesTo: ChartType[]
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
// chart's `type`. Independent of which fields the config currently populates —
// the panel shows all available fields, each control displays in default/off
// state until the user opts in.
export type RenderableField = {
  key: string
  component: Component
}

export function getRenderableFields(config: ChartConfig): RenderableField[] {
  const fields: RenderableField[] = []
  for (const [key, meta] of Object.entries(fieldRegistry)) {
    if (meta.appliesTo.includes(config.type)) {
      fields.push({ key, component: meta.component })
    }
  }
  return fields
}
