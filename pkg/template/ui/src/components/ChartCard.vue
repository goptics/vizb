<script setup lang="ts">
import { toRefs } from 'vue'
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
const { sortOrder, showLabels, isDark, chartType } = (() => {
  const store = useSettingsStore()
  return {
    sortOrder: store.sortOrder,
    showLabels: store.showLabels,
    isDark: store.isDark,
    chartType: store.chartType,
  }
})()

const { options } = useChartOptions(chartData, sortOrder, showLabels, isDark, chartType)

const initOptions = {
  renderer: 'canvas',
  devicePixelRatio: window.devicePixelRatio,
} as const
</script>

<template>
  <div
    class="rounded-lg border border-border bg-card p-6 shadow-sm transition-shadow hover:shadow-md"
  >
    <h3 class="text-lg font-semibold text-card-foreground">
      {{ chartData.title }}
    </h3>
    <VChart :option="options" :init-options="initOptions" :autoresize="true" class="h-[500px]" />
  </div>
</template>
