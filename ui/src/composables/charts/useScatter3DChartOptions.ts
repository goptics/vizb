import type { BaseChartConfig } from './baseChartOptions'
import { use3DChartOptions } from './use3DChartOptions'

export function useScatter3DChartOptions(config: BaseChartConfig) {
  return use3DChartOptions(config, 'scatter3D')
}
