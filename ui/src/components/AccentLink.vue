<script setup lang="ts">
defineProps<{
  href: string
}>()
</script>

<template>
  <a
    :href="href"
    target="_blank"
    rel="noopener noreferrer"
    class="font-semibold text-foreground hover:text-primary"
  >
    <slot />
  </a>
</template>
