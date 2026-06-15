<script setup lang="ts">
import {
  toRefs,
  ref,
  shallowRef,
  computed,
  watch,
  nextTick,
  defineAsyncComponent,
  h,
  type Component,
} from 'vue'
import type { EChartsOption } from 'echarts'
import { Sigma } from 'lucide-vue-next'
import { useChartOptions } from '../composables/useChartOptions'
import type { ChartData, ChartType, Sort, ScaleType } from '../types'
import { useSettingsStore } from '../composables/useSettingsStore'
import { useFullscreen } from '../composables/useFullscreen'
import { is3D } from '../lib/utils'
import StatsPanel from './StatsPanel.vue'
import Badge from './Badge.vue'
import Button from './ui/Button.vue'

// Every chart renderer (2D bar/line/pie + 3D) is loaded via defineAsyncComponent
// so the echarts runtime stays out of the eager startup bundle: nothing in the
// initial parse imports echarts, and each chart's module body lands in its own
// chunk that the browser only parses when that type actually renders. The
// chart-area skeleton shows while a chunk loads (once per type, then cached).
const ChartLoading = () => h('div', { class: 'h-[600px] animate-pulse rounded bg-muted' })
const ChartLoadError = () =>
  h(
    'div',
    { class: 'flex h-[600px] items-center justify-center text-sm text-muted-foreground' },
    'Failed to load chart'
  )
const mk = (loader: () => Promise<Component>) =>
  defineAsyncComponent({
    loader,
    loadingComponent: ChartLoading,
    errorComponent: ChartLoadError,
    delay: 0,
  })

const RENDERERS: Record<ChartType, Component> = {
  bar: mk(() => import('./ChartBar.vue')),
  line: mk(() => import('./ChartLine.vue')),
  pie: mk(() => import('./ChartPie.vue')),
  heatmap: mk(() => import('./ChartHeatmap.vue')),
  radar: mk(() => import('./ChartRadar.vue')),
}
const Chart3D = mk(() => import('./Chart3D.vue'))

const props = defineProps<{
  chartData: ChartData
  // True while the transform worker is recomputing (sort/swap/group switch).
  // Shows the chart-area skeleton instead of a stale chart until it resolves.
  loading?: boolean
}>()

// Convert props to refs
const { chartData } = toRefs(props)

// Drives which renderer mounts; only the 3D branch loads echarts-gl.
const is3DChart = computed(() => is3D(chartData))

// Pull settings from centralized store
const { settings, chartType, resolved } = useSettingsStore()

// Pick the lazily-loaded renderer for the active chart shape/type. Pie has no
// 3D form (it renders per-dimension 2D pies even for x/y/z data), so it always
// routes to ChartPie — never Chart3D, which doesn't register the pie module.
const ActiveChart = computed<Component>(() => {
  // Pie, heatmap, and radar have no 3D form — each renders its own 2D layout even for
  // x/y/z data (pie: per-dimension pies; heatmap: z on legend; radar: per-dimension radars),
  // so they must route past the is3D check that otherwise hands x/y/z off to Chart3D.
  if (chartType.value === 'pie') return RENDERERS.pie
  if (chartType.value === 'heatmap') return RENDERERS.heatmap
  if (chartType.value === 'radar') return RENDERERS.radar
  return is3DChart.value ? Chart3D : (RENDERERS[chartType.value] ?? RENDERERS.bar)
})

const { isDark } = toRefs(settings)
// Per-chart resolved values — each chart type can override global defaults.
const sort = computed(() => resolved('sort') as Sort)
const showLabels = computed(() => resolved('showLabels') as boolean)
const scale = computed(() => resolved('scale') as ScaleType)
const autoRotate = computed(() => resolved('autoRotate') as boolean)

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
  visibleZ
)

const initOptions = {
  renderer: 'canvas',
  devicePixelRatio: window.devicePixelRatio,
} as const

const { containerRef, isFullscreen, withFullscreenToolbox } = useFullscreen()

// Stats panel is collapsed by default so the chart view is unchanged until the
// user opts in. Offered for any chart with at least one series; the actual
// profiles/correlation are computed lazily off-thread when the panel opens
// (see StatsPanel.vue + useStatsWorker.ts), so this check stays payload-free.
const showStats = ref(false)
const hasStats = computed(() => chartData.value.series.length > 0)

const mergedOptions = computed<EChartsOption>(() => withFullscreenToolbox(options.value))

// Double-buffer the chart so a worker recompute never flashes a stale or
// half-drawn frame. The live `<component>` renders `renderedChart`/`renderedOption`
// — a buffer we control — not `ActiveChart`/`mergedOptions` directly.
const renderedChart = shallowRef<Component>(ActiveChart.value)
const renderedOption = shallowRef<EChartsOption>(mergedOptions.value)
const showSkeleton = ref(!!props.loading)

// Pass-through for updates that are NOT a worker recompute (theme toggle, axis
// labels, fullscreen toolbox, legend z-toggle). They mutate mergedOptions/
// ActiveChart without `loading`, and must reach the chart immediately.
watch(mergedOptions, (o) => {
  if (!props.loading) renderedOption.value = o
})
watch(ActiveChart, (c) => {
  if (!props.loading) renderedChart.value = c
})

// Worker recompute (swap / sort / group switch). While `loading` we raise the
// skeleton and FREEZE the buffer on the old data underneath it — so nothing the
// chart shows changes while the overlay is up. When the new data lands we apply
// it to the buffer *behind* the still-raised overlay, wait for it to actually
// paint (nextTick = option handed to echarts; two rAF = canvas repainted), then
// drop the overlay. The new frame is therefore on screen before the reveal, so
// neither the old chart nor a half-drawn new one is ever visible. A shape flip
// (3D↔2D) swaps the renderer component; its own async loadingComponent — an
// identical skeleton — covers any chunk-load gap, so the swap is seamless too.
watch(
  () => props.loading,
  (l) => {
    if (l) {
      showSkeleton.value = true
      return
    }
    renderedChart.value = ActiveChart.value
    renderedOption.value = mergedOptions.value
    nextTick(() =>
      requestAnimationFrame(() => requestAnimationFrame(() => (showSkeleton.value = false)))
    )
  }
)
</script>

<template>
  <div
    ref="containerRef"
    class="rounded-lg border border-border bg-card p-6 shadow-sm transition-shadow hover:shadow-md"
    :class="{ 'fixed inset-0 z-50 rounded-none': isFullscreen }"
  >
    <div class="flex items-center justify-between gap-2 mb-2">
      <h3 class="text-lg font-semibold text-card-foreground">
        {{ chartData.title }}
      </h3>
      <div class="flex flex-wrap justify-end gap-1.5">
        <Badge
          :label="chartData.axisLabels?.x || 'Series'"
          :value="String(chartData.series.length)"
        />
        <Badge
          :label="chartData.axisLabels?.y || 'Y-axis'"
          :value="String(chartData.yAxis.length)"
        />
        <Badge
          v-if="is3DChart"
          :label="chartData.axisLabels?.z || 'Z-axis'"
          :value="String(chartData.zAxis.length)"
        />
      </div>
    </div>
    <!-- Keep the chart mounted and overlay the skeleton; unmounting would reset
         the echarts-gl camera on every 3D switch (zoom-in flash). -->
    <div class="relative" :class="isFullscreen ? 'h-[calc(100vh-4rem)]' : 'h-[600px]'">
      <component
        :is="renderedChart"
        :option="renderedOption"
        :init-options="initOptions"
        class="h-full w-full"
        @legendselectchanged="onLegendSelectChanged"
      />
      <div v-if="showSkeleton" class="absolute inset-0 z-10 animate-pulse rounded bg-muted" />
      <Button
        v-if="hasStats"
        class="absolute bottom-2 right-2 z-20 h-8 bg-transparent border border-border px-2.5 py-0 text-xs leading-none text-muted-foreground transition-colors hover:bg-accent hover:text-primary"
        :class="{ 'bg-accent text-primary': showStats }"
        :aria-pressed="showStats"
        title="Toggle statistics"
        @click="showStats = !showStats"
      >
        <template #icon>
          <Sigma class="h-4 w-4" />
        </template>
        Stats
      </Button>
    </div>
    <StatsPanel v-if="hasStats && showStats" :chart-data="chartData" />
  </div>
</template>
