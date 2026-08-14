<script setup lang="ts">
import { computed, type Component } from 'vue'
import {
  BarChart3,
  TrendingUp,
  CircleDot,
  PieChart,
  Table,
  Radar,
  GitBranch,
  Circle,
} from 'lucide-vue-next'
import { Card, CardContent, CardHeader, CardTitle, Separator } from './ui'
import Selector from './Selector.vue'
import SettingHeader from './SettingHeader.vue'
import { useSettingsStore } from '../composables/useSettingsStore'
import { useDataPoint } from '../composables/useDataPoint'
import { useActiveChartShape } from '../composables/useActiveChartShape'
import { resetColor } from '../lib/utils'
import {
  getRenderableFields,
  partitionRenderableFields,
  type SettingFieldKey,
  type SettingFieldValueMap,
} from '../composables/settings/fieldRegistry'
import type { ChartType } from '../types'

// Generic, schema-less settings panel: walks `Object.keys(activeConfig)` via
// `getRenderableFields` and renders the registered control for each key. The
// only chart-type-aware element is the picker at the top that switches the
// active chart — the field rendering itself is fully data-driven (no
// `if/else` on chart type). Adding a new field = one entry in `fieldRegistry`;
// adding a new chart type requires no change here at all.
const {
  activeConfig,
  chartType,
  setChartType,
  setSort,
  setScale,
  setStack,
  setShowLabels,
  setSmooth,
  setHorizontal,
  setThreeDRotate,
  setSwap,
  setThreeD,
  setThreeDVisualMap,
  setVisualMap,
} = useSettingsStore()

const {
  activeDataset,
  activeDatasetId,
  activeDataDimension,
  setArrangement,
  activeGroupId,
  isValueMode,
  chartMode,
} = useDataPoint()

const { hasZOnChart, hasThreeDOption, threeD, stack } = useActiveChartShape()

const CHART_ICONS: Record<ChartType, Component> = {
  bar: BarChart3,
  line: TrendingUp,
  scatter: CircleDot,
  pie: PieChart,
  heatmap: Table,
  radar: Radar,
  sankey: GitBranch,
  chord: Circle,
}

// Chart-type picker. Shown only when the dataset bundles more than one chart type.
const availableTypes = computed<ChartType[]>(
  () => activeDataset.value?.settings.map((s) => s.type) ?? []
)
const showChartTypeSelection = computed(() => availableTypes.value.length > 1)
const chartOptions = computed(() =>
  availableTypes.value.map((type) => ({
    value: type,
    label: type.charAt(0).toUpperCase() + type.slice(1),
    icon: CHART_ICONS[type] ?? BarChart3,
  }))
)
const activeChartTypeIndex = computed(() =>
  chartOptions.value.findIndex((o) => o.value === chartType.value)
)
const onChartTypeSelect = (id: number) => {
  const opt = chartOptions.value[id]
  if (opt) setChartType(opt.value)
}

const rendering3D = computed(() => hasZOnChart.value || threeD.value)

const fieldGroups = computed(() => {
  const cfg = activeConfig.value
  if (!cfg) return { general: [], threeD: [] }
  const fields = getRenderableFields(cfg, {
    dimension: activeDataDimension.value,
    rendering3D: rendering3D.value,
    hasThreeDOption: hasThreeDOption.value,
    hasZAxis: hasZOnChart.value,
    chartMode: chartMode.value,
  })
  return partitionRenderableFields(fields)
})

// Value/mixed axes: hide sort; swap only for pure value mode.
const filterTransformModeFields = computed(() => chartMode.value !== 'grouped')

const filteredGeneral = computed(() => {
  if (!filterTransformModeFields.value) return fieldGroups.value.general
  return fieldGroups.value.general.filter((f) => {
    if (f.key === 'sort') return false
    if (f.key === 'swap') return isValueMode.value
    return true
  })
})

// Each control emits `update:modelValue` with the appropriate type for its
// field. The store's setters handle the writeback to `dataset.value.settings[i]`
// and ignore writes for fields that don't exist on the active chart's config
// (e.g. `setScale` on a pie config). Swap has extra side effects beyond the
// wire format: it must also update useDataPoint's arrangement (which the
// pipeline watches to post `setArrangement` to the worker so it re-projects /
// re-groups off-thread) and reset the group + recolor on a new arrangement.
const handlers = {
  sort: setSort,
  scale: setScale,
  stack: setStack,
  showLabels: setShowLabels,
  smooth: setSmooth,
  horizontal: setHorizontal,
  threeDRotate: setThreeDRotate,
  threeD: setThreeD,
  threeDVisualMap: setThreeDVisualMap,
  visualMap: setVisualMap,
  swap: (target: SettingFieldValueMap['swap']) => {
    if (target === undefined) return
    setArrangement(activeDatasetId.value, chartType.value, target)
    activeGroupId.value = 0
    resetColor()
    setSwap(target)
  },
}

const valueFor = (key: SettingFieldKey) => {
  if (!activeConfig.value) return undefined
  if (key === 'scale' && stack.value) return 'linear'
  return (activeConfig.value as Partial<SettingFieldValueMap>)[key]
}

const disabledFor = (key: SettingFieldKey) => key === 'scale' && stack.value

const onUpdate = (key: SettingFieldKey, value: unknown) => {
  ;(handlers[key] as (val: unknown) => void)(value)
}
</script>

<template>
  <Card class="w-full">
    <CardHeader>
      <CardTitle class="text-lg">Settings</CardTitle>
    </CardHeader>
    <CardContent class="space-y-4">
      <template v-if="showChartTypeSelection">
        <div class="flex items-center justify-between">
          <SettingHeader label="Chart type" description="Switch the active chart." />
          <Selector
            :items="chartOptions.map((o) => ({ name: o.label, icon: o.icon }))"
            :activeId="activeChartTypeIndex"
            @select="onChartTypeSelect"
            class="w-36"
          />
        </div>
        <Separator />
      </template>

      <template v-for="field in filteredGeneral" :key="field.key">
        <component
          :is="field.component"
          :model-value="valueFor(field.key)"
          :disabled="disabledFor(field.key)"
          :id="field.id"
          :label="field.label"
          :description="field.description"
          :separator="field.separator"
          @update:model-value="(val: unknown) => onUpdate(field.key, val)"
        />
      </template>

      <template v-if="fieldGroups.threeD.length > 0">
        <Separator />
        <template v-for="field in fieldGroups.threeD" :key="field.key">
          <component
            :is="field.component"
            :model-value="valueFor(field.key)"
            :disabled="disabledFor(field.key)"
            :id="field.id"
            :label="field.label"
            :description="field.description"
            :separator="field.separator"
            @update:model-value="(val: unknown) => onUpdate(field.key, val)"
          />
        </template>
      </template>
    </CardContent>
  </Card>
</template>
