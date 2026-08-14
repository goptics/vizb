<script setup lang="ts">
import { computed } from 'vue'
import { CalendarSync, Cpu, Monitor } from 'lucide-vue-next'
import type { Dataset, HistoryEntry } from '../types'
import { CPUtoString } from '../lib/utils'
import GroupSelector from './Selector.vue'
import MetaHistoryBadge from './MetaHistoryBadge.vue'

const props = defineProps<{
  dataset: Dataset
  datasets: { name: string }[]
  activeDatasetId: number
  resultGroups: { name: string }[]
  activeGroupId: number
}>()

const emit = defineEmits<{
  selectDataset: [id: number]
  selectGroup: [id: number]
}>()

const mainTitle = computed(() => props.datasets[0]?.name || 'Datasets')
const hasCPU = computed(() => props.dataset.meta?.cpu?.name || props.dataset.meta?.cpu?.cores)
const osLabel = computed(() => props.dataset.meta?.os ?? '')

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

const cpuHistoryFilter = (e: HistoryEntry) => !!(e.meta?.cpu?.name || e.meta?.cpu?.cores)
const osHistoryFilter = (e: HistoryEntry) => !!e.meta?.os
</script>

<template>
  <header class="space-y-3 py-5 text-center">
    <GroupSelector
      v-if="datasets.length > 1"
      :items="datasets"
      :activeId="activeDatasetId"
      @select="emit('selectDataset', $event)"
      class="mx-auto min-w-80"
      placeholder="Search Dataset..."
      notFoundText="No dataset found."
      :resultLimit="100"
    />

    <h1 v-else class="text-4xl font-bold">{{ mainTitle }}</h1>

    <div class="flex flex-col items-center gap-2">
      <MetaHistoryBadge
        v-if="hasCPU"
        :icon="Cpu"
        label="CPU"
        history-title="CPU History"
        :value="CPUtoString(dataset.meta?.cpu)"
        :history="dataset.history"
        :filter-fn="cpuHistoryFilter"
        content-width="w-80"
      >
        <template #entry="{ entry }">
          <span class="min-w-0 truncate text-right tabular-nums"
            >{{ CPUtoString(entry.meta?.cpu) }}</span
          >
        </template>
      </MetaHistoryBadge>
      <MetaHistoryBadge
        v-if="osLabel"
        :icon="Monitor"
        label="OS"
        history-title="OS History"
        :value="osLabel"
        :history="dataset.history"
        :filter-fn="osHistoryFilter"
      >
        <template #entry="{ entry }">
          <span class="shrink-0 tabular-nums">{{ entry.meta?.os }}</span>
        </template>
      </MetaHistoryBadge>
    </div>

    <MetaHistoryBadge
      v-if="dataset.timestamp"
      :icon="CalendarSync"
      label="Updated"
      history-title="Update History"
      :value="formatDate(dataset.timestamp)"
      :history="dataset.history"
    >
      <template #entry="{ entry }">
        <span class="shrink-0 tabular-nums">{{ formatDate(entry.timestamp) }}</span>
      </template>
    </MetaHistoryBadge>

    <p v-if="dataset.description" class="text-muted-foreground">
      {{ dataset.description }}
    </p>

    <GroupSelector
      v-if="resultGroups.length > 1"
      :items="resultGroups"
      :activeId="activeGroupId"
      @select="emit('selectGroup', $event)"
      placeholder="Search Group..."
      notFoundText="No group found."
      class="mx-auto min-w-80"
    />
  </header>
</template>
