<script setup lang="ts">
import { computed, type Component } from 'vue'
import type { HistoryEntry } from '../types'
import { useSortedHistory } from '../composables/useSortedHistory'
import HistoryPopover from './HistoryPopover.vue'

const props = defineProps<{
  icon: Component
  label: string
  historyTitle: string
  value: string
  history?: HistoryEntry[]
  filterFn?: (entry: HistoryEntry) => boolean
  contentWidth?: string
}>()

const historyRef = computed(() => props.history)
const { sortedHistory, hasHistory } = useSortedHistory(historyRef, (entry) =>
  props.filterFn ? props.filterFn(entry) : true
)
</script>

<template>
  <HistoryPopover
    v-if="value"
    :icon="icon"
    :label="label"
    :value="value"
    :history-title="historyTitle"
    :entries="sortedHistory"
    :has-history="hasHistory"
    :content-width="contentWidth"
  >
    <template #entry="slotProps">
      <slot name="entry" v-bind="slotProps" />
    </template>
  </HistoryPopover>
</template>
