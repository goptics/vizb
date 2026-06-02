<script setup lang="ts">
import { computed, watch } from 'vue'
import { Moon, Sun, Package, Cpu } from 'lucide-vue-next'
import { useBenchmarkData } from '../composables/useBenchmarkData'
import { useChartData } from '../composables/useChartData'
import { useSettingsStore } from '../composables/useSettingsStore'
import { useUrlRouter } from '../composables/useUrlRouter'
import ChartSettingsPopover from '../components/ChartSettingsPopover.vue'
import GroupSelector from '../components/Selector.vue'
import ChartCard from '../components/ChartCard.vue'
import Badge from '../components/Badge.vue'
import TimestampBadge from '../components/TimestampBadge.vue'
import IconButton from '../components/IconButton.vue'
import AccentLink from '../components/AccentLink.vue'
import { CPUtoString } from '../lib/utils'

const version = window.VIZB_VERSION || 'v0.0.0-dev'

const {
  // Top level benchmark selection
  benchmarks,
  activeBenchmark,
  activeBenchmarkId,
  selectBenchmark,

  // Inner level group selection
  resultGroups,
  activeGroup,
  activeGroupId,
  selectGroup,

  loading,
  loadError,
} = useBenchmarkData()

// Use the active group's results for chart data
const activeResults = computed(() => activeGroup.value?.data || [])
const { chartData } = useChartData(activeResults)

const { settings, toggleDark, initializeFromBenchmark } = useSettingsStore()
const { initFromUrl } = useUrlRouter()

let urlInitialized = false

// Initialize settings from the active benchmark settings and update title
watch(
  activeBenchmark,
  (b) => {
    // Update document title
    if (b?.name) {
      document.title = `Vizb | ${b.name}`
    }

    // Initialize settings from the active benchmark settings
    if (b?.settings) {
      initializeFromBenchmark(b.settings)
    }
  },
  { immediate: true }
)

// Initialize from URL once benchmarks are loaded
watch(
  benchmarks,
  (b) => {
    if (b.length && !urlInitialized) {
      initFromUrl()
      urlInitialized = true
    }
  },
  { immediate: true }
)

// Get the main constant title (use description as main title)
const mainTitle = computed(() => {
  // Use the description from the first benchmark as the constant title
  return benchmarks.value[0]?.name || 'Benchmarks'
})

const hasCPU = computed(() => activeBenchmark.value?.cpu?.name || activeBenchmark.value?.cpu?.cores)
</script>

<template>
  <nav class="fixed right-6 top-6 z-50 flex items-center gap-2">
    <IconButton
      v-if="activeBenchmark?.pkg"
      :href="`https://${activeBenchmark.pkg}`"
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

  <div v-if="loading" class="mx-auto min-h-screen max-w-7xl animate-pulse px-4 py-8 sm:px-6 lg:px-8">
    <header class="space-y-3 py-5 text-center">
      <div class="mx-auto h-9 w-64 rounded-md bg-muted"></div>
      <div class="flex justify-center gap-2">
        <div class="h-6 w-28 rounded-full bg-muted"></div>
        <div class="h-6 w-24 rounded-full bg-muted"></div>
      </div>
      <div class="mx-auto h-4 w-48 rounded bg-muted"></div>
    </header>
    <div class="space-y-5">
      <div
        v-for="i in 3"
        :key="i"
        class="rounded-lg border border-border bg-card p-6 shadow-sm"
      >
        <div class="mb-4 h-5 w-40 rounded bg-muted"></div>
        <div class="h-[500px] rounded bg-muted"></div>
      </div>
    </div>
  </div>

  <div
    v-else-if="loadError"
    class="flex min-h-screen items-center justify-center px-4 text-center"
  >
    <div class="max-w-md space-y-2">
      <p class="font-medium text-destructive">Failed to load benchmark data</p>
      <p class="break-all text-sm text-muted-foreground">{{ loadError }}</p>
      <p class="text-xs text-muted-foreground/60">
        Ensure the data URL is reachable and the server sends
        <code>Access-Control-Allow-Origin: *</code>.
      </p>
    </div>
  </div>

  <main v-else-if="activeBenchmark" class="mx-auto min-h-screen max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
    <header class="space-y-3 py-5 text-center">
      <GroupSelector
        v-if="benchmarks.length > 1"
        :items="benchmarks"
        :activeId="activeBenchmarkId"
        @select="selectBenchmark"
        class="mx-auto min-w-80"
        placeholder="Search Benchmark..."
        notFoundText="No benchmark found."
      />

      <h1 v-else class="text-4xl font-bold">{{ mainTitle }}</h1>

      <div class="flex justify-center">
        <Badge v-if="hasCPU" :icon="Cpu" label="CPU" :value="CPUtoString(activeBenchmark?.cpu)" />
      </div>

      <TimestampBadge
        v-if="activeBenchmark?.timestamp"
        :timestamp="activeBenchmark.timestamp"
        :history="activeBenchmark.history"
      />

      <p v-if="activeBenchmark?.description" class="text-muted-foreground">
        {{ activeBenchmark?.description }}
      </p>

      <!-- Inner Group Selector -->
      <GroupSelector
        v-if="resultGroups.length > 1"
        :items="resultGroups"
        :activeId="activeGroupId"
        @select="selectGroup"
        placeholder="Search Group..."
        notFoundText="No group found."
        class="mx-auto min-w-80"
      />
    </header>

    <!-- Charts Grid -->
    <div class="space-y-5">
      <ChartCard
        v-for="(chart, index) in chartData"
        :key="`${activeBenchmarkId}-${activeGroupId}-${index}`"
        :chartData="chart"
        class="animate-fade-in"
        :style="{ animationDelay: `${index * 50}ms` }"
      />
    </div>
  </main>

  <footer v-if="activeBenchmark && !loading && !loadError" class="pb-5 text-center text-sm text-muted-foreground">
    Generated by
    <AccentLink href="https://vizb.goptics.org"> Vizb </AccentLink>
    | Made with ❤️ -
    <AccentLink href="https://github.com/goptics"> Goptics </AccentLink>
    © {{ new Date().getFullYear() }}
    <p class="text-muted-foreground/50">
      {{ version }}
    </p>
  </footer>
</template>
