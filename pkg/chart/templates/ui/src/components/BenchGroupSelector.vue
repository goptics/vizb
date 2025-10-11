<script setup lang="ts">
import { ref } from 'vue'
import { List } from 'lucide-vue-next'
import type { Benchmark } from '../types/benchmark'

defineProps<{
  benchmarks: Benchmark[]
  activeBenchmarkId: number
}>()

const emit = defineEmits<{
  select: [id: number]
}>()

const isOpen = ref(false)

const handleSelect = (id: number) => {
  emit('select', id)
  isOpen.value = false
}
</script>

<template>
  <div class="relative" v-if="benchmarks.length > 1">
    <button
      @click="isOpen = !isOpen"
      class="inline-flex items-center gap-2 h-10 px-4 text-sm font-semibold rounded-lg border border-border bg-card text-card-foreground hover:bg-accent hover:text-accent-foreground transition-colors shadow-sm"
      aria-label="Select benchmark group"
      aria-haspopup="true"
      :aria-expanded="isOpen"
    >
      <List class="w-5 h-5" />
      <span>Bench Groups</span>
    </button>

    <!-- Popover Content -->
    <Transition name="popover">
      <div
        v-if="isOpen"
        class="absolute top-full left-0 mt-2 w-64 bg-popover border border-border rounded-lg shadow-lg z-50 overflow-hidden"
        @click.stop
      >
        <div class="max-h-80 overflow-y-auto p-2">
          <button
            v-for="(benchmark, index) in benchmarks"
            :key="index"
            @click="handleSelect(index)"
            :class="[
              'w-full px-3 py-2 text-sm font-medium text-left rounded-md transition-colors',
              activeBenchmarkId === index
                ? 'bg-primary text-primary-foreground'
                : 'text-popover-foreground hover:bg-accent hover:text-accent-foreground'
            ]"
            :aria-selected="activeBenchmarkId === index"
          >
            {{ benchmark.name }}
          </button>
        </div>
      </div>
    </Transition>

    <!-- Backdrop to close popover -->
    <div
      v-if="isOpen"
      class="fixed inset-0 z-40"
      @click="isOpen = false"
      aria-hidden="true"
    />
  </div>
</template>

<style scoped>
.popover-enter-active,
.popover-leave-active {
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.popover-enter-from {
  opacity: 0;
  transform: translateY(-8px);
}

.popover-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
