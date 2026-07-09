<script setup lang="ts">
import { computed } from 'vue'
import { BarChart3, TrendingUp } from 'lucide-vue-next'
import { Separator } from '../ui'
import SelectionTabs from '../SelectionTabs.vue'
import type { ScaleType } from '@/types'

const props = defineProps<{
  modelValue: ScaleType | undefined
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: ScaleType): void
}>()

const value = computed<ScaleType>(() => props.modelValue ?? 'linear')

const scaleOptions = [
  { value: 'linear', label: 'Linear', icon: BarChart3 },
  { value: 'log', label: 'Logarithmic', icon: TrendingUp },
]

const onUpdate = (val: string | number) => {
  if (props.disabled) return
  emit('update:modelValue', val as ScaleType)
}
</script>

<template>
  <div class="space-y-3" :class="{ 'opacity-60': props.disabled }">
    <p class="text-sm font-medium">Data Scale</p>
    <SelectionTabs
      :model-value="value"
      :options="scaleOptions"
      :disabled="props.disabled"
      @update:model-value="onUpdate"
    />
  </div>
  <Separator />
</template>
