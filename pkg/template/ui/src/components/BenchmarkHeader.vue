<script setup lang="ts">
import type { Benchmark } from "../types/benchmark";

defineProps<{
  benchmark: Benchmark
  mainTitle: string
  hideTitle?: boolean
}>()
</script>

<template>
  <header class="text-center" :class="{ 'mb-8': !hideTitle }">
    <template v-if="!hideTitle">
      <h1 class="text-4xl font-bold tracking-tight mb-2">
        {{ mainTitle }}
      </h1>
    </template>
    
    <div class="flex items-center justify-center gap-4 mb-2">
      <span
        v-if="benchmark.cpu && benchmark.cpu.name && benchmark.cpu.cores"
        class="inline-flex items-center px-3 py-1 text-sm font-semibold rounded-lg border border-border bg-secondary text-secondary-foreground"
      >
        CPU: {{ benchmark.cpu.name }} ({{ benchmark.cpu.cores }} cores)
      </span>
      <span
        v-else-if="benchmark.cpu && benchmark.cpu.name"
        class="inline-flex items-center px-3 py-1 text-sm font-semibold rounded-lg border border-border bg-secondary text-secondary-foreground"
      >
        CPU: {{ benchmark.cpu.name }}
      </span>
      <span
        v-else-if="benchmark.cpu && benchmark.cpu.cores"
        class="inline-flex items-center px-3 py-1 text-sm font-semibold rounded-lg border border-border bg-secondary text-secondary-foreground"
      >
        CPU: {{ benchmark.cpu.cores }} cores
      </span>
    </div>
    
    <p v-if="benchmark.description" class="text-base text-muted-foreground">
      {{ benchmark.description }}
    </p>
  </header>
</template>

<style scoped>
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
