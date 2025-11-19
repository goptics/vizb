<script setup lang="ts">
import { ComboboxTrigger, type ComboboxTriggerProps } from 'radix-vue'
import { computed, type HTMLAttributes } from 'vue'
import { cn } from '../../lib/utils'

const props = defineProps<ComboboxTriggerProps & { class?: HTMLAttributes['class'] }>()

const delegatedProps = computed(() => {
  const { class: _, ...delegated } = props
  return delegated
})
</script>

<template>
  <ComboboxTrigger
    v-bind="delegatedProps"
    :class="cn('', props.class)"
  >
    <slot />
  </ComboboxTrigger>
</template>
