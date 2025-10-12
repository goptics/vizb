<script setup lang="ts">
import { ComboboxAnchor, type ComboboxAnchorProps } from 'radix-vue'
import { computed, type HTMLAttributes } from 'vue'
import { cn } from '../../lib/utils'

const props = defineProps<ComboboxAnchorProps & { class?: HTMLAttributes['class'] }>()

const delegatedProps = computed(() => {
  const { class: _, ...delegated } = props
  return delegated
})
</script>

<template>
  <ComboboxAnchor
    v-bind="delegatedProps"
    :class="cn('relative', props.class)"
  >
    <slot />
  </ComboboxAnchor>
</template>
