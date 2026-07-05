import type { Component } from 'vue'
import type { ChartConfig, ChartType, ScaleType, Sort } from '@/types'
import type { Dimension } from '@/lib/utils'
import SortControl from '@/components/settings/SortControl.vue'
import ScaleControl from '@/components/settings/ScaleControl.vue'
import ShowLabelsControl from '@/components/settings/ShowLabelsControl.vue'
import SmoothControl from '@/components/settings/SmoothControl.vue'
import HorizontalControl from '@/components/settings/HorizontalControl.vue'
import ThreeDRotateControl from '@/components/settings/ThreeDRotateControl.vue'
import ThreeDControl from '@/components/settings/ThreeDControl.vue'
import ThreeDVisualMapControl from '@/components/settings/ThreeDVisualMapControl.vue'
import VisualMapControl from '@/components/settings/VisualMapControl.vue'
import SwapControl from '@/components/settings/SwapControl.vue'

// Re-exported so SettingsPanel can import the chart-type picker threshold
// from the same module as the field registry — both are "what to render in the
// settings panel" decisions.
export { shouldUseTabPicker, CHART_PICKER_TAB_THRESHOLD } from '@/lib/pickerRule'

/** Value type each settings control emits for its field key. */
export type SettingFieldValueMap = {
  sort: Sort
  scale: ScaleType
  showLabels: boolean
  smooth: boolean
  horizontal: boolean
  threeDRotate: boolean
  threeD: boolean
  threeDVisualMap: boolean
  visualMap: boolean
  swap: string | undefined
}

export type SettingFieldKey = keyof SettingFieldValueMap

/** 3D-related settings rendered in a dedicated panel section. */
export const THREE_D_FIELD_KEYS: readonly SettingFieldKey[] = [
  'threeD',
  'threeDVisualMap',
  'threeDRotate',
]

type FieldMeta = {
  component: Component
  appliesTo: ChartType[]
  appliesOn?: Dimension[]
  visible?: (ctx: RenderContext) => boolean
}

export const fieldRegistry: Record<SettingFieldKey, FieldMeta> = {
  sort: {
    component: SortControl,
    appliesTo: ['bar', 'line', 'scatter', 'pie', 'heatmap', 'radar'],
  },
  scale: {
    component: ScaleControl,
    appliesTo: ['bar', 'line', 'scatter'],
  },
  showLabels: {
    component: ShowLabelsControl,
    appliesTo: ['bar', 'line', 'scatter', 'pie', 'heatmap', 'radar'],
  },
  smooth: {
    component: SmoothControl,
    appliesTo: ['line'],
    visible: (ctx) => ctx.rendering3D !== true,
  },
  horizontal: {
    component: HorizontalControl,
    appliesTo: ['bar'],
    visible: (ctx) => ctx.rendering3D !== true,
  },
  threeD: {
    component: ThreeDControl,
    appliesTo: ['bar', 'line', 'scatter'],
    // Value-mode toggle when the 3D engine is bundled and z is off chart axes.
    visible: (ctx) => ctx.hasThreeDOption === true && ctx.hasZAxis !== true,
  },
  threeDVisualMap: {
    component: ThreeDVisualMapControl,
    appliesTo: ['bar', 'line', 'scatter'],
    visible: (ctx) => ctx.rendering3D === true || ctx.dimension === undefined,
  },
  visualMap: {
    component: VisualMapControl,
    appliesTo: ['scatter'],
    visible: (ctx) => ctx.rendering3D !== true,
  },
  threeDRotate: {
    component: ThreeDRotateControl,
    appliesTo: ['bar', 'line', 'scatter'],
    visible: (ctx) => ctx.rendering3D === true || ctx.dimension === undefined,
  },
  swap: {
    component: SwapControl,
    appliesTo: ['bar', 'line', 'scatter', 'pie', 'heatmap', 'radar'],
  },
}

export function getControl(key: string): Component | undefined {
  if (key === 'type' || !(key in fieldRegistry)) return undefined
  return fieldRegistry[key as SettingFieldKey].component
}

export type RenderableField<K extends SettingFieldKey = SettingFieldKey> = {
  key: K
  component: Component
}

export type RenderContext = {
  dimension?: Dimension
  rendering3D?: boolean
  hasThreeDOption?: boolean
  /** z mapped to chart zAxis in the active swap (not raw-data z presence). */
  hasZAxis?: boolean
}

export function getRenderableFields(
  config: ChartConfig,
  ctx: RenderContext = {}
): RenderableField[] {
  const fields: RenderableField[] = []
  for (const key of Object.keys(fieldRegistry) as SettingFieldKey[]) {
    const meta = fieldRegistry[key]
    if (!meta.appliesTo.includes(config.type)) continue
    if (meta.visible) {
      if (!meta.visible(ctx)) continue
    } else if (meta.appliesOn && ctx.dimension && !meta.appliesOn.includes(ctx.dimension)) {
      continue
    }
    fields.push({ key, component: meta.component })
  }
  return fields
}

export function partitionRenderableFields(fields: RenderableField[]) {
  const threeDKeySet = new Set<string>(THREE_D_FIELD_KEYS)
  const threeD = THREE_D_FIELD_KEYS.flatMap((key) => {
    const field = fields.find((f) => f.key === key)
    return field ? [field] : []
  })
  const general = fields.filter((f) => !threeDKeySet.has(f.key))
  return { general, threeD }
}
