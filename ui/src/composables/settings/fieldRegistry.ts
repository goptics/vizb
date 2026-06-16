import { defineAsyncComponent, type Component } from 'vue'
import type { ChartConfig } from '../../types'

// Field registry: maps a JSON field name to the control component that renders
// it. `SettingsPanel.vue` walks `Object.keys(activeConfig)` and looks each key
// up here — unknown keys are silently skipped (no error, no broken UI). Adding
// a new field with its own control = one new line in this map plus a Vue file.
// Controls are loaded async so each one only parses when an active chart's
// config exposes its field; matches the rest of the codebase's code-splitting
// pattern (see `RENDERERS` in `ChartCard.vue`).
export const fieldRegistry: Record<string, Component> = {
  sort: defineAsyncComponent(() => import('../../components/settings/SortControl.vue')),
  scale: defineAsyncComponent(() => import('../../components/settings/ScaleControl.vue')),
  showLabels: defineAsyncComponent(() => import('../../components/settings/ShowLabelsControl.vue')),
  autoRotate: defineAsyncComponent(() => import('../../components/settings/AutoRotateControl.vue')),
  swap: defineAsyncComponent(() => import('../../components/settings/SwapControl.vue')),
}

// Returns the control component registered for `key`, or undefined if the key
// is the `type` discriminator or an unrecognised field. SettingsPanel treats
// `undefined` as "skip this key" — the panel renders only what the config and
// the registry both know about.
export function getControl(key: string): Component | undefined {
  if (key === 'type') return undefined
  return fieldRegistry[key]
}

// Renderable field metadata for a single chart config: an ordered list of
// `{ key, component }` pairs for every key present on the config that has a
// registered control. Preserves `Object.keys` order so the panel layout follows
// the field order declared on the typed `Config` interface.
export type RenderableField = {
  key: string
  component: Component
}

export function getRenderableFields(config: ChartConfig): RenderableField[] {
  const fields: RenderableField[] = []
  for (const key of Object.keys(config)) {
    const component = getControl(key)
    if (component) fields.push({ key, component })
  }
  return fields
}
