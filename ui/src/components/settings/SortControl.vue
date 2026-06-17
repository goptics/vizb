<script setup lang="ts">
import { computed } from 'vue'
import { SortAsc, SortDesc } from 'lucide-vue-next'
import { Separator } from '../ui'
import type { Sort, SortOrder } from '../../types'
import SettingsToggle from '../SettingsToggle.vue'
import SelectionTabs from '../SelectionTabs.vue'

const props = defineProps<{
  modelValue: Sort | undefined
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: Sort): void
}>()

const enabled = computed(() => props.modelValue?.enabled ?? false)
const order = computed<SortOrder>(() => props.modelValue?.order ?? 'asc')

const sortDirectionOptions = [
  { value: 'asc', label: 'Ascending', icon: SortAsc },
  { value: 'desc', label: 'Descending', icon: SortDesc },
]

const setEnabled = (val: boolean) => emit('update:modelValue', { enabled: val, order: order.value })

const setOrder = (val: string | number) =>
  emit('update:modelValue', { enabled: enabled.value, order: val as SortOrder })
</script>

<template>
  <div class="space-y-3">
    <SettingsToggle
      id="sorting-switch"
      label="Enable sorting"
      description="Sort your data by the selected axis."
      :checked="enabled"
      @update:checked="setEnabled"
    />
    <SelectionTabs
      v-if="enabled"
      :model-value="order"
      :options="sortDirectionOptions"
      @update:model-value="setOrder"
    />
  </div>
  <Separator />
</template>
