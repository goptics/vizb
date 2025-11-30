<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { SortAsc, SortDesc, BarChart3, TrendingUp, PieChart } from 'lucide-vue-next'
import { Card, CardContent, CardHeader, CardTitle } from './ui'
import { Separator } from './ui'
import SettingsToggle from './SettingsToggle.vue'
import SelectionTabs from './SelectionTabs.vue'
import type { ChartType, SortOrder } from '../types/benchmark'
import { useSettingsStore } from '../composables/useSettingsStore'

const {
  sortOrder,
  showLabels: showLabelsStore,
  charts,
  chartType: chartTypeStore,
  setSort,
  setShowLabels,
  setChartType,
} = useSettingsStore()

const chartType = ref(chartTypeStore.value)
const isSortingEnabled = ref(sortOrder.value.enabled)
const sortDirection = ref<SortOrder>(sortOrder.value.order)
const showLabels = ref(showLabelsStore.value)

watch(chartType, (val) => setChartType(val))
watch(showLabels, (val) => setShowLabels(val))
watch([isSortingEnabled, sortDirection], ([enabled, order]) => setSort({ enabled, order }))

const showChartTypeSelection = computed(() => charts.value.length > 1)
const chartOptions = computed(() =>
  charts.value.map((type) => ({
    value: type,
    label: type.charAt(0).toUpperCase() + type.slice(1),
    icon: getChartIcon(type),
  }))
)

const sortDirectionOptions = [
  { value: 'asc', label: 'Ascending', icon: SortAsc },
  { value: 'desc', label: 'Descending', icon: SortDesc },
]

const handleChartTypeChange = (value: string | number) => {
  chartType.value = String(value) as ChartType
}

const handleSortingToggle = (checked: boolean) => {
  isSortingEnabled.value = checked
}

const handleSortDirectionChange = (value: string | number) => {
  sortDirection.value = String(value) as SortOrder
}

const handleShowLabelsChange = (checked: boolean) => {
  showLabels.value = checked
}

const getChartIcon = (type: ChartType) => {
  switch (type) {
    case 'bar':
      return BarChart3
    case 'line':
      return TrendingUp
    case 'pie':
      return PieChart
    default:
      return BarChart3
  }
}
</script>

<template>
  <Card class="w-full">
    <CardHeader>
      <CardTitle class="text-lg">Settings</CardTitle>
    </CardHeader>
    <CardContent class="space-y-4">
      <!-- Chart Type Section -->
      <SelectionTabs
        v-if="showChartTypeSelection"
        :model-value="chartType"
        :options="chartOptions"
        @update:model-value="handleChartTypeChange"
      />

      <Separator v-if="showChartTypeSelection" />

      <!-- Sort Controls Section -->
      <div class="space-y-3">
        <SettingsToggle
          id="sorting-switch"
          label="Enable sorting"
          description="Sort your data by the selected axis."
          :checked="isSortingEnabled"
          @update:checked="handleSortingToggle"
        />

        <SelectionTabs
          v-if="isSortingEnabled"
          :model-value="sortDirection"
          :options="sortDirectionOptions"
          @update:model-value="handleSortDirectionChange"
        />
      </div>

      <Separator />

      <!-- Show Labels Section -->
      <SettingsToggle
        id="labels-switch"
        label="Show labels"
        description="Display data labels on chart elements."
        :checked="showLabels"
        @update:checked="handleShowLabelsChange"
      />
    </CardContent>
  </Card>
</template>
