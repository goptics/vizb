<script setup lang="ts">
import type { EChartsOption } from 'echarts'
import { use } from 'echarts/core'
import { SankeyChart } from 'echarts/charts'
import VChart from 'vue-echarts'
import { BASE_2D } from './charts/base'

// Reached only through a dynamic import() (see ChartCard.vue). Sankey is a
// graph layout (no cartesian grid); chunk stays light like pie/radar.
use([...BASE_2D, SankeyChart])

defineProps<{
  option: EChartsOption
  initOptions: Record<string, unknown>
}>()

defineEmits<{
  legendselectchanged: [e: { selected: Record<string, boolean> }]
}>()
</script>

<template>
  <VChart
    :option="option"
    :init-options="initOptions"
    :autoresize="true"
    @legendselectchanged="$emit('legendselectchanged', $event)"
  />
</template>
