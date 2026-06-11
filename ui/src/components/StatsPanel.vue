<script setup lang="ts">
import { computed, onMounted, ref, watch, nextTick } from 'vue'
import { Download, ArrowUp, ArrowDown } from 'lucide-vue-next'
import type { ChartData, DescriptiveStats, SeriesProfile, CorrelationMatrix } from '../types'
import { computeStats } from '../composables/useStatsWorker'
import { descriptiveCsv, correlationCsv } from '../lib/csv'
import SelectionTabs from './SelectionTabs.vue'

const props = defineProps<{ chartData: ChartData }>()

// Stats are computed lazily off-thread (dedicated stats worker) the first time
// the panel opens for a chart, and recomputed when the chart's data changes. A
// per-ChartData cache in useStatsWorker makes reopening the same chart instant.
const loading = ref(true)
const profiles = ref<SeriesProfile[]>([])
const correlation = ref<CorrelationMatrix | undefined>(undefined)

// Monotonic token so a slow reply for a superseded chartData is discarded.
let token = 0
async function load() {
  const mine = ++token
  loading.value = true
  const result = await computeStats(props.chartData)
  if (mine !== token) return // a newer chartData superseded this request
  // New chart starts in natural series order with no active search filter.
  sortKey.value = null
  sortDir.value = 'desc'
  searchQuery.value = ''
  debouncedQuery.value = ''
  profiles.value = result.seriesProfiles
  correlation.value = result.correlation
  loading.value = false
  // If correlation vanished (e.g. fewer series after a recompute) while we were
  // sitting on the correlation tab, fall back so the view isn't blank.
  if (!correlation.value && view.value === 'correlation') view.value = 'descriptive'
  resetScroll()
}

onMounted(load)
watch(() => props.chartData, load)

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
const seriesLabel = computed(() => props.chartData.axisLabels?.x || 'Series')

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

// Widened to string|number so SelectionTabs' v-model (emits string|number) types
// cleanly; only ever set to the literal option values below.
const view = ref<string | number>('descriptive')
const method = ref<string | number>('pearson')

const viewOptions = computed(() => {
  const opts = [{ value: 'descriptive', label: 'Descriptive' }]
  if (correlation.value) opts.push({ value: 'correlation', label: 'Correlation' })
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
watch([view, loading], () => nextTick(measure))

const startIndex = computed(() => Math.max(0, Math.floor(scrollTop.value / ROW_H) - OVERSCAN))
const endIndex = computed(() =>
  Math.min(filteredProfiles.value.length, startIndex.value + Math.ceil(viewportH.value / ROW_H) + OVERSCAN * 2)
)
const visibleRows = computed(() => filteredProfiles.value.slice(startIndex.value, endIndex.value))
const topPad = computed(() => startIndex.value * ROW_H)
const bottomPad = computed(() => (filteredProfiles.value.length - endIndex.value) * ROW_H)

// Diverging heatmap colour for r ∈ [-1,1]: red (negative) → neutral → green
// (positive). Explicit bg + text colour so cells read in both light and dark
// themes. NaN cells (constant series / too few pairs) stay neutral grey.
function corrColor(v: number): string {
  if (!Number.isFinite(v)) return 'transparent'
  const t = Math.min(Math.abs(v), 1)
  const hue = v >= 0 ? 145 : 8
  const light = 92 - t * 42
  return `hsl(${hue} 70% ${light}%)`
}
function corrText(v: number): string {
  return Number.isFinite(v) && Math.abs(v) > 0.55 ? '#fff' : 'inherit'
}
function corrCell(v: number): string {
  return Number.isFinite(v) ? Number(v.toFixed(2)).toString() : '—'
}

// --- CSV export (active view) ----------------------------------------------
// Exports whichever table is shown: descriptive (series × metrics, full
// precision) or the correlation matrix for the active method. Built on the main
// thread from already-in-memory state — a one-shot user action, no worker hop.
const canDownload = computed(
  () =>
    !loading.value &&
    (view.value === 'descriptive' ? profiles.value.length > 0 : !!correlation.value)
)

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

    <!-- Lazy compute in progress. -->
    <div v-if="loading" class="space-y-2 py-2">
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
      <div v-if="profiles.length > 20" class="mb-2 flex items-center gap-2">
        <input
          v-model="searchQuery"
          type="search"
          :placeholder="`Search ${seriesLabel}...`"
          class="max-w-xs rounded-md border border-border bg-background px-3 py-1.5 text-xs text-card-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring"
        />
        <span v-if="debouncedQuery" class="text-xs text-muted-foreground">
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

    <!-- Correlation heatmap (capped series count, no virtualization needed). -->
    <div v-else-if="correlation" class="max-h-[28rem] overflow-auto">
      <table class="border-collapse text-xs">
        <thead>
          <tr>
            <th class="sticky left-0 top-0 z-20 bg-card px-2 py-1"></th>
            <th
              v-for="label in correlation.labels"
              :key="label"
              class="sticky top-0 z-10 max-w-[8rem] truncate bg-muted px-2 py-1 font-medium text-muted-foreground"
              :title="label"
            >
              {{ label || '—' }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(label, i) in correlation.labels" :key="label">
            <th
              class="sticky left-0 z-10 max-w-[10rem] truncate bg-card px-2 py-1 text-left font-medium text-card-foreground"
              :title="label"
            >
              {{ label || '—' }}
            </th>
            <td
              v-for="(val, j) in corrMatrix[i]"
              :key="j"
              class="px-2 py-1 text-center tabular-nums"
              :style="{ background: corrColor(val), color: corrText(val) }"
            >
              {{ corrCell(val) }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
