import { computed } from 'vue'
import { useBenchmarkData } from './useBenchmarkData'
import { useSettingsStore } from './useSettingsStore'

export function useActiveChartShape() {
  const { activeBenchmark } = useBenchmarkData()
  const { settings } = useSettingsStore()

  const is3DChart = computed(() => {
    const data = activeBenchmark.value?.data
    if (!data) return false
    return data.some((d) => d.xAxis) && data.some((d) => d.yAxis) && data.some((d) => d.zAxis)
  })

  const isAxisChart = computed(
    () => (settings.charts[settings.activeChartIndex] ?? 'bar') !== 'pie' && !is3DChart.value
  )

  return { is3DChart, isAxisChart }
}
