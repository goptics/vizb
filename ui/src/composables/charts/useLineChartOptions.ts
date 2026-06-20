import { type BaseChartConfig } from './baseChartOptions'
import { useCategorySeriesChartOptions } from './useCategorySeriesChartOptions'

export function useLineChartOptions(config: BaseChartConfig) {
  return useCategorySeriesChartOptions(config, 'line')
}
