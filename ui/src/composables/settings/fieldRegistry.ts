import type { Component } from 'vue'
import type { ChartConfig, ChartType, ScaleType, Sort } from '@/types'
import type { Dimension } from '@/lib/utils'
import SortControl from '@/components/settings/SortControl.vue'
import ScaleControl from '@/components/settings/ScaleControl.vue'
import BooleanControl from '@/components/settings/BooleanControl.vue'
import SwapControl from '@/components/settings/SwapControl.vue'

/** Value type each settings control emits for its field key. */
export type SettingFieldValueMap = {
  sort: Sort
  scale: ScaleType
  stack: boolean
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
  visible?: (ctx: RenderContext) => boolean
  id?: string
  label?: string
  description?: string
  separator?: boolean
}

export const fieldRegistry: Record<SettingFieldKey, FieldMeta> = {
  sort: {
    component: SortControl,
    appliesTo: ['bar', 'line', 'scatter', 'pie', 'heatmap', 'radar', 'sankey', 'chord'],
  },
  scale: {
    component: ScaleControl,
    appliesTo: ['bar', 'line', 'scatter'],
  },
  stack: {
    component: BooleanControl,
    appliesTo: ['bar', 'line'],
    id: 'stack-switch',
    label: 'Stack series',
    description: 'Stack grouped bar or line series into category totals.',
    separator: true,
    visible: (ctx) =>
      ctx.rendering3D !== true &&
      (ctx.dimension === undefined || ctx.dimension === '2D') &&
      ctx.chartMode !== 'value' &&
      ctx.chartMode !== 'mixed',
  },
  showLabels: {
    component: BooleanControl,
    appliesTo: ['bar', 'line', 'scatter', 'pie', 'heatmap', 'radar', 'sankey', 'chord'],
    id: 'labels-switch',
    label: 'Show labels',
    description: 'Display data labels on chart elements.',
    separator: true,
  },
  smooth: {
    component: BooleanControl,
    appliesTo: ['line'],
    id: 'smooth-lines-switch',
    label: 'Smooth lines',
    description: 'Render curved segments between line chart points.',
    separator: true,
    visible: (ctx) => ctx.rendering3D !== true,
  },
  horizontal: {
    component: BooleanControl,
    appliesTo: ['bar'],
    id: 'horizontal-bars-switch',
    label: 'Horizontal bars',
    description: 'Swap axes so bars grow rightward and categories appear on the Y axis.',
    separator: true,
    visible: (ctx) => ctx.rendering3D !== true,
  },
  threeD: {
    component: BooleanControl,
    appliesTo: ['bar', 'line', 'scatter'],
    id: 'three-d-switch',
    label: '3D view',
    description: 'Render as a 3D chart with metric height on the z axis.',
    // Value-mode toggle when the 3D engine is bundled and z is off chart axes.
    visible: (ctx) => ctx.hasThreeDOption === true && ctx.hasZAxis !== true,
  },
  threeDVisualMap: {
    component: BooleanControl,
    appliesTo: ['bar', 'line', 'scatter'],
    id: 'three-d-visualmap-switch',
    label: 'Visual map',
    description: 'Color bars and lines by metric value using a gradient scale.',
    visible: (ctx) => ctx.rendering3D === true || ctx.dimension === undefined,
  },
  visualMap: {
    component: BooleanControl,
    appliesTo: ['scatter'],
    id: 'visualmap-switch',
    label: 'Visual map',
    description: 'Color scatter points by metric value using a gradient scale.',
    visible: (ctx) => ctx.rendering3D !== true,
  },
  threeDRotate: {
    component: BooleanControl,
    appliesTo: ['bar', 'line', 'scatter'],
    id: 'three-d-rotate-switch',
    label: 'Auto rotate',
    description: 'Continuously rotate the 3D chart.',
    visible: (ctx) => ctx.rendering3D === true || ctx.dimension === undefined,
  },
  swap: {
    component: SwapControl,
    appliesTo: ['bar', 'line', 'scatter', 'pie', 'heatmap', 'radar', 'sankey', 'chord'],
  },
}

export type RenderableField<K extends SettingFieldKey = SettingFieldKey> = {
  key: K
  component: Component
  id?: string
  label?: string
  description?: string
  separator?: boolean
}

export type RenderContext = {
  dimension?: Dimension
  rendering3D?: boolean
  hasThreeDOption?: boolean
  /** z mapped to chart zAxis in the active swap (not raw-data z presence). */
  hasZAxis?: boolean
  chartMode?: 'grouped' | 'value' | 'mixed'
}

export function getRenderableFields(
  config: ChartConfig,
  ctx: RenderContext = {}
): RenderableField[] {
  const fields: RenderableField[] = []
  for (const key of Object.keys(fieldRegistry) as SettingFieldKey[]) {
    const meta = fieldRegistry[key]
    if (!meta.appliesTo.includes(config.type)) continue
    if (meta.visible && !meta.visible(ctx)) continue
    fields.push({
      key,
      component: meta.component,
      id: meta.id,
      label: meta.label,
      description: meta.description,
      separator: meta.separator,
    })
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
