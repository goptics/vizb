<script setup lang="ts">
import { computed } from 'vue'
import type { DataSet } from '../types'
import CpuBadge from './CpuBadge.vue'
import OsBadge from './OsBadge.vue'
import TimestampBadge from './TimestampBadge.vue'
import GroupSelector from './Selector.vue'

const props = defineProps<{
  dataSet: DataSet
  dataSets: { name: string }[]
  activeDataSetId: number
  resultGroups: { name: string }[]
  activeGroupId: number
}>()

const emit = defineEmits<{
  selectDataSet: [id: number]
  selectGroup: [id: number]
}>()

const mainTitle = computed(() => props.dataSets[0]?.name || 'DataSets')
const hasCPU = computed(() => props.dataSet.meta?.cpu?.name || props.dataSet.meta?.cpu?.cores)
const hasOS = computed(() => props.dataSet.meta?.os)
</script>

<template>
  <header class="space-y-3 py-5 text-center">
    <GroupSelector
      v-if="dataSets.length > 1"
      :items="dataSets"
      :activeId="activeDataSetId"
      @select="emit('selectDataSet', $event)"
      class="mx-auto min-w-80"
      placeholder="Search DataSet..."
      notFoundText="No dataSet found."
    />

    <h1 v-else class="text-4xl font-bold">{{ mainTitle }}</h1>

    <div class="flex flex-col items-center gap-2">
      <CpuBadge v-if="hasCPU" :cpu="dataSet.meta?.cpu" :history="dataSet.history" />
      <OsBadge v-if="hasOS" :os="dataSet.meta?.os" :history="dataSet.history" />
    </div>

    <TimestampBadge
      v-if="dataSet.timestamp"
      :timestamp="dataSet.timestamp"
      :history="dataSet.history"
    />

    <p v-if="dataSet.description" class="text-muted-foreground">
      {{ dataSet.description }}
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
