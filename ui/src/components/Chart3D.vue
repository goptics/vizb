<script setup lang="ts">
import type { EChartsOption } from 'echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  ToolboxComponent,
} from 'echarts/components'
import { Bar3DChart, Line3DChart, Scatter3DChart } from 'echarts-gl/charts'
import { Grid3DComponent } from 'echarts-gl/components'
import VChart from 'vue-echarts'

// This component owns every echarts-gl import. Because it is only ever reached
// through a dynamic import() (see ChartCard.vue), the gl engine lands in its own
// rollup chunk and is parsed/compiled by the browser only when a 3D chart is
// actually rendered — keeping it off the 2D-only startup path.
use([
  CanvasRenderer,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  ToolboxComponent,
  Bar3DChart,
  Line3DChart,
  Scatter3DChart,
  Grid3DComponent,
])

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
