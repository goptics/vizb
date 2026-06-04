<script setup lang="ts">
import { BarChart3, TrendingUp } from 'lucide-vue-next'
import { Separator } from './ui'
import SelectionTabs from './SelectionTabs.vue'
import type { ScaleType } from '../types'
import { useSettingsStore } from '../composables/useSettingsStore'
import { useActiveChartShape } from '../composables/useActiveChartShape'
import { useSyncedSetting } from '../composables/useSyncedSetting'

const { settings, setScale } = useSettingsStore()
const { isAxisChart } = useActiveChartShape()

const scaleType = useSyncedSetting<ScaleType>(
  () => settings.scale,
  (val) => setScale(val)
)

const scaleOptions = [
  { value: 'linear', label: 'Linear', icon: BarChart3 },
  { value: 'log', label: 'Logarithmic', icon: TrendingUp },
]
</script>

<template>
  <Separator v-if="isAxisChart" />
  <div v-if="isAxisChart" class="space-y-3">
    <p class="text-sm font-medium">Data Scale</p>
    <SelectionTabs v-model="scaleType" :options="scaleOptions" />
  </div>
</template>
