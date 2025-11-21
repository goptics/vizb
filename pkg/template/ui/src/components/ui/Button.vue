<script setup lang="ts">
import { cn } from "../../lib/utils";

const props = defineProps<{
  class?: string;
  type?: "button" | "submit" | "reset";
}>();
</script>

<template>
  <button
    v-bind="$attrs"
    :type="props.type || 'button'"
    :class="
      cn(
        'inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background  focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground hover:bg-primary/90 h-10 px-4 py-2',
        props.class
      )
    "
  >
    <slot />
  </button>
</template>
