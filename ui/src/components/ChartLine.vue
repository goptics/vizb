<script setup lang="ts">
import type { EChartsOption } from 'echarts'
import { use } from 'echarts/core'
import { GridComponent } from 'echarts/components'
import { LineChart } from 'echarts/charts'
import VChart from 'vue-echarts'
import { BASE_2D } from './charts/base'
import {
  createLegendSelectChangedForwarder,
  type LegendSelectChangedEvent,
} from './charts/legendEvents'

// Reached only through a dynamic import() (see ChartCard.vue), so the LineChart
// module lands in its own chunk and is parsed only when a line chart renders.
use([...BASE_2D, GridComponent, LineChart])

defineProps<{
  option: EChartsOption
  initOptions: Record<string, unknown>
}>()

const emit = defineEmits<{
  legendselectchanged: [e: LegendSelectChangedEvent]
}>()

const onLegendSelectChanged = createLegendSelectChangedForwarder((event) =>
  emit('legendselectchanged', event)
)
</script>

<template>
  <VChart
    :option="option"
    :init-options="initOptions"
    :autoresize="true"
    @legendselectchanged="onLegendSelectChanged"
  />
</template>
