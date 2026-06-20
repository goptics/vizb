<script setup lang="ts">
import type { Component } from 'vue'
import { type ClassValue } from 'clsx'
import { cn } from '@/lib/utils'

const props = defineProps<{
  icon?: Component
  label: string
  active?: boolean
  class?: ClassValue
  title?: string
}>()
</script>

<template>
  <button
    type="button"
    :title="props.title"
    :aria-pressed="props.active"
    :class="
      cn(
        'inline-flex items-center gap-1.5 rounded-lg border border-border px-2 py-1 text-sm font-semibold transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
        props.active
          ? 'bg-accent text-primary'
          : 'bg-secondary text-secondary-foreground hover:bg-accent hover:text-primary',
        props.class
      )
    "
  >
    <component :is="icon" v-if="icon" class="h-3.5 w-3.5 shrink-0" />
    {{ label }}
  </button>
</template>
