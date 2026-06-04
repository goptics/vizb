<script setup lang="ts">
import { computed } from 'vue'
import { SortAsc, SortDesc, BarChart3, TrendingUp, PieChart } from 'lucide-vue-next'
import type { Component } from 'vue'
import { Card, CardContent, CardHeader, CardTitle } from './ui'
import { Separator } from './ui'
import SettingsToggle from './SettingsToggle.vue'
import SelectionTabs from './SelectionTabs.vue'
import AxisSwapper from './AxisSwapper.vue'
import type { ChartType, SortOrder } from '../types'
import { useSettingsStore } from '../composables/useSettingsStore'
import { useSyncedSetting } from '../composables/useSyncedSetting'

const { settings, setSort, setShowLabels, setChartType } = useSettingsStore()

const CHART_ICONS: Record<ChartType, Component> = {
  bar: BarChart3,
  line: TrendingUp,
  pie: PieChart,
}

const chartType = useSyncedSetting<ChartType>(
  () => settings.charts[settings.activeChartIndex] ?? 'bar',
  (val) => setChartType(val)
)

const isSortingEnabled = useSyncedSetting(
  () => settings.sort.enabled,
  (val: boolean) => setSort({ enabled: val, order: settings.sort.order })
)

const sortDirection = useSyncedSetting<SortOrder>(
  () => settings.sort.order,
  (val) => setSort({ enabled: settings.sort.enabled, order: val })
)

const showLabels = useSyncedSetting(
  () => settings.showLabels,
  (val: boolean) => setShowLabels(val)
)

const showChartTypeSelection = computed(() => settings.charts.length > 1)
const chartOptions = computed(() =>
  settings.charts.map((type) => ({
    value: type,
    label: type.charAt(0).toUpperCase() + type.slice(1),
    icon: CHART_ICONS[type] ?? BarChart3,
  }))
)

const sortDirectionOptions = [
  { value: 'asc', label: 'Ascending', icon: SortAsc },
  { value: 'desc', label: 'Descending', icon: SortDesc },
]
</script>

<template>
  <Card class="w-full">
    <CardHeader>
      <CardTitle class="text-lg">Settings</CardTitle>
    </CardHeader>
    <CardContent class="space-y-4">
      <SelectionTabs v-if="showChartTypeSelection" v-model="chartType" :options="chartOptions" />

      <Separator v-if="showChartTypeSelection" />

      <div class="space-y-3">
        <SettingsToggle
          id="sorting-switch"
          label="Enable sorting"
          description="Sort your data by the selected axis."
          :checked="isSortingEnabled"
          @update:checked="isSortingEnabled = $event"
        />

        <SelectionTabs
          v-if="isSortingEnabled"
          v-model="sortDirection"
          :options="sortDirectionOptions"
        />
      </div>

      <Separator />

      <SettingsToggle
        id="labels-switch"
        label="Show labels"
        description="Display data labels on chart elements."
        :checked="showLabels"
        @update:checked="showLabels = $event"
      />

      <slot name="scale" />

      <slot name="autoRotate" />

      <Separator />

      <AxisSwapper />
    </CardContent>
  </Card>
</template>
