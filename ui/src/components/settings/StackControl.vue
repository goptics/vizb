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
    id="stack-switch"
    label="Stack series"
    description="Stack grouped bar or line series into category totals."
    :checked="value"
    @update:checked="emit('update:modelValue', $event)"
  />
  <Separator />
</template>
