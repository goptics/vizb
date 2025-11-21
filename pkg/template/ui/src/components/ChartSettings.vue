<template>
  <Card class="w-full">
    <CardHeader>
      <CardTitle>Settings</CardTitle>
    </CardHeader>
    <CardContent class="space-y-6">
      <!-- Chart Type Section -->
      <div v-if="charts.length > 1" class="space-y-4">
        <div class="space-y-2">
          <Label>Chart type</Label>
          <Tabs 
            :model-value="chartType" 
            @update:model-value="handleChartTypeChange"
          >
            <TabsList :class="['grid w-full', gridColsClass]">
              <TabsTrigger 
                v-for="type in charts" 
                :key="type"
                :value="type"
              >
                <component :is="getChartIcon(type)" class="h-4 w-4" />
                <span class="ml-2">{{ type.charAt(0).toUpperCase() + type.slice(1) }}</span>
              </TabsTrigger>
            </TabsList>
          </Tabs>
        </div>
      </div>

      <Separator />

      <!-- Sort Controls Section -->
      <div class="space-y-4">
        <div class="flex justify-between items-center">
          <div class="space-y-1">
            <Label for="sorting-switch">Enable sorting</Label>
            <p class="text-sm text-muted-foreground">Sort your data by the selected axis.</p>
          </div>
          <Switch
            id="sorting-switch"
            v-model:checked="isSortingEnabled"
            @update:checked="handleSortingToggle"
          />
        </div>
        
        <div v-if="isSortingEnabled" class="space-y-2">
          <Label>Sort Direction</Label>
          <Tabs 
            :model-value="sortDirection" 
            @update:model-value="handleSortDirectionChange"
          >
            <TabsList class="grid w-full grid-cols-2">
              <TabsTrigger value="asc">
                <SortAsc class="h-4 w-4" />
                <span class="ml-2">Ascending</span>
              </TabsTrigger>
              <TabsTrigger value="desc">
                <SortDesc class="h-4 w-4" />
                <span class="ml-2">Descending</span>
              </TabsTrigger>
            </TabsList>
          </Tabs>
        </div>
      </div>

      <Separator />

      <!-- Show Labels Section -->
      <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-2">
        <div class="space-y-1">
          <Label for="labels-switch" class="flex items-center gap-2">
            <LayersIcon class="h-4 w-4" />
            Show labels
          </Label>
          <p class="text-sm text-muted-foreground">Display data labels on chart elements.</p>
        </div>
        <Switch
          id="labels-switch"
          v-model:checked="showLabels"
          @update:checked="handleShowLabelsChange"
          class="sm:ml-auto"
        />
      </div>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { SortAsc, SortDesc, LayersIcon, BarChart3, TrendingUp, PieChart } from 'lucide-vue-next'
import { Card, CardContent, CardHeader, CardTitle } from './ui'
import { Switch } from './ui'
import { Tabs, TabsList, TabsTrigger } from './ui'
import { Separator } from './ui'
import { Label } from './ui'
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
    case 'bar': return BarChart3
    case 'line': return TrendingUp
    case 'pie': return PieChart
    default: return BarChart3
  }
}

const gridColsClass = computed(() => {
  const len = charts.value.length
  if (len <= 1) return 'grid-cols-1'
  if (len === 2) return 'grid-cols-2'
  if (len === 3) return 'grid-cols-3'
  return 'grid-cols-4'
})
</script>

