<script setup lang="ts">
import { PopoverRoot } from 'radix-vue'

defineProps<{ class?: string }>()
</script>

<template>
  <PopoverRoot v-bind="$attrs">
    <slot />
  </PopoverRoot>
  
</template>


