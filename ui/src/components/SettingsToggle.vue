<script setup lang="ts">
import { Switch } from './ui'
import SettingHeader from './SettingHeader.vue'

defineProps<{
  id: string
  label: string
  description: string
  checked: boolean
}>()

defineEmits<{
  (e: 'update:checked', value: boolean): void
}>()
</script>

<template>
  <div class="flex items-center justify-between">
    <SettingHeader :id="id" :label="label" :description="description" />
    <Switch :id="id" :checked="checked" @update:checked="$emit('update:checked', $event)" />
  </div>
</template>
