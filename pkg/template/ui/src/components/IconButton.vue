<script setup lang="ts">
defineProps<{
  href?: string;
}>();
</script>

<template>
  <component
    :is="href ? 'a' : 'button'"
    :href="href"
    :target="href ? '_blank' : undefined"
    :rel="href ? 'noopener noreferrer' : undefined"
    :type="href ? undefined : 'button'"
    class="inline-flex items-center justify-center w-12 h-12 rounded-lg border border-border bg-card text-card-foreground hover:bg-accent hover:text-primary shadow-sm"
  >
    <slot />
  </component>
</template>
