<script setup lang="ts">
import { ref } from 'vue'
import { Settings, ArrowUp, ArrowDown, ArrowUpDown } from 'lucide-vue-next'
import type { SortOrder } from '../types/benchmark'

const props = defineProps<{
  sortOrder: SortOrder
  showLabels: boolean
}>()

const emit = defineEmits<{
  'update:sortOrder': [value: SortOrder]
  'update:showLabels': [value: boolean]
}>()

const isOpen = ref(false)

const handleSortChange = (order: SortOrder) => {
  emit('update:sortOrder', order)
}

const handleLabelsToggle = () => {
  emit('update:showLabels', !props.showLabels)
}
</script>

<template>
  <div class="relative">
    <!-- Settings Button -->
    <button
      @click="isOpen = !isOpen"
      class="inline-flex items-center justify-center w-12 h-12 rounded-lg border border-border bg-card text-card-foreground hover:bg-accent hover:text-accent-foreground transition-colors shadow-sm"
      aria-label="Open settings"
      aria-haspopup="true"
      :aria-expanded="isOpen"
    >
      <Settings class="w-5 h-5" />
    </button>

    <!-- Popover Content -->
    <Transition name="popover">
      <div
        v-if="isOpen"
        class="absolute top-full right-0 mt-2 w-80 bg-popover border border-border rounded-lg shadow-lg z-50 overflow-hidden"
        @click.stop
      >
        <div class="p-4">
          <!-- Header -->
          <div class="flex items-center justify-between mb-4 pb-3 border-b border-border">
            <h3 class="text-base font-semibold text-foreground">Chart Settings</h3>
            <button
              @click="isOpen = false"
              class="inline-flex items-center justify-center w-8 h-8 rounded-md hover:bg-accent transition-colors"
              aria-label="Close settings"
            >
              <svg width="15" height="15" viewBox="0 0 15 15" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M11.7816 4.03157C12.0062 3.80702 12.0062 3.44295 11.7816 3.2184C11.5571 2.99385 11.193 2.99385 10.9685 3.2184L7.50005 6.68682L4.03164 3.2184C3.80708 2.99385 3.44301 2.99385 3.21846 3.2184C2.99391 3.44295 2.99391 3.80702 3.21846 4.03157L6.68688 7.49999L3.21846 10.9684C2.99391 11.193 2.99391 11.557 3.21846 11.7816C3.44301 12.0061 3.80708 12.0061 4.03164 11.7816L7.50005 8.31316L10.9685 11.7816C11.193 12.0061 11.5571 12.0061 11.7816 11.7816C12.0062 11.557 12.0062 11.193 11.7816 10.9684L8.31322 7.49999L11.7816 4.03157Z" fill="currentColor"/>
              </svg>
            </button>
          </div>

          <!-- Sort Order Section -->
          <div class="mb-4">
            <label class="text-sm font-semibold text-foreground block mb-2">Sort Order</label>
            <div class="flex flex-col gap-2">
              <button
                @click="handleSortChange('default')"
                :class="[
                  'inline-flex items-center gap-2 h-9 px-4 text-sm font-medium rounded-md border transition-all',
                  sortOrder === 'default'
                    ? 'bg-primary text-primary-foreground border-primary shadow-sm'
                    : 'bg-background text-foreground border-border hover:bg-accent hover:text-accent-foreground'
                ]"
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
              >
                <ArrowDown class="w-4 h-4" />
                Descending
              </button>
            </div>
          </div>

          <!-- Show Labels Section -->
          <div class="flex items-center justify-between">
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
      </div>
    </Transition>

    <!-- Backdrop -->
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
