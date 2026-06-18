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
    id="three-d-switch"
    label="3D view"
    description="Render as a 3D chart with metric height on the z axis."
    :checked="value"
    @update:checked="emit('update:modelValue', $event)"
  />
</template>
