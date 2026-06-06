<script setup lang="ts">
import { toRefs, ref, computed } from 'vue'
import type { EChartsOption } from 'echarts'
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
const { sort, showLabels, isDark, scale, autoRotate } = toRefs(settings)

// Legend z-toggle state, kept in sync via the legendselectchanged event so
// tooltip/label sums reflect only the visible z series.
const visibleZ = ref<Record<string, boolean>>({})
function onLegendSelectChanged(e: { selected: Record<string, boolean> }) {
  visibleZ.value = { ...e.selected }
}

const { options } = useChartOptions(
  chartData,
  sort,
  showLabels,
  isDark,
  chartType,
  scale,
  autoRotate,
  visibleZ,
)

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

// Line-style corner-bracket icons so they match echarts' stroke-only toolbox
// icons (filled glyphs render hollow). Enter = outward corners, exit = inward.
const ENTER_FULLSCREEN_ICON = 'path://M3 9V3H9 M21 9V3H15 M3 15V21H9 M21 15V21H15'
const EXIT_FULLSCREEN_ICON = 'path://M9 3V9H3 M15 3V9H21 M9 21V15H3 M15 21V15H21'

// Inject fullscreen as a custom toolbox feature so it sits inline with
// saveAsImage in echarts' horizontal toolbar (instead of a separate button).
const mergedOptions = computed<EChartsOption>(() => {
  const opt = options.value
  const toolbox = opt.toolbox as Record<string, unknown> | undefined
  const feature = (toolbox?.feature ?? {}) as Record<string, unknown>
  return {
    ...opt,
    toolbox: {
      ...toolbox,
      feature: {
        ...feature,
        myFullScreen: {
          show: true,
          title: isFullscreen.value ? 'Exit fullscreen' : 'Fullscreen',
          icon: isFullscreen.value ? EXIT_FULLSCREEN_ICON : ENTER_FULLSCREEN_ICON,
          onclick: toggleFullscreen,
        },
      },
    },
  } as EChartsOption
})
</script>

<template>
  <div
    ref="containerRef"
    class="rounded-lg border border-border bg-card p-6 shadow-sm transition-shadow hover:shadow-md"
    :class="{ 'fixed inset-0 z-50 rounded-none': isFullscreen }"
  >
    <h3 class="text-lg font-semibold text-card-foreground">
      {{ chartData.title }}
    </h3>
    <VChart
      :option="mergedOptions"
      :init-options="initOptions"
      :autoresize="true"
      :class="isFullscreen ? 'h-[calc(100vh-4rem)]' : 'h-[500px]'"
      @legendselectchanged="onLegendSelectChanged"
    />
  </div>
</template>
