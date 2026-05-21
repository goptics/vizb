<script setup lang="ts">
import { ComboboxItemIndicator, type ComboboxItemIndicatorProps } from 'radix-vue'
import { computed, type HTMLAttributes } from 'vue'
import { cn } from '../../lib/utils'

const props = defineProps<ComboboxItemIndicatorProps & { class?: HTMLAttributes['class'] }>()

const delegatedProps = computed(() => {
  const { class: _, ...delegated } = props
  return delegated
})
</script>

<template>
  <ComboboxItemIndicator v-bind="delegatedProps" :class="cn('ml-auto', props.class)">
    <slot />
  </ComboboxItemIndicator>
</template>
