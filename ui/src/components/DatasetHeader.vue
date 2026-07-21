<script setup lang="ts">
import { computed } from 'vue'
import type { Dataset } from '../types'
import CpuBadge from './CpuBadge.vue'
import OsBadge from './OsBadge.vue'
import TimestampBadge from './TimestampBadge.vue'
import GroupSelector from './Selector.vue'

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
const hasOS = computed(() => props.dataset.meta?.os)
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
      <CpuBadge v-if="hasCPU" :cpu="dataset.meta?.cpu" :history="dataset.history" />
      <OsBadge v-if="hasOS" :os="dataset.meta?.os" :history="dataset.history" />
    </div>

    <TimestampBadge
      v-if="dataset.timestamp"
      :timestamp="dataset.timestamp"
      :history="dataset.history"
    />

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
