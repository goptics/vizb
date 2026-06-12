<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, ref, toRefs, watch, nextTick } from 'vue'
import { Download, ArrowUp, ArrowDown } from 'lucide-vue-next'
import type {
  ChartData,
  DescriptiveStats,
  SeriesProfile,
  CorrelationMatrix,
} from '../types'
import { computeDescriptive, computeCorrelation } from '../composables/useStatsWorker'
import { availableViews } from '../lib/stats'
import { buildCorrelationOption } from '../composables/charts/useCorrelationOption'
import { useSettingsStore } from '../composables/useSettingsStore'
import { descriptiveCsv, correlationCsv } from '../lib/csv'
import SelectionTabs from './SelectionTabs.vue'

// Correlation renders as an echarts canvas heatmap, lazy-loaded so its echarts
// modules parse only when the correlation tab opens (same discipline as ChartCard).
const ChartCorrelation = defineAsyncComponent(() => import('./ChartCorrelation.vue'))

const props = defineProps<{ chartData: ChartData }>()

const { settings } = useSettingsStore()
const { isDark } = toRefs(settings)
const initOptions = { renderer: 'canvas', devicePixelRatio: window.devicePixelRatio } as const

// Active view + its sub-toggles. Widened to string|number so SelectionTabs'
// v-model (emits string|number) types cleanly; only ever set to literal values.
const view = ref<string | number>('descriptive')
const method = ref<string | number>('pearson')

// Descriptive stats are computed eagerly off-thread when the panel opens;
// correlation is deferred until its tab is first clicked (lazy), then cached per
// ChartData in useStatsWorker. Each piece has its own loading flag so the active
// tab shows a skeleton only while *its* data is in flight.
const profiles = ref<SeriesProfile[]>([])
const correlation = ref<CorrelationMatrix | undefined>(undefined)
const descLoading = ref(true)
const corrLoading = ref(false)

// Which views are valid for this data shape — cheap O(K) precondition flags
// (counts + one numeric-axis parse, no heavy math), so the tabs render before any
// expensive compute runs. Recomputed synchronously per chartData.
const available = ref(availableViews([], []))

// Skeleton shows while the *active* view's data is being computed.
const activeLoading = computed(() => {
  switch (view.value) {
    case 'correlation':
      return corrLoading.value
    default:
      return descLoading.value
  }
})

// Monotonic token so a slow reply for a superseded chartData is discarded.
let token = 0

// Lazily fetch the heavy piece backing a view (no-op for descriptive, or when
// already loaded / in flight / unavailable). The worker layer caches per chart,
// so re-clicking a tab never re-posts.
async function ensureCorrelation() {
  if (correlation.value !== undefined || corrLoading.value || !available.value.correlation) return
  corrLoading.value = true
  const mine = token
  const r = await computeCorrelation(props.chartData)
  if (mine !== token) return
  correlation.value = r
  corrLoading.value = false
}
function ensureForView(v: string | number) {
  if (v === 'correlation') ensureCorrelation()
}

async function load() {
  const mine = ++token
  // Reset per-chart state. Availability is synchronous, so the tabs are correct
  // immediately; the descriptive table then fills in off-thread, and the heavier
  // tabs compute on first open.
  descLoading.value = true
  corrLoading.value = false
  correlation.value = undefined
  available.value = availableViews(
    props.chartData.series.map((s) => s.xAxis),
    props.chartData.yAxis,
    props.chartData.zAxis
  )
  // New chart starts in natural series order with no active search filter.
  sortKey.value = null
  sortDir.value = 'desc'
  searchQuery.value = ''
  debouncedQuery.value = ''
  // If the active tab is no longer valid for this data shape, fall back so the
  // view isn't blank.
  if (!viewOptions.value.some((o) => o.value === view.value)) view.value = 'descriptive'

  const result = await computeDescriptive(props.chartData)
  if (mine !== token) return // a newer chartData superseded this request
  profiles.value = result
  descLoading.value = false
  // Kick off the lazy load for whatever non-descriptive tab is active.
  ensureForView(view.value)
  resetScroll()
}

onMounted(load)
watch(() => props.chartData, load)
// Fetch a tab's backing data the first time it's selected.
watch(view, (v) => ensureForView(v))

// Descriptive table columns, in display order. `key` indexes DescriptiveStats.
const COLUMNS: { key: keyof DescriptiveStats; label: string }[] = [
  { key: 'count', label: 'Count' },
  { key: 'missing', label: 'Missing' },
  { key: 'unique', label: 'Unique' },
  { key: 'mean', label: 'Mean' },
  { key: 'median', label: 'Median' },
  { key: 'mode', label: 'Mode' },
  { key: 'stdDev', label: 'SD' },
  { key: 'variance', label: 'Variance' },
  { key: 'min', label: 'Min' },
  { key: 'max', label: 'Max' },
  { key: 'range', label: 'Range' },
  { key: 'iqr', label: 'IQR' },
  { key: 'mad', label: 'MAD' },
  { key: 'cv', label: 'CV' },
  { key: 'skewness', label: 'Skew' },
  { key: 'kurtosis', label: 'Kurtosis' },
  { key: 'p5', label: 'P5' },
  { key: 'p25', label: 'P25' },
  { key: 'p75', label: 'P75' },
  { key: 'p95', label: 'P95' },
]

// Integer-valued metadata columns shown as plain integers; CV as a percentage;
// everything else as a compact significant-figure number.
const INT_KEYS = new Set<keyof DescriptiveStats>(['count', 'missing', 'unique'])

function fmt(key: keyof DescriptiveStats, v: number): string {
  if (!Number.isFinite(v)) return '—'
  if (INT_KEYS.has(key)) return String(v)
  if (key === 'cv') return `${(v * 100).toFixed(1)}%`
  if (Number.isInteger(v)) return String(v)
  const a = Math.abs(v)
  if (a !== 0 && (a < 1e-3 || a >= 1e6)) return v.toExponential(2)
  return Number(v.toPrecision(4)).toString()
}

// --- Click-to-sort the descriptive table ----------------------------------
// Pure view concern over the already-computed profiles (no worker round-trip).
// `'name'` is the Series column; every other key indexes DescriptiveStats.
type SortKey = keyof DescriptiveStats | 'name'
const sortKey = ref<SortKey | null>(null) // null = original series order
const sortDir = ref<'asc' | 'desc'>('desc')

// Tri-state cycle on a header click: new col → desc; same col desc → asc;
// same col asc → back to original (null).
function toggleSort(key: SortKey) {
  if (sortKey.value !== key) {
    sortKey.value = key
    sortDir.value = 'desc'
  } else if (sortDir.value === 'desc') {
    sortDir.value = 'asc'
  } else {
    sortKey.value = null
  }
  resetScroll() // top of the freshly ordered list
}

// Sorted view of `profiles`. Non-finite values (NaN/±Inf — the "—" cells) always
// sink to the bottom regardless of direction, so a mostly-missing column still
// surfaces its real min/max. Stable copy — never mutate profiles.value.
const sortedProfiles = computed(() => {
  const key = sortKey.value
  if (!key) return profiles.value
  const dir = sortDir.value
  return profiles.value.slice().sort((a, b) => {
    if (key === 'name') {
      const c = a.name.localeCompare(b.name)
      return dir === 'asc' ? c : -c
    }
    const av = a.stats[key]
    const bv = b.stats[key]
    const an = !Number.isFinite(av)
    const bn = !Number.isFinite(bv)
    if (an && bn) return 0
    if (an) return 1 // NaN always last
    if (bn) return -1
    return dir === 'asc' ? av - bv : bv - av
  })
})

// User-supplied label for the series (xAxis) dimension, falls back to "Series".
// Used by the descriptive table and the search box — both key rows by series (xAxis).
const seriesLabel = computed(() => props.chartData.axisLabels?.x || 'Series')

// --- Correlation caption ---------------------------------------------------
// The correlation matrix auto-picks which axis supplies its entities (x → y → z;
// see selectCorrelationAxis). Caption names the chosen dimension and what it's
// correlated across, flipping with the worker's choice.
const AXIS_FALLBACK = { x: 'Series', y: 'the y-axis', z: 'the z-axis' } as const
const corrAxis = computed<'x' | 'y' | 'z'>(() => correlation.value?.axis ?? 'x')
const corrEntityLabel = computed(() => {
  const a = corrAxis.value
  return props.chartData.axisLabels?.[a] || AXIS_FALLBACK[a]
})
// The other present dimensions (≥2 distinct values) the entities are correlated over.
const corrObsLabel = computed(() => {
  const a = corrAxis.value
  const counts = {
    x: props.chartData.series.length,
    y: props.chartData.yAxis.length,
    z: props.chartData.zAxis.length,
  }
  const labels = props.chartData.axisLabels
  const names = (['x', 'y', 'z'] as const)
    .filter((ax) => ax !== a && counts[ax] >= 2)
    .map((ax) => labels?.[ax] || AXIS_FALLBACK[ax])
  return names.length ? names.join(' × ') : 'the other dimensions'
})

// --- Search / filter the descriptive table --------------------------------
// Shown only when > 20 series. Debounced 300 ms so fast typists don't trigger
// a filter recompute on every keystroke. No external debounce lib in project.
const searchQuery = ref('')
const debouncedQuery = ref('')
let debounceTimer: ReturnType<typeof setTimeout> | null = null
watch(searchQuery, (val) => {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => { debouncedQuery.value = val }, 300)
})

// Filter sortedProfiles by case-insensitive substring on series name. When
// the query is empty, return sortedProfiles directly (no allocation).
const filteredProfiles = computed(() => {
  const q = debouncedQuery.value.trim().toLowerCase()
  if (!q) return sortedProfiles.value
  return sortedProfiles.value.filter((p) => p.name.toLowerCase().includes(q))
})

// Tabs are driven by the cheap availability flags, not by whether the heavy
// result is already in memory — so a tab can show (and trigger its lazy compute)
// before its data exists.
const viewOptions = computed(() => {
  const opts = [{ value: 'descriptive', label: 'Descriptive' }]
  const a = available.value
  if (a.correlation) opts.push({ value: 'correlation', label: 'Correlation' })
  return opts
})
const methodOptions = [
  { value: 'pearson', label: 'Pearson' },
  { value: 'spearman', label: 'Spearman' },
]

const corrMatrix = computed(() => {
  const c = correlation.value
  if (!c) return []
  return method.value === 'spearman' ? c.spearman : c.pearson
})

// Heatmap option for the active correlation method, re-themed on dark-mode toggle.
const corrOption = computed(() => {
  const c = correlation.value
  if (!c) return null
  return buildCorrelationOption(c.labels, corrMatrix.value, isDark.value)
})

// --- Virtualized descriptive table -----------------------------------------
// The descriptive table can have a very large number of series rows. Render only
// the rows in (and just around) the viewport, padding the rest with two spacer
// rows so the scrollbar still reflects the full height.
const ROW_H = 32 // px; must match the fixed row height in the template
const OVERSCAN = 8

const scrollEl = ref<HTMLElement | null>(null)
const scrollTop = ref(0)
const viewportH = ref(448) // ≈ max-h-[28rem]; replaced by the real height on mount

function measure() {
  if (scrollEl.value) viewportH.value = scrollEl.value.clientHeight
}
function onScroll() {
  if (scrollEl.value) scrollTop.value = scrollEl.value.scrollTop
  measure()
}
function resetScroll() {
  scrollTop.value = 0
  nextTick(() => {
    if (scrollEl.value) scrollEl.value.scrollTop = 0
    measure()
  })
}

// Re-measure whenever the descriptive table (re)mounts (e.g. switching back from
// the correlation tab, or after the loading skeleton clears).
watch([view, activeLoading], () => nextTick(measure))

const startIndex = computed(() => Math.max(0, Math.floor(scrollTop.value / ROW_H) - OVERSCAN))
const endIndex = computed(() =>
  Math.min(filteredProfiles.value.length, startIndex.value + Math.ceil(viewportH.value / ROW_H) + OVERSCAN * 2)
)
const visibleRows = computed(() => filteredProfiles.value.slice(startIndex.value, endIndex.value))
const topPad = computed(() => startIndex.value * ROW_H)
const bottomPad = computed(() => (filteredProfiles.value.length - endIndex.value) * ROW_H)

// --- CSV export (active view) ----------------------------------------------
// Exports whichever table is shown: descriptive (series × metrics, full
// precision) or the correlation matrix for the active method. Built on the main
// thread from already-in-memory state — a one-shot user action, no worker hop.
const canDownload = computed(() => {
  if (activeLoading.value) return false
  switch (view.value) {
    case 'correlation':
      return !!correlation.value
    default:
      return profiles.value.length > 0
  }
})

function slug(s: string): string {
  return (s || 'stats').toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '') || 'stats'
}

function downloadCsv() {
  if (!canDownload.value) return
  let csv: string
  let suffix: string
  if (view.value === 'correlation' && correlation.value) {
    csv = correlationCsv(correlation.value.labels, corrMatrix.value)
    suffix = `correlation-${method.value}`
  } else {
    csv = descriptiveCsv(filteredProfiles.value, COLUMNS)
    suffix = 'descriptive'
  }
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${slug(props.chartData.statType)}-${suffix}.csv`
  a.click()
  URL.revokeObjectURL(url)
}
</script>

<template>
  <div class="mt-4 rounded-md border border-border bg-muted/30 p-4">
    <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
      <p class="text-sm font-medium text-card-foreground">
        Statistics
        <span class="font-normal text-muted-foreground">· {{ chartData.statType }}</span>
      </p>
      <div class="flex items-center gap-2">
        <SelectionTabs
          v-if="view === 'correlation'"
          v-model="method"
          :options="methodOptions"
        />
        <SelectionTabs v-if="viewOptions.length > 1" v-model="view" :options="viewOptions" />
        <button
          type="button"
          class="inline-flex h-8 items-center gap-1.5 rounded-md border border-border px-2.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-primary disabled:pointer-events-none disabled:opacity-50"
          :disabled="!canDownload"
          title="Download current view as CSV"
          @click="downloadCsv"
        >
          <Download class="h-4 w-4" />
          CSV
        </button>
      </div>
    </div>

    <!-- Active tab's data is being computed off-thread. -->
    <div v-if="activeLoading" class="space-y-2 py-2">
      <div class="h-4 w-40 animate-pulse rounded bg-muted" />
      <div class="h-8 w-full animate-pulse rounded bg-muted" />
      <div class="h-8 w-full animate-pulse rounded bg-muted" />
      <div class="h-8 w-full animate-pulse rounded bg-muted" />
    </div>

    <p
      v-else-if="!profiles.length"
      class="py-6 text-center text-sm text-muted-foreground"
    >
      No statistics available for this chart.
    </p>

    <!-- Descriptive: series (rows) × metrics (cols). Virtualized vertically,
         horizontal scroll for the wide metric set. -->
    <template v-else-if="view === 'descriptive'">
      <div v-if="profiles.length > 20" class="mb-2 flex items-center">
        <input
          v-model="searchQuery"
          type="search"
          :placeholder="`Search ${seriesLabel}...`"
          class="max-w-xs rounded-md border border-border bg-background px-3 py-1.5 text-xs text-card-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring"
        />
        <span v-if="debouncedQuery" class="ml-auto text-xs text-muted-foreground">
          {{ filteredProfiles.length }} / {{ profiles.length }}
        </span>
      </div>
      <div
        ref="scrollEl"
        class="max-h-[28rem] overflow-auto"
        @scroll="onScroll"
      >
      <table class="w-full border-collapse text-right text-xs">
        <thead>
          <tr class="border-b border-border text-muted-foreground">
            <th
              class="sticky left-0 top-0 z-20 cursor-pointer select-none bg-card px-2 py-1.5 text-left font-medium hover:text-primary"
              @click="toggleSort('name')"
            >
              <span class="inline-flex items-center gap-1">
                {{ seriesLabel }}
                <ArrowDown v-if="sortKey === 'name' && sortDir === 'desc'" class="h-3 w-3" />
                <ArrowUp v-else-if="sortKey === 'name' && sortDir === 'asc'" class="h-3 w-3" />
              </span>
            </th>
            <th
              v-for="col in COLUMNS"
              :key="col.key"
              class="sticky top-0 z-10 cursor-pointer select-none bg-muted px-2 py-1.5 font-medium hover:text-primary"
              @click="toggleSort(col.key)"
            >
              <span class="inline-flex items-center justify-end gap-1">
                {{ col.label }}
                <ArrowDown v-if="sortKey === col.key && sortDir === 'desc'" class="h-3 w-3" />
                <ArrowUp v-else-if="sortKey === col.key && sortDir === 'asc'" class="h-3 w-3" />
              </span>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="topPad" :style="{ height: topPad + 'px' }">
            <td :colspan="COLUMNS.length + 1" class="p-0" />
          </tr>
          <tr v-if="debouncedQuery && !filteredProfiles.length">
            <td :colspan="COLUMNS.length + 1" class="py-6 text-center text-muted-foreground">
              No series match "{{ debouncedQuery }}"
            </td>
          </tr>
          <tr
            v-for="p in visibleRows"
            :key="p.name"
            class="border-b border-border/50"
            :style="{ height: ROW_H + 'px' }"
          >
            <th
              class="sticky left-0 z-10 max-w-[12rem] truncate bg-card px-2 text-left font-medium text-card-foreground"
              :title="p.name"
            >
              {{ p.name || '—' }}
            </th>
            <td
              v-for="col in COLUMNS"
              :key="col.key"
              class="px-2 tabular-nums text-card-foreground"
            >
              {{ fmt(col.key, p.stats[col.key]) }}
            </td>
          </tr>
          <tr v-if="bottomPad" :style="{ height: bottomPad + 'px' }">
            <td :colspan="COLUMNS.length + 1" class="p-0" />
          </tr>
        </tbody>
      </table>
    </div>
    </template>

    <!-- Correlation heatmap — echarts canvas, scales past what a DOM grid can.
         Entity axis is auto-picked (x → y → z); the caption names it. -->
    <template v-else-if="view === 'correlation' && corrOption">
      <p class="mb-2 text-xs text-muted-foreground">
        Correlating <span class="font-medium">{{ corrEntityLabel }}</span> across
        {{ corrObsLabel }}.
      </p>
      <div class="h-[28rem]">
        <ChartCorrelation :option="corrOption" :init-options="initOptions" class="h-full w-full" />
      </div>
    </template>
  </div>
</template>
