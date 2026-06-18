<script setup lang="ts">
import { computed } from 'vue'
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
    id="three-d-rotate-switch"
    label="Auto rotate"
    description="Continuously rotate the 3D chart."
    :checked="value"
    @update:checked="emit('update:modelValue', $event)"
  />
</template>
