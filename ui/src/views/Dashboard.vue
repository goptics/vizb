<script setup lang="ts">
import { computed } from 'vue'
import { Moon, Sun, Package } from 'lucide-vue-next'
import { useBenchmarkData } from '../composables/useBenchmarkData'
import { useChartData } from '../composables/useChartData'
import { useSettingsStore } from '../composables/useSettingsStore'
import { useDashboardInit } from '../composables/useDashboardInit'
import ChartSettingsPopover from '../components/ChartSettingsPopover.vue'
import ChartCard from '../components/ChartCard.vue'
import BenchmarkHeader from '../components/BenchmarkHeader.vue'
import LoadingSkeleton from '../components/LoadingSkeleton.vue'
import LoadError from '../components/LoadError.vue'
import AppFooter from '../components/AppFooter.vue'
import IconButton from '../components/IconButton.vue'

const version = window.VIZB_VERSION || 'v0.0.0-dev'

const {
  benchmarks,
  activeBenchmark,
  activeBenchmarkId,
  selectBenchmark,
  resultGroups,
  activeGroup,
  activeGroupId,
  selectGroup,
  loading,
  loadError,
} = useBenchmarkData()

const activeResults = computed(() => activeGroup.value?.data || [])
const { chartData } = useChartData(activeResults)

const { settings, toggleDark } = useSettingsStore()

useDashboardInit()
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

  <LoadingSkeleton v-if="loading" />

  <LoadError v-else-if="loadError" :message="loadError" />

  <main v-else-if="activeBenchmark" class="mx-auto min-h-screen max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
    <BenchmarkHeader
      :benchmark="activeBenchmark"
      :benchmarks="benchmarks"
      :activeBenchmarkId="activeBenchmarkId"
      :resultGroups="resultGroups"
      :activeGroupId="activeGroupId"
      @selectBenchmark="selectBenchmark"
      @selectGroup="selectGroup"
    />

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

  <AppFooter v-if="activeBenchmark && !loading && !loadError" :version="version" />
</template>
