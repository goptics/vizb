<script setup lang="ts">
import { DialogContent } from 'radix-vue'
import { cn } from '../../lib/utils'

const props = defineProps<{
  class?: string
  side?: 'top' | 'right' | 'bottom' | 'left'
}>()
</script>

<template>
  <DialogContent
    v-bind="$attrs"
    :class="cn(
      'z-50 bg-transparent p-0',
      props.class
    )"
  >
    <slot />
  </DialogContent>
  
</template>
