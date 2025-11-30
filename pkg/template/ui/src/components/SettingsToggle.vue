<script setup lang="ts">
import { Label, Switch } from './ui'

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
    <div class="space-y-1">
      <Label :for="id" class="flex items-center gap-2">
        {{ label }}
      </Label>
      <p class="text-sm text-muted-foreground">{{ description }}</p>
    </div>
    <Switch :id="id" :checked="checked" @update:checked="$emit('update:checked', $event)" />
  </div>
</template>
