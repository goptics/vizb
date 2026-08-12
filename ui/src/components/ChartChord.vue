<script setup lang="ts">
import { ref } from 'vue'
import type { EChartsOption } from 'echarts'
import { use } from 'echarts/core'
import { ChordChart } from 'echarts/charts'
import VChart from 'vue-echarts'
import { BASE_2D } from './charts/base'
import {
  createLegendSelectChangedForwarder,
  type LegendSelectChangedEvent,
} from './charts/legendEvents'

// Reached only through a dynamic import (see ChartCard.vue), keeping the Chord
// renderer out of the default embedded chart bundle.
use([...BASE_2D, ChordChart])

const props = defineProps<{
  option: EChartsOption
  initOptions: Record<string, unknown>
}>()

const emit = defineEmits<{
  legendselectchanged: [e: LegendSelectChangedEvent]
}>()

const onLegendSelectChanged = createLegendSelectChangedForwarder((event) =>
  emit('legendselectchanged', event)
)

// ECharts chord gradients use absolute edge coords and go stale on resize
// (fullscreen). notMerge recreates edges so gradients recompute.
const chartRef = ref<{
  chart?: { isDisposed: () => boolean; setOption: (o: EChartsOption, opts?: object) => void }
} | null>(null)

const autoresize = {
  throttle: 100,
  onResize: () => {
    const chart = chartRef.value?.chart
    if (!chart || chart.isDisposed()) return
    chart.setOption(props.option, { notMerge: true })
  },
}
</script>

<template>
  <VChart
    ref="chartRef"
    :option="option"
    :init-options="initOptions"
    :autoresize="autoresize"
    @legendselectchanged="onLegendSelectChanged"
  />
</template>
