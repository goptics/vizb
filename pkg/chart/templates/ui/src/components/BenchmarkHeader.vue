<script setup lang="ts">
import type { Benchmark } from "../types/benchmark";

const props = defineProps<{
  benchmark: Benchmark;
  mainTitle: string;
}>();
</script>

<template>
  <header class="text-center mb-8 animate-fade-in">
    <h1 class="text-4xl font-bold tracking-tight mb-2">
      {{ mainTitle }}
      <span
        v-if="benchmark.cpu"
        class="inline-flex items-center px-3 py-1 text-sm font-semibold rounded-lg border border-border bg-secondary text-secondary-foreground ml-2"
      >
        CPU: {{ benchmark.cpu }}
      </span>
    </h1>
    <p v-if="benchmark.description" class="text-base text-muted-foreground">
      {{ benchmark.description }}
    </p>
  </header>
</template>

<style scoped>
@keyframes fade-in {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-fade-in {
  animation: fade-in 0.6s ease-out;
}

@media (max-width: 768px) {
  h1 {
    font-size: 2rem;
  }

  h1 span {
    display: block;
    margin: 0.75rem auto 0;
    width: fit-content;
  }
}
</style>
