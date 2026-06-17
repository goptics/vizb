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
  <!--
    `:update-options="{ notMerge: false }"` is the toggle that fixes the
    autoRotate lag. vue-echarts defaults to
    `setOption(option, { notMerge: option !== oldOption, ... })`, so every
    time our options computed produces a new reference (theme, sort, scale,
    autoRotate, anything), ECharts receives `notMerge: true` and the
    3D scene is torn down + rebuilt (re-uploads bar/line geometry, re-binds
    lights, restarts the view-control animation state). That's the lag.

    Overriding to `notMerge: false` makes ECharts MERGE the new option into
    the existing one. Incremental changes — a single field like
    `grid3D.viewControl.autoRotate` — just patch in place; the ViewGL flips
    the rotation animation without rebuilding the scene.

    Heavy changes are still handled correctly: ECharts' default
    `replaceMerge: ['series']` replaces series wholesale on dataset swaps,
    so a new dataset doesn't leave stale series behind. The merge vs.
    replace is decided per top-level component by `replaceMerge`, not by us.
  -->
  <VChart
    :option="option"
    :init-options="initOptions"
    :update-options="{ notMerge: false }"
    :autoresize="true"
    @legendselectchanged="$emit('legendselectchanged', $event)"
  />
</template>
