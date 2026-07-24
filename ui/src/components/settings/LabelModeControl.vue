<script setup lang="ts">
import { computed } from 'vue'
import { EyeOff, Hash, Percent } from 'lucide-vue-next'
import { Separator } from '../ui'
import SelectionTabs from '../SelectionTabs.vue'
import type { LabelMode } from '@/types'

const props = defineProps<{ modelValue: LabelMode | undefined }>()
const emit = defineEmits<{ (e: 'update:modelValue', value: LabelMode): void }>()

const value = computed<LabelMode>(() => props.modelValue ?? 'none')
const options = [
  { value: 'none', label: 'Off', icon: EyeOff },
  { value: 'value', label: 'Value', icon: Hash },
  { value: 'percentage', label: 'Percentage', icon: Percent },
]
</script>

<template>
  <div class="space-y-3">
    <p class="text-sm font-medium">Labels</p>
    <SelectionTabs
      :model-value="value"
      :options="options"
      @update:model-value="emit('update:modelValue', $event as LabelMode)"
    />
  </div>
  <Separator />
</template>
