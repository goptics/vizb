<script setup lang="ts">
import { computed } from 'vue'
import { Monitor } from 'lucide-vue-next'
import type { HistoryEntry } from '../types'
import Badge from './Badge.vue'
import Popover from './ui/Popover.vue'
import PopoverContent from './ui/PopoverContent.vue'
import PopoverTrigger from './ui/PopoverTrigger.vue'

const props = defineProps<{
  os?: string
  history?: HistoryEntry[]
}>()

const sortedHistory = computed(() => {
  if (!props.history?.length) return []
  return [...props.history]
    .filter((e) => e.os)
    .sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
})

const hasHistory = computed(() => sortedHistory.value.length > 0)
</script>

<template>
  <Popover v-if="os" class="flex justify-center">
    <PopoverTrigger class="cursor-pointer">
      <Badge :icon="Monitor" label="OS" :value="os!" />
    </PopoverTrigger>
    <PopoverContent v-if="hasHistory" class="w-72 p-3" align="center">
      <p class="mb-2 text-sm font-medium">OS History</p>
      <div class="space-y-1.5">
        <div
          v-for="entry in sortedHistory"
          :key="entry.tag"
          class="flex items-center justify-between gap-2 text-sm"
        >
          <span class="truncate text-muted-foreground">{{ entry.tag }}</span>
          <span class="shrink-0 tabular-nums">{{ entry.os }}</span>
        </div>
      </div>
    </PopoverContent>
  </Popover>
</template>
