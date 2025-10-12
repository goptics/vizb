<template>
  <div class="space-y-2">
    <Label>Chart Type</Label>
    <ToggleGroup
      type="single"
      :model-value="modelValue"
      @update:model-value="(value) => $emit('update:modelValue', value as ChartType)"
      class="justify-start w-full"
    >
      <ToggleGroupItem 
        v-for="type in chartTypes" 
        :key="type.value"
        :value="type.value" 
        :aria-label="`Switch to ${type.label}`"
        class="flex-1"
      >
        <component :is="type.icon" class="h-4 w-4" />
        <span class="ml-2">{{ type.label }}</span>
      </ToggleGroupItem>
    </ToggleGroup>
  </div>
</template>

<script setup lang="ts">
import { BarChart3, LineChart, PieChart } from 'lucide-vue-next'
import { ToggleGroup, ToggleGroupItem } from './ui'
import { Label } from './ui'
import type { ChartType } from '../types/benchmark'

interface ChartTypeOption {
  value: ChartType
  label: string
  icon: any
}

const chartTypes: ChartTypeOption[] = [
  { value: 'bar', label: 'Bar', icon: BarChart3 },
  { value: 'line', label: 'Line', icon: LineChart },
  { value: 'pie', label: 'Pie', icon: PieChart },
]

defineProps<{
  modelValue: ChartType
}>()

defineEmits<{
  'update:modelValue': [value: ChartType]
}>()
</script>
