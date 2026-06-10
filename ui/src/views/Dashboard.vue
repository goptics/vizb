<script setup lang="ts">
import { computed, toRef, watch } from 'vue'
import { Moon, Sun, Package } from 'lucide-vue-next'
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
  activeGroupName,
  selectGroup,
  setGroupNames,
  loading,
  loadError,
} = useDataPoint()

const { settings, toggleDark } = useSettingsStore()

// The full raw rows — the worker owns grouping/projection, so we pass the dataset
// as-is (no main-thread grouping or swap mutation). Only a dataset switch re-clones.
const activeResults = computed(() => activeDataSet.value?.data || [])
// Display labels are derived from the arrangement (swap rotates which dimension
// each label sits on), not mutated onto the dataset.
const activeLabels = computed(() =>
  swapAxisLabels(
    activeArrangement.value.identityString,
    activeArrangement.value.targetString,
    activeDataSet.value?.axisLabels
  )
)
// Charts are computed off-thread in a worker, one at a time (queue-based). Each
// slot carries its own `pending` so its card drives an independent skeleton and
// reveals progressively.
const { charts, hasAny, groupNames } = useChartPipeline(
  activeResults,
  activeArrangement,
  activeLabels,
  activeGroupName,
  toRef(settings, 'sort'),
  toRef(settings, 'showLabels'),
  toRef(settings, 'scale')
)

// The worker owns grouping; feed its group list back into useDataPoint so the
// selector and URL router stay worker-backed.
watch(groupNames, (names) => setGroupNames(names))

// Full-page skeleton only while loading the dataset or on the very first compute
// (no chart has data yet). Later recomputes keep existing charts visible and let
// each card show its own skeleton.
const showSkeleton = computed(() => loading.value || (!hasAny.value && charts.value.length > 0))

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
      </template>
    </div>
  </main>

  <AppFooter v-if="activeDataSet && !showSkeleton && !loadError" :version="version" />
</template>
