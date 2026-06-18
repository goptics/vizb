<script setup lang="ts">
import { TabsRoot, type TabsRootEmits, type TabsRootProps, useForwardPropsEmits } from 'radix-vue'
import type { HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<TabsRootProps & { class?: HTMLAttributes['class'] }>()
const emits = defineEmits<TabsRootEmits>()

const forwarded = useForwardPropsEmits(props, emits)
</script>

<template>
  <TabsRoot v-bind="forwarded" :class="cn('', props.class)">
    <slot />
  </TabsRoot>
</template>
