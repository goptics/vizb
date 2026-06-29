<script setup lang="ts">
import { computed, type Component } from 'vue'
import { BarChart3, TrendingUp, CircleDot, PieChart, Table, Radar } from 'lucide-vue-next'
import { Card, CardContent, CardHeader, CardTitle, Separator } from './ui'
import Selector from './Selector.vue'
import SettingHeader from './SettingHeader.vue'
import { useSettingsStore } from '../composables/useSettingsStore'
import { useDataPoint } from '../composables/useDataPoint'
import { resetColor } from '../lib/utils'
import {
  getRenderableFields,
  partitionRenderableFields,
  type SettingFieldKey,
  type SettingFieldValueMap,
} from '../composables/settings/fieldRegistry'
import { arrangementHasChartZ } from '../lib/swap'
import { canOfferValue3D } from '../lib/utils'
import type { BarConfig, ChartType, LineConfig, ScatterConfig } from '../types'

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
  setShowLabels,
  setThreeDRotate,
  setSwap,
  setThreeD,
  setThreeDVisualMap,
  setVisualMap,
} = useSettingsStore()

const {
  activeDataSet,
  activeDataSetId,
  activeDataDimension,
  activeArrangement,
  getArrangement,
  setArrangement,
  activeGroupId,
  isValueMode,
  chartMode,
} = useDataPoint()

const CHART_ICONS: Record<ChartType, Component> = {
  bar: BarChart3,
  line: TrendingUp,
  scatter: CircleDot,
  pie: PieChart,
  heatmap: Table,
  radar: Radar,
}

// Chart-type picker. Shown only when the dataset bundles more than one chart type.
const availableTypes = computed<ChartType[]>(
  () => activeDataSet.value?.settings.map((s) => s.type) ?? []
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

// z on a chart axis under the effective swap (map → wire → identity).
const effectiveSwapTarget = computed(() => {
  const fromMap = getArrangement(activeDataSetId.value, chartType.value)
  if (fromMap) return fromMap
  const wire = (activeConfig.value as BarConfig | LineConfig | ScatterConfig | undefined)?.swap
  return wire || activeArrangement.value.targetString
})
const hasZAxis = computed(() => arrangementHasChartZ(effectiveSwapTarget.value))

const hasThreeDOption = computed(() =>
  canOfferValue3D(
    chartType.value,
    activeDataSet.value?.data,
    hasZAxis.value,
    activeConfig.value as BarConfig | LineConfig | ScatterConfig | undefined,
    activeDataSet.value?.axes
  )
)

const rendering3D = computed(() => {
  const cfg = activeConfig.value as BarConfig | LineConfig | ScatterConfig | undefined
  return hasZAxis.value || cfg?.threeD === true
})

const fieldGroups = computed(() => {
  const cfg = activeConfig.value
  if (!cfg) return { general: [], threeD: [] }
  const fields = getRenderableFields(cfg, {
    dimension: activeDataDimension.value,
    rendering3D: rendering3D.value,
    hasThreeDOption: hasThreeDOption.value,
    hasZAxis: hasZAxis.value,
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
type SettingsHandlers = {
  [K in SettingFieldKey]: (val: SettingFieldValueMap[K]) => void
}

const handlers: SettingsHandlers = {
  sort: setSort,
  scale: setScale,
  showLabels: setShowLabels,
  threeDRotate: setThreeDRotate,
  threeD: setThreeD,
  threeDVisualMap: setThreeDVisualMap,
  visualMap: setVisualMap,
  swap: (target) => {
    if (target === undefined) return
    setArrangement(activeDataSetId.value, chartType.value, target)
    activeGroupId.value = 0
    resetColor()
    setSwap(target)
  },
}

const valueFor = (key: SettingFieldKey) =>
  activeConfig.value ? (activeConfig.value as Partial<SettingFieldValueMap>)[key] : undefined

// Generic keeps key/value correlated; Vue emits force a single cast at the boundary.
const updateSetting = <K extends SettingFieldKey>(key: K, value: SettingFieldValueMap[K]) => {
  handlers[key](value)
}

const onUpdate = (key: SettingFieldKey, value: unknown) => {
  updateSetting(key, value as SettingFieldValueMap[SettingFieldKey])
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
          @update:model-value="(val: unknown) => onUpdate(field.key, val)"
        />
      </template>

      <template v-if="fieldGroups.threeD.length > 0">
        <Separator />
        <template v-for="field in fieldGroups.threeD" :key="field.key">
          <component
            :is="field.component"
            :model-value="valueFor(field.key)"
            @update:model-value="(val: unknown) => onUpdate(field.key, val)"
          />
        </template>
      </template>
    </CardContent>
  </Card>
</template>
