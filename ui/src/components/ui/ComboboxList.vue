<script setup lang="ts">
import { ComboboxContent, type ComboboxContentProps } from 'radix-vue'
import { computed, type HTMLAttributes } from 'vue'
import { cn } from '../../lib/utils'

const props = defineProps<ComboboxContentProps & { class?: HTMLAttributes['class'] }>()

const delegatedProps = computed(() => {
  const { class: _, ...delegated } = props
  return delegated
})
</script>

<template>
  <ComboboxContent
    v-bind="delegatedProps"
    :class="
      cn(
        'data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 absolute left-0 top-full z-50 mt-1 max-h-[300px] w-full overflow-y-auto rounded-md border bg-popover text-popover-foreground shadow-lg',
        props.class
      )
    "
    :side-offset="4"
    :align-offset="0"
  >
    <div class="p-1">
      <slot />
    </div>
  </ComboboxContent>
</template>
