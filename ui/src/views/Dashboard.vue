<script setup lang="ts">
import { computed, watch } from 'vue'
import { Moon, Sun, Package } from 'lucide-vue-next'
import type { Sort, ScaleType, Axis, AxisLabels } from '../types'
import { useDataPoint } from '../composables/useDataPoint'
import { useChartPipeline } from '../composables/useChartPipeline'
import { useSettingsStore } from '../composables/useSettingsStore'
import { useDashboardInit } from '../composables/useDashboardInit'
import { swapAxisLabels } from '../lib/swap'
import ChartSettingsPopover from '../components/ChartSettingsPopover.vue'
import ChartCard from '../components/ChartCard.vue'
import DataSetHeader from '../components/DataSetHeader.vue'
import LoadingSkeleton from '../components/LoadingSkeleton.vue'
import LoadError from '../components/LoadError.vue'
import AppFooter from '../components/AppFooter.vue'
import IconButton from '../components/IconButton.vue'

const version = window.VIZB_VERSION || 'v0.0.0-dev'

const {
  dataSets,
  activeDataSet,
  activeDataSetId,
  selectDataSet,
  activeArrangement,
  resultGroups,
  activeGroupId,
  selectGroup,
  setGroupNames,
  loading,
  loadError,
} = useDataPoint()

const { settings, resolved, toggleDark } = useSettingsStore()

// Build an AxisLabels object from axes[] for swapAxisLabels. Falls back to the
// legacy axisLabels field for datasets migrated from old JSON.
const axisLabelsFromAxes = (axes: Axis[] | undefined): AxisLabels | undefined => {
  if (!axes?.length) return undefined
  const result: AxisLabels = {}
  for (const a of axes) {
    if (a.label) (result as Record<string, string>)[a.key] = a.label
  }
  return Object.keys(result).length ? result : undefined
}

// The full raw rows — the worker owns grouping/projection, so we pass the dataset
// as-is (no main-thread grouping or swap mutation). Only a dataset switch re-clones.
const activeResults = computed(() => activeDataSet.value?.data || [])
// Display labels: prefer axes[] (new schema), fall back to legacy axisLabels field.
// swapAxisLabels permutes them to match the active arrangement.
const activeLabels = computed(() =>
  swapAxisLabels(
    activeArrangement.value.identityString,
    activeArrangement.value.targetString,
    axisLabelsFromAxes(activeDataSet.value?.settings?.axes) ?? activeDataSet.value?.axisLabels
  )
)

// Per-chart resolved compute params — each chart type carries its own sort/showLabels/scale.
const resolvedSort = computed(() => resolved('sort') as Sort)
const resolvedShowLabels = computed(() => resolved('showLabels') as boolean)
const resolvedScale = computed(() => resolved('scale') as ScaleType)

// Charts are computed off-thread in a worker, one at a time (queue-based). Each
// slot carries its own `pending` so its card drives an independent skeleton and
// reveals progressively.
const { charts, groupNames } = useChartPipeline(
  activeResults,
  activeArrangement,
  activeLabels,
  activeGroupId,
  resolvedSort,
  resolvedShowLabels,
  resolvedScale
)

// The worker owns grouping; feed its group list back into useDataPoint so the
// selector and URL router stay worker-backed.
watch(groupNames, (names) => setGroupNames(names))

// Full-page skeleton only while loading the dataset. Once data is ready the header
// and layout appear immediately; each chart card drives its own skeleton while pending.
const showSkeleton = computed(() => loading.value)

useDashboardInit()
</script>

<template>
  <nav class="fixed right-6 top-6 z-50 flex items-center gap-2">
    <IconButton
      v-if="activeDataSet?.pkg"
      :href="`https://${activeDataSet.pkg}`"
      aria-label="View Package Source"
    >
      <Package class="h-5 w-5" />
    </IconButton>

    <ChartSettingsPopover />

    <IconButton @click="toggleDark()" aria-label="Toggle theme">
      <Sun v-if="settings.isDark" class="h-5 w-5" />
      <Moon v-else class="h-5 w-5" />
    </IconButton>
  </nav>

  <LoadError v-if="loadError" :message="loadError" />

  <LoadingSkeleton v-else-if="showSkeleton" />

  <main v-else-if="activeDataSet" class="mx-auto min-h-screen max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
    <DataSetHeader
      :dataSet="activeDataSet"
      :dataSets="dataSets"
      :activeDataSetId="activeDataSetId"
      :resultGroups="resultGroups"
      :activeGroupId="activeGroupId"
      @selectDataSet="selectDataSet"
      @selectGroup="selectGroup"
    />

    <div class="space-y-5">
      <template v-for="(state, index) in charts" :key="state.key">
        <ChartCard
          v-if="state.data"
          :chartData="state.data"
          :loading="state.pending"
          class="animate-fade-in"
          :style="{ animationDelay: `${index * 50}ms` }"
        />
        <div
          v-else
          class="rounded-lg border border-border bg-card p-6 shadow-sm"
        >
          <div class="mb-4 h-6 w-48 animate-pulse rounded bg-muted" />
          <div class="h-[600px] animate-pulse rounded bg-muted" />
        </div>
      </template>
    </div>
  </main>

  <AppFooter v-if="activeDataSet && !showSkeleton && !loadError" :version="version" />
</template>
