import type { BaseChartConfig } from './baseChartOptions'
import { use3DChartOptions } from './use3DChartOptions'

export function useBar3DChartOptions(config: BaseChartConfig) {
  return use3DChartOptions(config, 'bar3D')
}
