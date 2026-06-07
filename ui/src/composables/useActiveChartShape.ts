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

  const isAxisChart = computed(
    () => (settings.charts[settings.activeChartIndex] ?? 'bar') !== 'pie' && !is3DChart.value
  )

  return { is3DChart, isAxisChart }
}
