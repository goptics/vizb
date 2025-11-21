<script setup lang="ts">
defineProps<{
  href: string;
  newTab?: boolean;
}>();
</script>

<template>
  <a
    :href="href"
    :target="newTab ? '_blank' : undefined"
    :rel="newTab ? 'noopener noreferrer' : undefined"
    class="text-foreground font-semibold hover:text-primary"
  >
    <slot />
  </a>
</template>
