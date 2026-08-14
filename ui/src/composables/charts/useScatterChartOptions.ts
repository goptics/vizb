import { type BaseChartConfig } from './baseChartOptions'
import { useCategorySeriesChartOptions } from './useCategorySeriesChartOptions'

export function useScatterChartOptions(config: BaseChartConfig) {
  return useCategorySeriesChartOptions(config, 'scatter')
}
