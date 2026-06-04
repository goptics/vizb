<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Separator } from './ui'
import SettingsToggle from './SettingsToggle.vue'
import { useSettingsStore } from '../composables/useSettingsStore'
import { useBenchmarkData } from '../composables/useBenchmarkData'

const { settings, setAutoRotate } = useSettingsStore()
const { activeBenchmark } = useBenchmarkData()

const autoRotate = ref(settings.autoRotate)

// Auto-rotate only applies to 3D charts (x+y+z all present).
const is3DChart = computed(() => {
  const data = activeBenchmark.value?.data
  if (!data) return false
  return data.some((d) => d.xAxis) && data.some((d) => d.yAxis) && data.some((d) => d.zAxis)
})

watch(autoRotate, (val) => setAutoRotate(val))

const handleChange = (checked: boolean) => {
  autoRotate.value = checked
}
</script>

<template>
  <template v-if="is3DChart">
    <Separator />
    <SettingsToggle
      id="auto-rotate-switch"
      label="Auto rotate"
      description="Continuously rotate the 3D chart."
      :checked="autoRotate"
      @update:checked="handleChange"
    />
  </template>
</template>
