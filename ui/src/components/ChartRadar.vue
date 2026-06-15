<script setup lang="ts">
import type { EChartsOption } from 'echarts'
import { use } from 'echarts/core'
import { RadarChart } from 'echarts/charts'
import { RadarComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { BASE_2D } from './charts/base'

use([...BASE_2D, RadarComponent, RadarChart])

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
