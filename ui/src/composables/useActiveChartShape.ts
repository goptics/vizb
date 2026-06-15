import { computed } from 'vue'
import { useDataPoint } from './useDataPoint'
import { useSettingsStore } from './useSettingsStore'

export function useActiveChartShape() {
  const { activeDataSet } = useDataPoint()
  const { settings } = useSettingsStore()

  const is3DChart = computed(() => {
    const data = activeDataSet.value?.data
    if (!data) return false
    return data.some((d) => d.xAxis) && data.some((d) => d.yAxis) && data.some((d) => d.zAxis)
  })

  // Scale (linear/log) applies to bar/line (2D and 3D). Not valid for pie or heatmap.
  const isAxisChart = computed(() => {
    const ct = settings.charts[settings.activeChartIndex] ?? 'bar'
    return ct !== 'pie' && ct !== 'heatmap'
  })

  return { is3DChart, isAxisChart }
}
