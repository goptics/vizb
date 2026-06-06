<script setup lang="ts">
import { computed } from 'vue'
import { CalendarSync } from 'lucide-vue-next'
import type { HistoryEntry } from '../types'
import HistoryPopover from './HistoryPopover.vue'
import { useSortedHistory } from '../composables/useSortedHistory'

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

const formattedDate = computed(() => (props.timestamp ? formatDate(props.timestamp) : ''))
const historyRef = computed(() => props.history)
const { sortedHistory, hasHistory } = useSortedHistory(historyRef)
</script>

<template>
  <HistoryPopover
    v-if="timestamp"
    :icon="CalendarSync"
    label="Updated"
    :value="formattedDate"
    history-title="Update History"
    :entries="sortedHistory"
    :has-history="hasHistory"
  >
    <template #entry="{ entry }">
      <span class="shrink-0 tabular-nums">{{ formatDate(entry.timestamp) }}</span>
    </template>
  </HistoryPopover>
</template>
