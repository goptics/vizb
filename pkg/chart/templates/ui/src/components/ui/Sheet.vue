<script setup lang="ts">
import { DialogRoot } from 'radix-vue'

defineProps<{ class?: string }>()
</script>

<template>
  <DialogRoot v-bind="$attrs">
    <slot />
  </DialogRoot>
</template>
