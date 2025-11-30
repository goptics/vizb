<script setup lang="ts">
import { ComboboxRoot } from 'radix-vue'
import { cn } from '../../lib/utils'

const props = defineProps<{
  class?: string
}>()
</script>

<template>
  <ComboboxRoot v-bind="$attrs" :class="cn(props.class)">
    <slot />
  </ComboboxRoot>
</template>
