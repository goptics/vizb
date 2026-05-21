<script setup lang="ts">
import { computed } from 'vue'
import { CalendarSync } from 'lucide-vue-next'
import type { HistoryEntry } from '../types'
import Badge from './Badge.vue'
import Popover from './ui/Popover.vue'
import PopoverContent from './ui/PopoverContent.vue'
import PopoverTrigger from './ui/PopoverTrigger.vue'

const props = defineProps<{
  timestamp?: string
  history?: HistoryEntry[]
}>()

const formatDate = (ts: string) => {
  const date = new Date(ts)
  if (isNaN(date.getTime())) return ts
  return date.toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  })
}

const formattedDate = computed(() => {
  if (!props.timestamp) return ''
  return formatDate(props.timestamp)
})

const sortedHistory = computed(() => {
  if (!props.history?.length) return []
  return [...props.history].sort(
    (a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
  )
})

const hasHistory = computed(() => sortedHistory.value.length > 0)
</script>

<template>
  <Popover v-if="timestamp" class="flex justify-center">
    <PopoverTrigger class="cursor-pointer">
      <Badge :icon="CalendarSync" label="Updated" :value="formattedDate" />
    </PopoverTrigger>
    <PopoverContent v-if="hasHistory" class="w-72 p-3" align="center">
      <p class="mb-2 text-sm font-medium">Update History</p>
      <div class="space-y-1.5">
        <div
          v-for="entry in sortedHistory"
          :key="entry.tag"
          class="flex items-center justify-between gap-2 text-sm"
        >
          <span class="truncate text-muted-foreground">{{ entry.tag }}</span>
          <span class="shrink-0 tabular-nums">{{ formatDate(entry.timestamp) }}</span>
        </div>
      </div>
    </PopoverContent>
  </Popover>
</template>
