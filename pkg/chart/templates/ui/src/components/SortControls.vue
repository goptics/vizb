<script setup lang="ts">
import { ArrowUp, ArrowDown, ArrowUpDown } from 'lucide-vue-next'
import type { SortOrder } from '../types/benchmark'

const props = defineProps<{
  sortOrder: SortOrder
  showLabels: boolean
}>()

const emit = defineEmits<{
  'update:sortOrder': [value: SortOrder]
  'update:showLabels': [value: boolean]
}>()

const handleSortChange = (order: SortOrder) => {
  emit('update:sortOrder', order)
}

const handleLabelsToggle = () => {
  emit('update:showLabels', !props.showLabels)
}
</script>

<template>
  <div class="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between mb-6">
    <!-- Sort Buttons -->
    <div class="flex flex-col gap-2 w-full sm:w-auto">
      <label class="text-sm font-semibold text-foreground">Sort Order</label>
      <div class="inline-flex gap-1">
        <button
          @click="handleSortChange('default')"
          :class="[
            'inline-flex items-center gap-2 h-9 px-4 text-sm font-medium rounded-md border transition-all',
            sortOrder === 'default'
              ? 'bg-primary text-primary-foreground border-primary shadow-sm'
              : 'bg-background text-foreground border-border hover:bg-accent hover:text-accent-foreground'
          ]"
          aria-label="Default sort"
        >
          <ArrowUpDown class="w-4 h-4" />
          Default
        </button>
        <button
          @click="handleSortChange('asc')"
          :class="[
            'inline-flex items-center gap-2 h-9 px-4 text-sm font-medium rounded-md border transition-all',
            sortOrder === 'asc'
              ? 'bg-primary text-primary-foreground border-primary shadow-sm'
              : 'bg-background text-foreground border-border hover:bg-accent hover:text-accent-foreground'
          ]"
          aria-label="Sort ascending"
        >
          <ArrowUp class="w-4 h-4" />
          Ascending
        </button>
        <button
          @click="handleSortChange('desc')"
          :class="[
            'inline-flex items-center gap-2 h-9 px-4 text-sm font-medium rounded-md border transition-all',
            sortOrder === 'desc'
              ? 'bg-primary text-primary-foreground border-primary shadow-sm'
              : 'bg-background text-foreground border-border hover:bg-accent hover:text-accent-foreground'
          ]"
          aria-label="Sort descending"
        >
          <ArrowDown class="w-4 h-4" />
          Descending
        </button>
      </div>
    </div>

    <!-- Show Labels Toggle -->
    <div class="flex items-center justify-between gap-3 w-full sm:w-auto">
      <label class="text-sm font-semibold text-foreground">Show Labels</label>
      <button
        @click="handleLabelsToggle"
        :class="[
          'relative w-11 h-6 rounded-full transition-colors',
          showLabels ? 'bg-primary' : 'bg-muted'
        ]"
        role="switch"
        :aria-checked="showLabels"
        aria-label="Toggle chart labels"
      >
        <span
          :class="[
            'absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full transition-transform shadow-sm',
            showLabels ? 'translate-x-5' : 'translate-x-0'
          ]"
        />
      </button>
    </div>
  </div>
</template>

<style scoped>
@media (max-width: 640px) {
  .inline-flex {
    flex-direction: column;
    width: 100%;
  }

  button {
    width: 100%;
    justify-content: flex-start;
  }
}
</style>
