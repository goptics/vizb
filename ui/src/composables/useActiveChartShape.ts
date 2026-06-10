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

  // Scale (linear/log) applies to any value-axis chart — 2D and 3D bar/line.
  // Only pie has no value axis to scale.
  const isAxisChart = computed(
    () => (settings.charts[settings.activeChartIndex] ?? 'bar') !== 'pie'
  )

  return { is3DChart, isAxisChart }
}
