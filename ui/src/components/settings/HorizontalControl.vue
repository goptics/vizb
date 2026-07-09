<script setup lang="ts">
import { computed } from 'vue'
import { Separator } from '../ui'
import SettingsToggle from '../SettingsToggle.vue'

const props = defineProps<{
  modelValue: boolean | undefined
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const value = computed(() => props.modelValue ?? false)
</script>

<template>
  <SettingsToggle
    id="horizontal-bars-switch"
    label="Horizontal bars"
    description="Swap axes so bars grow rightward and categories appear on the Y axis."
    :checked="value"
    @update:checked="emit('update:modelValue', $event)"
  />
  <Separator />
</template>
