<script setup lang="ts">
import type { EChartsOption } from 'echarts'
import { use } from 'echarts/core'
import { PieChart } from 'echarts/charts'
import VChart from 'vue-echarts'
import { BASE_2D } from './charts/base'

// Reached only through a dynamic import() (see ChartCard.vue). Pie needs no grid
// or cartesian coord system, so this chunk stays the lightest of the three.
use([...BASE_2D, PieChart])

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
