<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { BarChart3, TrendingUp } from 'lucide-vue-next'
import { Separator } from './ui'
import SelectionTabs from './SelectionTabs.vue'
import type { ScaleType } from '../types'
import { useSettingsStore } from '../composables/useSettingsStore'

const { settings, setScale } = useSettingsStore()

const scaleType = ref<ScaleType>(settings.scale)
const isAxisChart = computed(() => (settings.charts[settings.activeChartIndex] ?? 'bar') !== 'pie')

watch(scaleType, (val) => setScale(val))

const scaleOptions = [
  { value: 'linear', label: 'Linear', icon: BarChart3 },
  { value: 'log', label: 'Logarithmic', icon: TrendingUp },
]

const handleScaleChange = (value: string | number) => {
  scaleType.value = String(value) as ScaleType
}
</script>

<template>
  <Separator v-if="isAxisChart" />
  <div v-if="isAxisChart" class="space-y-3">
    <p class="text-sm font-medium">Y-Axis Scale</p>
    <SelectionTabs
      :model-value="scaleType"
      :options="scaleOptions"
      @update:model-value="handleScaleChange"
    />
  </div>
</template>