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
  DataZoomComponent
} from 'echarts/components'
import VChart from 'vue-echarts'
import { useEChartOptions } from '../composables/useEChartOptions'
import type { ChartData } from '../types/benchmark'
import type { SortOrder } from '../types/benchmark'
import type { ChartType } from '../types/benchmark'

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
  DataZoomComponent
])

const props = defineProps<{
  chartData: ChartData
  sortOrder: SortOrder
  showLabels: boolean
  isDark: boolean
  chartType: ChartType
}>()

// Convert props to refs and pass them directly to maintain reactivity
const { chartData, sortOrder, showLabels, isDark, chartType } = toRefs(props)

const { options } = useEChartOptions(
  chartData,
  sortOrder,
  showLabels,
  isDark,
  chartType
)
</script>

<template>
  <div class="bg-card border border-border rounded-lg p-6 shadow-sm hover:shadow-md transition-shadow">
    <h3 class="text-lg font-semibold text-card-foreground mb-4">
      {{ chartData.title }}
    </h3>
    <div class="w-full h-[500px]">
      <VChart
        :option="options"
        :autoresize="true"
        class="w-full h-full"
      />
    </div>
  </div>
</template>

<style scoped>
@media (max-width: 768px) {
  .h-\[500px\] {
    height: 350px;
  }
}
</style>
