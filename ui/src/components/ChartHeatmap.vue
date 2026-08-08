<script setup lang="ts">
import type { EChartsOption } from 'echarts'
import { use } from 'echarts/core'
import { GridComponent, VisualMapComponent } from 'echarts/components'
import { HeatmapChart, ScatterChart } from 'echarts/charts'
import VChart from 'vue-echarts'
import { BASE_2D } from './charts/base'
import {
  createLegendSelectChangedForwarder,
  type LegendSelectChangedEvent,
} from './charts/legendEvents'

use([...BASE_2D, GridComponent, VisualMapComponent, HeatmapChart, ScatterChart])

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
