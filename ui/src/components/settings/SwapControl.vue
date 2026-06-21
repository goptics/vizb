<script setup lang="ts">
import { computed } from 'vue'
import SettingHeader from '../SettingHeader.vue'
import Selector from '../Selector.vue'
import { useDataPoint } from '@/composables/useDataPoint'
import { swapOptionKeys } from '@/lib/swap'

const { activeDataSet, activeArrangement, isValueMode } = useDataPoint()

const swapOptions = computed(() =>
  swapOptionKeys(activeDataSet.value?.data, isValueMode.value).map((key) => ({ name: key }))
)

const props = defineProps<{
  modelValue: string | undefined
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

// Index of the current arrangement in swapOptions. Falls back to the active
// arrangement's targetString (identity when nothing is set) when the wire-format
// swap doesn't match any option (e.g., stale value or 1D data with bare 'n').
const selectedSwapIndex = computed(() => {
  const target = props.modelValue ?? activeArrangement.value.targetString
  const idx = swapOptions.value.findIndex((opt) => opt.name === target)
  return idx !== -1 ? idx : 0
})

const handleSelect = (index: number) => {
  const opt = swapOptions.value[index]
  if (!opt) return
  emit('update:modelValue', opt.name)
}
</script>

<template>
  <div class="flex items-center justify-between">
    <SettingHeader label="Swap axis" description="Swap the axis of your data." />
    <Selector
      :items="swapOptions"
      :activeId="selectedSwapIndex"
      @select="handleSelect"
      class="w-28"
    />
  </div>
</template>
