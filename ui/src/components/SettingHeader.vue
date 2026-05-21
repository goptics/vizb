<script setup lang="ts">
import { Label } from './ui'

defineProps<{
  label: string
  description?: string
  id?: string
}>()
</script>

<template>
  <div class="space-y-1">
    <Label :for="id" class="flex items-center gap-2">
      {{ label }}
    </Label>
    <p v-if="description" class="text-sm text-muted-foreground">{{ description }}</p>
  </div>
</template>
