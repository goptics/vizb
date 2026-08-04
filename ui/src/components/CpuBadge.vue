<script setup lang="ts">
import { computed } from 'vue'
import { Cpu } from 'lucide-vue-next'
import type { HistoryEntry } from '../types'
import HistoryPopover from './HistoryPopover.vue'
import { useSortedHistory } from '../composables/useSortedHistory'
import { CPUtoString } from '../lib/utils'

const props = defineProps<{
  cpu?: { name?: string; cores?: number }
  history?: HistoryEntry[]
}>()

const cpuString = computed(() => CPUtoString(props.cpu))
const historyRef = computed(() => props.history)
const { sortedHistory, hasHistory } = useSortedHistory(
  historyRef,
  (e) => !!(e.meta?.cpu?.name || e.meta?.cpu?.cores)
)
</script>

<template>
  <HistoryPopover
    v-if="cpu"
    :icon="Cpu"
    label="CPU"
    :value="cpuString"
    history-title="CPU History"
    :entries="sortedHistory"
    :has-history="hasHistory"
    content-width="w-80"
  >
    <template #entry="{ entry }">
      <span class="min-w-0 truncate text-right tabular-nums">{{
        CPUtoString(entry.meta?.cpu)
      }}</span>
    </template>
  </HistoryPopover>
</template>
