<script setup lang="ts">
import type { EChartsOption } from 'echarts'
import { use } from 'echarts/core'
import { GridComponent, VisualMapComponent } from 'echarts/components'
import { HeatmapChart } from 'echarts/charts'
import VChart from 'vue-echarts'
import { BASE_2D } from './charts/base'

// Reached only through a dynamic import() (see StatsPanel.vue), so the heatmap +
// visualMap echarts modules land in their own chunk and are parsed only when the
// correlation tab actually renders.
use([...BASE_2D, GridComponent, VisualMapComponent, HeatmapChart])

defineProps<{
  option: EChartsOption
  initOptions: Record<string, unknown>
}>()
</script>

<template>
  <VChart :option="option" :init-options="initOptions" :autoresize="true" />
</template>
