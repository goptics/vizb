<script setup lang="ts">
import { computed } from 'vue'
import { Monitor } from 'lucide-vue-next'
import type { HistoryEntry } from '../types'
import HistoryPopover from './HistoryPopover.vue'
import { useSortedHistory } from '../composables/useSortedHistory'

const props = defineProps<{
  os?: string
  history?: HistoryEntry[]
}>()

const historyRef = computed(() => props.history)
const { sortedHistory, hasHistory } = useSortedHistory(historyRef, (e) => !!e.os)
</script>

<template>
  <HistoryPopover
    v-if="os"
    :icon="Monitor"
    label="OS"
    :value="os!"
    history-title="OS History"
    :entries="sortedHistory"
    :has-history="hasHistory"
  >
    <template #entry="{ entry }">
      <span class="shrink-0 tabular-nums">{{ entry.os }}</span>
    </template>
  </HistoryPopover>
</template>
