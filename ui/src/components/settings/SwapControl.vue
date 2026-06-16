<script setup lang="ts">
import { computed } from 'vue'
import SettingHeader from '../SettingHeader.vue'
import { activeDataSet } from '../../composables/useDataPoint'
import { axisKeyConcat } from '../../lib/swap'

const props = defineProps<{
  modelValue: string | undefined
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

// Default to the active dataset's axis key concatenation (e.g. "xyn" for
// [{key:"x"},{key:"y"},{key:"name"}]) when no swap is set. The user can still
// type any permutation — the only constraint is validation downstream.
const defaultSwap = computed(() => axisKeyConcat(activeDataSet.value?.axes))

const value = computed(() => props.modelValue ?? defaultSwap.value)
</script>

<template>
  <div class="flex items-center justify-between">
    <SettingHeader id="swap-input" label="Swap axis" description="Swap the axis of your data." />
    <input
      id="swap-input"
      type="text"
      :value="value"
      class="w-28 rounded-md border border-border bg-background px-2 py-1 text-sm text-card-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring"
      :placeholder="defaultSwap || 'yxn'"
      @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
  </div>
</template>
