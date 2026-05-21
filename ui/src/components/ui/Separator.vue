<script setup lang="ts">
import { cn } from '../../lib/utils'

const props = defineProps<{
  class?: string
  orientation?: 'horizontal' | 'vertical'
}>()
</script>

<template>
  <div
    v-bind="$attrs"
    :class="
      cn(
        'shrink-0 bg-border',
        props.orientation === 'vertical' ? 'h-full w-px' : 'h-px w-full',
        props.class
      )
    "
    role="separator"
    :aria-orientation="props.orientation || 'horizontal'"
  ></div>
</template>
