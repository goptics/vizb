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
  VisualMapComponent,
} from 'echarts/components'
import { Bar3DChart, Line3DChart, Scatter3DChart } from 'echarts-gl/charts'
import { Grid3DComponent } from 'echarts-gl/components'
import VChart from 'vue-echarts'
import {
  createLegendSelectChangedForwarder,
  type LegendSelectChangedEvent,
} from './charts/legendEvents'

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
  VisualMapComponent,
  Bar3DChart,
  Line3DChart,
  Scatter3DChart,
  Grid3DComponent,
])

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
  <!--
    `:update-options="{ notMerge: false }"` fixes the autoRotate lag and makes
    the 3D update contract explicit. vue-echarts 8 normally plans option
    updates automatically, but a replacement update tears down + rebuilds the
    3D scene (re-uploads bar/line geometry, re-binds lights, and restarts the
    view-control animation state).

    Overriding to `notMerge: false` makes ECharts MERGE the new option into
    the existing one. Incremental changes — a single field like
    `grid3D.viewControl.autoRotate` — just patch in place; the ViewGL flips
    the rotation animation without rebuilding the scene.

    Heavy changes are still handled correctly: `replaceMerge: ['series']`
    replaces series wholesale on dataset swaps, so a new dataset doesn't leave
    stale series behind. `visualMap` is also replace-merged so toggling the
    3D visual-map setting off clears the gradient controller (merge alone
    would keep the previous visualMap mounted).
  -->
  <VChart
    :option="option"
    :init-options="initOptions"
    :update-options="{ notMerge: false, replaceMerge: ['series', 'visualMap'] }"
    :autoresize="true"
    @legendselectchanged="onLegendSelectChanged"
  />
</template>
