<script setup lang="ts">
import { PopoverTrigger } from 'radix-vue'
import { cn } from '../../lib/utils'

const props = defineProps<{ class?: string }>()
</script>

<template>
  <PopoverTrigger v-bind="$attrs" :class="cn(props.class)">
    <slot />
  </PopoverTrigger>
</template>








