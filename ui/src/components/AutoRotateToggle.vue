<script setup lang="ts">
import { Separator } from './ui'
import SettingsToggle from './SettingsToggle.vue'
import { useSettingsStore } from '../composables/useSettingsStore'
import { useActiveChartShape } from '../composables/useActiveChartShape'
import { useSyncedSetting } from '../composables/useSyncedSetting'

const { resolved, setForActiveChart } = useSettingsStore()
const { is3DChart } = useActiveChartShape()

const autoRotate = useSyncedSetting(
  () => resolved('autoRotate') as boolean,
  (val: boolean) => setForActiveChart({ autoRotate: val })
)
</script>

<template>
  <template v-if="is3DChart">
    <Separator />
    <SettingsToggle
      id="auto-rotate-switch"
      label="Auto rotate"
      description="Continuously rotate the 3D chart."
      :checked="autoRotate"
      @update:checked="autoRotate = $event"
    />
  </template>
</template>
