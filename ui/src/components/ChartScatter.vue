<script setup lang="ts">
import type { EChartsOption } from 'echarts'
import { use } from 'echarts/core'
import { GridComponent, VisualMapComponent } from 'echarts/components'
import { ScatterChart } from 'echarts/charts'
import VChart from 'vue-echarts'
import { BASE_2D } from './charts/base'

// Reached only through a dynamic import() (see ChartCard.vue), so the ScatterChart
// module lands in its own chunk and is parsed only when a scatter chart renders.
use([...BASE_2D, GridComponent, VisualMapComponent, ScatterChart])

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
    :update-options="{ notMerge: false, replaceMerge: ['visualMap'] }"
    @legendselectchanged="$emit('legendselectchanged', $event)"
  />
</template>
