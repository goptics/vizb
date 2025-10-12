<script setup lang="ts">
import { ComboboxGroup, type ComboboxGroupProps } from 'radix-vue'
import { computed, type HTMLAttributes } from 'vue'
import { cn } from '../../lib/utils'

const props = defineProps<ComboboxGroupProps & { class?: HTMLAttributes['class'] }>()

const delegatedProps = computed(() => {
  const { class: _, ...delegated } = props
  return delegated
})
</script>

<template>
  <ComboboxGroup
    v-bind="delegatedProps"
    :class="cn('overflow-hidden p-1 text-foreground', props.class)"
  >
    <slot />
  </ComboboxGroup>
</template>
