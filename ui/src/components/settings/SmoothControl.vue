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
    id="smooth-lines-switch"
    label="Smooth lines"
    description="Render curved segments between line chart points."
    :checked="value"
    @update:checked="emit('update:modelValue', $event)"
  />
  <Separator />
</template>
