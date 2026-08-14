import type { BaseChartConfig } from './baseChartOptions'
import { use3DChartOptions } from './use3DChartOptions'

export function useLine3DChartOptions(config: BaseChartConfig) {
  return use3DChartOptions(config, 'line3D')
}
