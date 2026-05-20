<script setup lang="ts">
import type { Component } from 'vue'

defineProps<{
  icon?: Component
  label: string
  value: string
}>()
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 rounded-lg border border-border bg-secondary px-2 py-1 text-sm font-semibold text-secondary-foreground"
  >
    <component :is="icon" v-if="icon" class="h-3.5 w-3.5" />
    <span class="font-normal text-muted-foreground">{{ label }}:</span>
    {{ value }}
  </span>
</template>
