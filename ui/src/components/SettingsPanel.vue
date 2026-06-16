<script setup lang="ts">
import { computed, type Component } from 'vue'
import { BarChart3, TrendingUp, PieChart, Table, Radar } from 'lucide-vue-next'
import { Card, CardContent, CardHeader, CardTitle, Separator } from './ui'
import SelectionTabs from './SelectionTabs.vue'
import Selector from './Selector.vue'
import SettingHeader from './SettingHeader.vue'
import { useSettingsStore } from '../composables/useSettingsStore'
import { useDataPoint } from '../composables/useDataPoint'
import { resetColor } from '../lib/utils'
import {
  getRenderableFields,
  shouldUseTabPicker,
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
  setShowLabels,
  setAutoRotate,
  setSwap,
} = useSettingsStore()

const { activeDataSet, activeDataSetId, activeDataDimension, setArrangement, activeGroupId } = useDataPoint()

const CHART_ICONS: Record<ChartType, Component> = {
  bar: BarChart3,
  line: TrendingUp,
  pie: PieChart,
  heatmap: Table,
  radar: Radar,
}

// Chart-type picker. The picker is shown only when the dataset bundles more
// than one chart type. The control on the right side branches on the count:
// ≤ CHART_PICKER_TAB_THRESHOLD (3) -> button row (SelectionTabs) that grows to
// fill the row; > threshold -> icon+name combobox (Selector). The shared
// "Chart type" header keeps the row's silhouette stable across the threshold
// so a user navigating between datasets doesn't see a layout jump.
const availableTypes = computed<ChartType[]>(
  () => activeDataSet.value?.settings.map((s) => s.type) ?? []
)
const showChartTypeSelection = computed(() => availableTypes.value.length > 1)
const useTabPicker = computed(() => shouldUseTabPicker(availableTypes.value.length))
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
const onChartTypeChange = (val: string | number) => setChartType(val as ChartType)
const onChartTypeSelect = (id: number) => {
  const opt = chartOptions.value[id]
  if (opt) setChartType(opt.value)
}

// The field list for the active chart. `getRenderableFields` walks the config's
// own keys (filtered by `appliesTo` + the active data's dimension) so pie/
// heatmap/radar configs naturally skip `scale` and `autoRotate` without any
// chart-type check here, and a 2D bar config also skips `autoRotate` (3D-only).
const fields = computed(() => {
  const cfg = activeConfig.value
  return cfg
    ? getRenderableFields(cfg, { dimension: activeDataDimension.value })
    : []
})

// Each control emits `update:modelValue` with the appropriate type for its
// field. The store's setters handle the writeback to `dataset.value.settings[i]`
// and ignore writes for fields that don't exist on the active chart's config
// (e.g. `setScale` on a pie config). Swap has extra side effects beyond the
// wire format: it must also update useDataPoint's arrangement (which the
// pipeline watches to post `setArrangement` to the worker so it re-projects /
// re-groups off-thread) and reset the group + recolor on a new arrangement.
const handlers: Record<string, (val: unknown) => void> = {
  sort: (val) => setSort(val as Parameters<typeof setSort>[0]),
  scale: (val) => setScale(val as Parameters<typeof setScale>[0]),
  showLabels: (val) => setShowLabels(val as boolean),
  autoRotate: (val) => setAutoRotate(val as boolean),
  swap: (val) => {
    const target = val as string | undefined
    if (target === undefined) return
    setArrangement(activeDataSetId.value, chartType.value, target)
    activeGroupId.value = 0
    resetColor()
    setSwap(target)
  },
}

const valueFor = (key: string) =>
  activeConfig.value ? (activeConfig.value as Record<string, unknown>)[key] : undefined

const onUpdate = (key: string, value: unknown) => {
  const fn = handlers[key]
  if (fn) fn(value)
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
          <SelectionTabs
            v-if="useTabPicker"
            :model-value="chartType"
            :options="chartOptions"
            @update:model-value="onChartTypeChange"
          />
          <Selector
            v-else
            :items="chartOptions.map((o) => ({ name: o.label, icon: o.icon }))"
            :activeId="activeChartTypeIndex"
            @select="onChartTypeSelect"
            class="w-40"
          />
        </div>
        <Separator />
      </template>

      <template v-for="field in fields" :key="field.key">
        <component
          :is="field.component"
          :model-value="valueFor(field.key)"
          @update:model-value="(val: unknown) => onUpdate(field.key, val)"
        />
      </template>
    </CardContent>
  </Card>
</template>
