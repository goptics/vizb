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
import { useChartOptions } from '../composables/useChartOptions'
import type { ChartData, ChartType } from '../types'
import { useSettingsStore } from '../composables/useSettingsStore'
import { is3D } from '../lib/utils'

// Every chart renderer (2D bar/line/pie + 3D) is loaded via defineAsyncComponent
// so the echarts runtime stays out of the eager startup bundle: nothing in the
// initial parse imports echarts, and each chart's module body lands in its own
// chunk that the browser only parses when that type actually renders. The
// chart-area skeleton shows while a chunk loads (once per type, then cached).
const ChartLoading = () => h('div', { class: 'h-[500px] animate-pulse rounded bg-muted' })
const ChartLoadError = () =>
  h(
    'div',
    { class: 'flex h-[500px] items-center justify-center text-sm text-muted-foreground' },
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
const { settings, chartType } = useSettingsStore()

// Pick the lazily-loaded renderer for the active chart shape/type. Pie has no
// 3D form (it renders per-dimension 2D pies even for x/y/z data), so it always
// routes to ChartPie — never Chart3D, which doesn't register the pie module.
const ActiveChart = computed<Component>(() => {
  if (chartType.value === 'pie') return RENDERERS.pie
  return is3DChart.value ? Chart3D : (RENDERERS[chartType.value] ?? RENDERERS.bar)
})

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
  visibleZ
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
    <h3 class="text-lg font-semibold text-card-foreground">
      {{ chartData.title }}
    </h3>
    <!-- Keep the chart mounted and overlay the skeleton; unmounting would reset
         the echarts-gl camera on every 3D switch (zoom-in flash). -->
    <div class="relative" :class="isFullscreen ? 'h-[calc(100vh-4rem)]' : 'h-[500px]'">
      <component
        :is="renderedChart"
        :option="renderedOption"
        :init-options="initOptions"
        class="h-full w-full"
        @legendselectchanged="onLegendSelectChanged"
      />
      <div v-if="showSkeleton" class="absolute inset-0 z-10 animate-pulse rounded bg-muted" />
    </div>
  </div>
</template>
