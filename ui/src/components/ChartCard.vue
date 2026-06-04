<script setup lang="ts">
import { toRefs, ref } from 'vue'
import { Maximize2, Minimize2 } from 'lucide-vue-next'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart, LineChart, PieChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  ToolboxComponent,
} from 'echarts/components'
import { Bar3DChart, Line3DChart, Scatter3DChart } from 'echarts-gl/charts'
import { Grid3DComponent } from 'echarts-gl/components'
import VChart from 'vue-echarts'
import { useChartOptions } from '../composables/useChartOptions'
import type { ChartData } from '../types'
import { useSettingsStore } from '../composables/useSettingsStore'

// Register ECharts components
use([
  CanvasRenderer,
  BarChart,
  LineChart,
  PieChart,
  Bar3DChart,
  Line3DChart,
  Scatter3DChart,
  Grid3DComponent,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  ToolboxComponent,
])

const props = defineProps<{
  chartData: ChartData
}>()

// Convert props to refs
const { chartData } = toRefs(props)

// Pull settings from centralized store
const { settings, chartType } = useSettingsStore()
const { sort, showLabels, isDark, scale } = toRefs(settings)

const { options } = useChartOptions(chartData, sort, showLabels, isDark, chartType, scale)

const initOptions = {
  renderer: 'canvas',
  devicePixelRatio: window.devicePixelRatio,
} as const

const containerRef = ref<HTMLElement | null>(null)
const isFullscreen = ref(false)

function toggleFullscreen() {
  if (!containerRef.value) return
  if (!document.fullscreenElement) {
    containerRef.value.requestFullscreen()
  } else {
    document.exitFullscreen()
  }
}

document.addEventListener('fullscreenchange', () => {
  isFullscreen.value = !!document.fullscreenElement
})
</script>

<template>
  <div
    ref="containerRef"
    class="rounded-lg border border-border bg-card p-6 shadow-sm transition-shadow hover:shadow-md"
    :class="{ 'fixed inset-0 z-50 rounded-none': isFullscreen }"
  >
    <div class="flex items-center justify-between">
      <h3 class="text-lg font-semibold text-card-foreground">
        {{ chartData.title }}
      </h3>
      <button
        class="rounded p-1 text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
        :title="isFullscreen ? 'Exit fullscreen' : 'Fullscreen'"
        @click="toggleFullscreen"
      >
        <Minimize2 v-if="isFullscreen" class="h-4 w-4" />
        <Maximize2 v-else class="h-4 w-4" />
      </button>
    </div>
    <VChart
      :option="options"
      :init-options="initOptions"
      :autoresize="true"
      :class="isFullscreen ? 'h-[calc(100vh-4rem)]' : 'h-[500px]'"
    />
  </div>
</template>
