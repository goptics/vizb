import { create3DVisualMap } from './3d'
import type { ChartStyling } from './chartConfig'

export function maxFromScatterValues(values: number[]): number {
  let max = 0
  for (const v of values) {
    if (v > max) max = v
  }
  return max || 1
}

/** ECharts merge keeps omitted visualMap — pass `[]` when off (with replaceMerge). */
export function resolve2DScatterVisualMap(
  enabled: boolean,
  values: number[],
  styling: ChartStyling,
  dimension: 1 | 2 = 2
) {
  if (!enabled || values.length === 0) return []
  return create3DVisualMap(maxFromScatterValues(values), styling, dimension)
}
