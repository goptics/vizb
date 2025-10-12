<script setup lang="ts">
import { DialogTrigger } from 'radix-vue'
import { cn } from '../../lib/utils'

const props = defineProps<{ class?: string }>()
</script>

<template>
  <DialogTrigger v-bind="$attrs" :class="cn(props.class)">
    <slot />
  </DialogTrigger>
</template>
