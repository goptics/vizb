<script setup lang="ts">
import { computed } from 'vue'
import { Separator } from '../ui'
import SettingsToggle from '../SettingsToggle.vue'

const props = withDefaults(
  defineProps<{
    modelValue: boolean | undefined
    id: string
    label: string
    description: string
    separator?: boolean
  }>(),
  { separator: false }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const value = computed(() => props.modelValue ?? false)
</script>

<template>
  <SettingsToggle
    :id="id"
    :label="label"
    :description="description"
    :checked="value"
    @update:checked="emit('update:modelValue', $event)"
  />
  <Separator v-if="separator" />
</template>
