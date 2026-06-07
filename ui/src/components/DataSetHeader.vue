<script setup lang="ts">
import { computed } from 'vue'
import type { DataSet } from '../types'
import CpuBadge from './CpuBadge.vue'
import OsBadge from './OsBadge.vue'
import TimestampBadge from './TimestampBadge.vue'
import GroupSelector from './Selector.vue'

const props = defineProps<{
  benchmark: DataSet
  benchmarks: { name: string }[]
  activeDataSetId: number
  resultGroups: { name: string }[]
  activeGroupId: number
}>()

const emit = defineEmits<{
  selectDataSet: [id: number]
  selectGroup: [id: number]
}>()

const mainTitle = computed(() => props.benchmarks[0]?.name || 'DataSets')
const hasCPU = computed(() => props.benchmark.cpu?.name || props.benchmark.cpu?.cores)
const hasOS = computed(() => props.benchmark.os)
</script>

<template>
  <header class="space-y-3 py-5 text-center">
    <GroupSelector
      v-if="benchmarks.length > 1"
      :items="benchmarks"
      :activeId="activeDataSetId"
      @select="emit('selectDataSet', $event)"
      class="mx-auto min-w-80"
      placeholder="Search DataSet..."
      notFoundText="No benchmark found."
    />

    <h1 v-else class="text-4xl font-bold">{{ mainTitle }}</h1>

    <div class="flex flex-col items-center gap-2">
      <CpuBadge v-if="hasCPU" :cpu="benchmark.cpu" :history="benchmark.history" />
      <OsBadge v-if="hasOS" :os="benchmark.os" :history="benchmark.history" />
    </div>

    <TimestampBadge
      v-if="benchmark.timestamp"
      :timestamp="benchmark.timestamp"
      :history="benchmark.history"
    />

    <p v-if="benchmark.description" class="text-muted-foreground">
      {{ benchmark.description }}
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
