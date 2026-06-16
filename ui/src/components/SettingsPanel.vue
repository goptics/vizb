<script setup lang="ts">
import { computed, type Component } from 'vue'
import { BarChart3, TrendingUp, PieChart, Table, Radar } from 'lucide-vue-next'
import { Card, CardContent, CardHeader, CardTitle, Separator } from './ui'
import SelectionTabs from './SelectionTabs.vue'
import { useSettingsStore } from '../composables/useSettingsStore'
import { getRenderableFields } from '../composables/settings/fieldRegistry'
import { activeDataSet } from '../composables/useDataPoint'
import type { ChartType } from '../types'

// Generic, schema-less settings panel: walks `Object.keys(activeConfig)` via
// `getRenderableFields` and renders the registered control for each key. The
// only chart-type-aware element is the tabs row that switches the active
// chart — the field rendering itself is fully data-driven (no `if/else` on
// chart type). Adding a new field = one entry in `fieldRegistry`; adding a
// new chart type requires no change here at all.
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

const CHART_ICONS: Record<ChartType, Component> = {
  bar: BarChart3,
  line: TrendingUp,
  pie: PieChart,
  heatmap: Table,
  radar: Radar,
}

// Chart-type tabs at the top — only shown when the dataset bundles more than
// one chart type. Selecting a tab updates the store's active index.
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
const onChartTypeChange = (val: string | number) => setChartType(val as ChartType)

// The field list for the active chart. `getRenderableFields` walks the config's
// own keys, so pie/heatmap/radar configs naturally skip `scale` and `autoRotate`
// without any chart-type check here.
const fields = computed(() => {
  const cfg = activeConfig.value
  return cfg ? getRenderableFields(cfg) : []
})

// Each control emits `update:modelValue` with the appropriate type for its
// field. The store's setters handle the writeback to `dataset.value.settings[i]`
// and ignore writes for fields that don't exist on the active chart's config
// (e.g. `setScale` on a pie config).
const handlers: Record<string, (val: unknown) => void> = {
  sort: (val) => setSort(val as Parameters<typeof setSort>[0]),
  scale: (val) => setScale(val as Parameters<typeof setScale>[0]),
  showLabels: (val) => setShowLabels(val as boolean),
  autoRotate: (val) => setAutoRotate(val as boolean),
  swap: (val) => setSwap(val as string | undefined),
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
      <SelectionTabs
        v-if="showChartTypeSelection"
        :model-value="chartType"
        :options="chartOptions"
        @update:model-value="onChartTypeChange"
      />
      <Separator v-if="showChartTypeSelection" />

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
