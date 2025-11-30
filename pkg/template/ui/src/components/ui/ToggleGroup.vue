<script setup lang="ts">
import { ToggleGroupRoot } from 'radix-vue'
import { cn } from '../../lib/utils'

const props = defineProps<{
  class?: string
}>()
</script>

<template>
  <ToggleGroupRoot
    v-bind="$attrs"
    :class="
      cn(
        'flex items-center justify-center rounded-md bg-muted p-1 text-muted-foreground',
        props.class
      )
    "
  >
    <slot />
  </ToggleGroupRoot>
</template>
