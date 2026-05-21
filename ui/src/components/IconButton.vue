<script setup lang="ts">
defineProps<{
  href?: string
}>()
</script>

<template>
  <component
    :is="href ? 'a' : 'button'"
    :href="href"
    :target="href ? '_blank' : undefined"
    :rel="href ? 'noopener noreferrer' : undefined"
    :type="href ? undefined : 'button'"
    class="inline-flex h-12 w-12 items-center justify-center rounded-lg border border-border bg-card text-card-foreground shadow-sm hover:bg-accent hover:text-primary"
  >
    <slot />
  </component>
</template>
