<script setup lang="ts">
import { computed, watch } from 'vue'
import { Moon, Sun, Package } from 'lucide-vue-next'
import type { Axis, AxisLabels } from '../types'
import { useDataPoint } from '../composables/useDataPoint'
import { useChartPipeline } from '../composables/useChartPipeline'
import { useSettingsStore } from '../composables/useSettingsStore'
import { useActiveChartShape } from '../composables/useActiveChartShape'
import { useDashboardInit } from '../composables/useDashboardInit'
import { swapAxisLabels } from '../lib/swap'
import ChartSettingsPopover from '../components/ChartSettingsPopover.vue'
import ChartCard from '../components/ChartCard.vue'
import DataSetHeader from '../components/DataSetHeader.vue'
import LoadingSkeleton from '../components/LoadingSkeleton.vue'
import LoadError from '../components/LoadError.vue'
import AppFooter from '../components/AppFooter.vue'
import IconButton from '../components/IconButton.vue'
import Selector from '../components/Selector.vue'
import { isThemeName, THEME_NAMES } from '../lib/themes'

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
  lazyCatalog,
  detailLoading,
  detailError,
  retryActiveDataSet,
} = useDataPoint()

const { isDark, toggleDark, chartType, themeName, setTheme } = useSettingsStore()
const { sort, showLabels, scale, threeD } = useActiveChartShape()

const themeItems = computed(() => {
  const items = THEME_NAMES.map((theme) => ({ name: theme, value: theme }))
  return isThemeName(themeName.value)
    ? items
    : [{ name: 'Custom palette', value: themeName.value }, ...items]
})

// Build an AxisLabels flat map from dataset.axes for swapAxisLabels.
const axisLabelsFromAxes = (axes: Axis[] | undefined): AxisLabels | undefined => {
  if (!axes?.length) return undefined
  const result: Record<string, string> = {}
  for (const a of axes) {
    if (a.label) result[a.key] = a.label
  }
  return Object.keys(result).length ? (result as AxisLabels) : undefined
}

// The full raw rows — the worker owns grouping/projection, so we pass the dataset
// as-is (no main-thread grouping or swap mutation). Only a dataset switch re-clones.
const activeResults = computed(() => activeDataSet.value?.data || [])
// Display labels from axes[], permuted to match the active arrangement.
const activeLabels = computed(() =>
  swapAxisLabels(
    activeArrangement.value.identityString,
    activeArrangement.value.targetString,
    axisLabelsFromAxes(activeDataSet.value?.axes)
  )
)

// Per-chart resolved compute params come from `useActiveChartShape`, which reads
// the active chart's typed config and applies `?? default` for missing fields.
// `sort` defaults to disabled when absent so the worker treats it as a no-op.
const resolvedSort = computed(() => sort.value ?? { enabled: false, order: 'asc' as const })

// Value-mode axes from the active dataset (undefined for category-mode datasets).
const activeAxes = computed(() => activeDataSet.value?.axes)

// Charts are computed off-thread in a worker, one at a time (queue-based). Each
// slot carries its own `pending` so its card drives an independent skeleton and
// reveals progressively.
const { charts, groupNames, datasetInitializing } = useChartPipeline(
  activeResults,
  activeArrangement,
  activeLabels,
  activeGroupId,
  resolvedSort,
  showLabels,
  scale,
  threeD,
  activeAxes,
  chartType,
  computed(() => activeDataSet.value?.preserveRows === true),
  computed(() => activeDataSet.value?.name)
)

// The worker owns grouping; feed its group list back into useDataPoint so the
// selector and URL router stay worker-backed.
watch(groupNames, (names) => setGroupNames(names))

// Full-page skeleton only while loading the dataset. Once data is ready the header
// and layout appear immediately; each chart card drives its own skeleton while pending.
const showSkeleton = computed(() => loading.value)
const showDetailSkeleton = computed(
  () =>
    lazyCatalog.value &&
    detailError.value === null &&
    (detailLoading.value ||
      datasetInitializing.value ||
      ((activeDataSet.value?.data.length ?? 0) > 0 && charts.value.length === 0))
)

useDashboardInit()
</script>

<template>
  <nav class="fixed right-6 top-6 z-50 flex items-center gap-2">
    <IconButton
      v-if="activeDataSet?.meta?.pkg"
      :href="`https://${activeDataSet.meta?.pkg}`"
      aria-label="View Package Source"
    >
      <Package class="h-5 w-5" />
    </IconButton>

    <ChartSettingsPopover />

    <Selector
      :items="themeItems"
      :activeValue="themeName"
      placeholder="Search themes"
      notFoundText="No theme found"
      ariaLabel="Chart color theme"
      class="w-36"
      triggerClass="h-12 px-3"
      @selectValue="setTheme"
    />

    <IconButton @click="toggleDark()" aria-label="Toggle dark mode">
      <Sun v-if="isDark" class="h-5 w-5" />
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

    <LoadError v-if="detailError" :message="detailError" :retry="retryActiveDataSet" inline />

    <LoadingSkeleton v-else-if="showDetailSkeleton" contentOnly />

    <div v-else class="space-y-5">
      <template v-for="(state, index) in charts" :key="state.key">
        <ChartCard
          v-if="state.data"
          :chartData="state.data"
          :loading="state.pending"
          class="animate-fade-in"
          :style="{ animationDelay: `${index * 50}ms` }"
        />
        <div v-else class="rounded-lg border border-border bg-card p-6 shadow-sm">
          <div class="mb-4 h-6 w-48 animate-pulse rounded bg-muted" />
          <div class="h-[600px] animate-pulse rounded bg-muted" />
        </div>
      </template>
    </div>
  </main>

  <AppFooter
    v-if="activeDataSet && !showSkeleton && !showDetailSkeleton && !loadError && !detailError"
    :version="version"
  />
</template>
