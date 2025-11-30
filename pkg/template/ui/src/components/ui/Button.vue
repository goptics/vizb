<script setup lang="ts">
import { cn } from '../../lib/utils'

const props = defineProps<{
  class?: string
  type?: 'button' | 'submit' | 'reset'
}>()
</script>

<template>
  <button
    v-bind="$attrs"
    :type="props.type || 'button'"
    :class="
      cn(
        'inline-flex h-10 items-center justify-center whitespace-nowrap rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground ring-offset-background hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50',
        props.class
      )
    "
  >
    <slot />
  </button>
</template>
