<script setup lang="ts">
import { PopoverContent } from 'radix-vue'
import { cn } from '@/lib/utils'

const props = defineProps<{ class?: string; align?: 'start' | 'center' | 'end' }>()
</script>

<template>
  <PopoverContent
    v-bind="$attrs"
    :class="cn('z-50 w-[380px] rounded-lg border bg-background p-0 shadow-xl', props.class)"
    :side-offset="8"
    :align="props.align ?? 'end'"
  >
    <slot />
  </PopoverContent>
</template>
