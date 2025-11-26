<script setup lang="ts">
defineProps<{
  href: string;
}>();
</script>

<template>
  <a
    :href="href"
    target="_blank"
    rel="noopener noreferrer"
    class="text-foreground font-semibold hover:text-primary"
  >
    <slot />
  </a>
</template>
