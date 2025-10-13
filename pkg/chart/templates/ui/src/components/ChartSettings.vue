<template>
  <Card class="w-full">
    <CardHeader>
      <CardTitle>Settings</CardTitle>
    </CardHeader>
    <CardContent class="space-y-6">
      <!-- Chart Type Section -->
      <div class="space-y-4">
        <div class="space-y-2">
          <Label>Chart Type</Label>
          <Tabs 
            :model-value="chartType" 
            @update:model-value="handleChartTypeChange"
          >
            <TabsList class="grid w-full grid-cols-3">
              <TabsTrigger value="bar">
                <BarChart3 class="h-4 w-4" />
                <span class="ml-2">Bar</span>
              </TabsTrigger>
              <TabsTrigger value="line">
                <TrendingUp class="h-4 w-4" />
                <span class="ml-2">Line</span>
              </TabsTrigger>
              <TabsTrigger value="pie">
                <PieChart class="h-4 w-4" />
                <span class="ml-2">Pie</span>
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
            <Label for="sorting-switch">Enable Sorting</Label>
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
      <div class="flex justify-between items-center">
        <div class="space-y-1">
          <Label for="labels-switch" class="flex items-center gap-2">
            <LayersIcon class="h-4 w-4" />
            Show Labels
          </Label>
          <p class="text-sm text-muted-foreground">Display data labels on chart elements.</p>
        </div>
        <Switch
          id="labels-switch"
          v-model:checked="showLabels"
          @update:checked="handleShowLabelsChange"
          class="ml-auto"
        />
      </div>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { SortAsc, SortDesc, LayersIcon, BarChart3, TrendingUp, PieChart } from 'lucide-vue-next'
import { Card, CardContent, CardHeader, CardTitle } from './ui'
import { Switch } from './ui'
import { Tabs, TabsList, TabsTrigger } from './ui'
import { Separator } from './ui'
import { Label } from './ui'
import type { SortOrder, ChartType } from '../types/benchmark'

const props = defineProps<{
  chartType: ChartType
  sortOrder: SortOrder
  showLabels: boolean
}>()

const emit = defineEmits<{
  'update:chartType': [value: ChartType]
  'update:sortOrder': [value: SortOrder]
  'update:showLabels': [value: boolean]
}>()

const chartType = ref(props.chartType)
const isSortingEnabled = ref(props.sortOrder !== '')
const sortDirection = ref<'asc' | 'desc'>(
  props.sortOrder === 'asc' || props.sortOrder === 'desc' ? props.sortOrder : 'asc'
)
const showLabels = ref(props.showLabels)

const handleChartTypeChange = (value: string | number) => {
  chartType.value = String(value) as ChartType
  emit('update:chartType', String(value) as ChartType)
}

const handleSortingToggle = (checked: boolean) => {
  isSortingEnabled.value = checked
  
  if (checked) {
    // Ensure we have a valid sort direction when enabling
    if (!sortDirection.value) {
      sortDirection.value = 'asc'
    }

    emit('update:sortOrder', sortDirection.value)
    return
  }
  
  emit('update:sortOrder', '')
}

const handleSortDirectionChange = (value: string | number) => {
  const stringValue = String(value)
  if (stringValue === 'asc' || stringValue === 'desc') {
    sortDirection.value = stringValue
    emit('update:sortOrder', stringValue as SortOrder)
  }
}

const handleShowLabelsChange = (checked: boolean) => {
  showLabels.value = checked
  emit('update:showLabels', checked)
}
</script>

<style scoped>
/* Responsive adjustments */
@media (max-width: 640px) {
  .flex.justify-between {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
  }
  
  .ml-auto {
    margin-left: 0;
    margin-top: 0.5rem;
  }
}
</style>
