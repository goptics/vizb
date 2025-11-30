<script setup lang="ts">
import { ComboboxInput, type ComboboxInputProps } from 'radix-vue'
import { computed, type HTMLAttributes } from 'vue'
import { cn } from '../../lib/utils'

const props = defineProps<ComboboxInputProps & { class?: HTMLAttributes['class'] }>()

const delegatedProps = computed(() => {
  const { class: _, ...delegated } = props
  return delegated
})
</script>

<template>
  <ComboboxInput
    v-bind="delegatedProps"
    :class="
      cn(
        'flex h-10 w-full rounded-md bg-transparent py-3 text-sm outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50',
        props.class
      )
    "
  />
</template>
