<script setup lang="ts">
import type { EChartsOption } from 'echarts'
import { use } from 'echarts/core'
import { GridComponent, VisualMapComponent } from 'echarts/components'
import { ScatterChart } from 'echarts/charts'
import VChart from 'vue-echarts'
import { BASE_2D } from './charts/base'
import {
  createLegendSelectChangedForwarder,
  type LegendSelectChangedEvent,
} from './charts/legendEvents'
import { createSeriesHover } from './charts/seriesHover'
import { ref } from 'vue'

// Reached only through a dynamic import() (see ChartCard.vue), so the ScatterChart
// module lands in its own chunk and is parsed only when a scatter chart renders.
use([...BASE_2D, GridComponent, VisualMapComponent, ScatterChart])

defineProps<{
  option: EChartsOption
  initOptions: Record<string, unknown>
}>()

const emit = defineEmits<{
  legendselectchanged: [e: LegendSelectChangedEvent]
}>()

const chart = ref<InstanceType<typeof VChart> | null>(null)
const { onMouseover, onMouseout } = createSeriesHover(() => chart.value?.chart)

const onLegendSelectChanged = createLegendSelectChangedForwarder((event) =>
  emit('legendselectchanged', event)
)
</script>

<template>
  <VChart
    ref="chart"
    :option="option"
    :init-options="initOptions"
    :autoresize="true"
    :update-options="{ notMerge: false, replaceMerge: ['series', 'visualMap'] }"
    @legendselectchanged="onLegendSelectChanged"
    @mouseover="onMouseover"
    @mouseout="onMouseout"
  />
</template>
