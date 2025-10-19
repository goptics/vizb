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
      'fixed inset-y-0 right-0 z-50 h-full w-[400px] border-l bg-background p-6 shadow-lg transition ease-in-out data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:duration-300 data-[state=open]:duration-500',
      props.class
    )"
  >
    <slot />
  </DialogContent>
  
</template>
