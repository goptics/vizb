<script setup lang="ts">
import { ToggleGroupItem } from 'radix-vue'
import { cn } from '../../lib/utils'

const props = defineProps<{
  class?: string
  value: string
}>()
</script>

<template>
  <ToggleGroupItem
    v-bind="$attrs"
    :value="props.value"
    :class="
      cn(
        'inline-flex items-center justify-center rounded-sm px-3 py-1.5 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 data-[state=on]:bg-background data-[state=on]:text-foreground data-[state=on]:shadow-sm',
        props.class
      )
    "
  >
    <slot />
  </ToggleGroupItem>
</template>
