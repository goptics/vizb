<script setup lang="ts">
import { ComboboxLabel, type ComboboxLabelProps } from 'radix-vue'
import { computed, type HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<ComboboxLabelProps & { class?: HTMLAttributes['class'] }>()

const delegatedProps = computed(() => {
  const { class: _, ...delegated } = props
  return delegated
})
</script>

<template>
  <ComboboxLabel
    v-bind="delegatedProps"
    :class="cn('px-2 py-1.5 text-xs font-semibold text-muted-foreground', props.class)"
  >
    <slot />
  </ComboboxLabel>
</template>
