<script setup lang="ts">
import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { use } from 'echarts/core'
import { GridComponent, VisualMapComponent } from 'echarts/components'
import { HeatmapChart } from 'echarts/charts'
import VChart from 'vue-echarts'
import { BASE_2D } from './charts/base'
import { useFullscreen } from '../composables/useFullscreen'

// Reached only through a dynamic import() (see StatsPanel.vue), so the heatmap +
// visualMap echarts modules land in their own chunk and are parsed only when the
// correlation tab actually renders.
use([...BASE_2D, GridComponent, VisualMapComponent, HeatmapChart])

const props = defineProps<{
  option: EChartsOption
  initOptions: Record<string, unknown>
}>()

const { containerRef, isFullscreen, withFullscreenToolbox } = useFullscreen()
const mergedOption = computed<EChartsOption>(() => withFullscreenToolbox(props.option))
</script>

<template>
  <div
    ref="containerRef"
    class="h-full"
    :class="{ 'fixed inset-0 z-50 bg-background': isFullscreen }"
  >
    <VChart :option="mergedOption" :init-options="initOptions" :autoresize="true" class="h-full w-full" />
  </div>
</template>
