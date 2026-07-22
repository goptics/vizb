<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'
import type { EChartsOption } from 'echarts'
import { use } from 'echarts/core'
import { BrushComponent, GridComponent } from 'echarts/components'
import { BarChart } from 'echarts/charts'
import VChart from 'vue-echarts'
import { brushSelectionStats, type BrushSelectionStats } from '../lib/brushSelection'
import { BASE_2D } from './charts/base'

// Reached only through a dynamic import() (see ChartCard.vue), so the BarChart
// module lands in its own chunk and is parsed only when a bar chart renders.
use([...BASE_2D, BrushComponent, GridComponent, BarChart])

const props = defineProps<{
  option: EChartsOption
  initOptions: Record<string, unknown>
}>()
const chartRef = ref<InstanceType<typeof VChart>>()

const emit = defineEmits<{
  legendselectchanged: [e: { selected: Record<string, boolean> }]
  brushselected: [stats: BrushSelectionStats]
}>()

function onBrushSelected(event: Parameters<typeof brushSelectionStats>[1]) {
  emit('brushselected', brushSelectionStats(props.option, event))
}

watch(
  () => props.option,
  () => nextTick(() => chartRef.value?.dispatchAction({ type: 'brush', areas: [] }))
)
</script>

<template>
  <VChart
    ref="chartRef"
    :option="option"
    :init-options="initOptions"
    :autoresize="true"
    @legendselectchanged="$emit('legendselectchanged', $event)"
    @brushselected="onBrushSelected"
  />
</template>
