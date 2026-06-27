import { VALUE_3D_COLOR_RANGE } from './3d'
import type { ChartStyling } from './chartConfig'

export function maxFromScatterValues(values: number[]): number {
  let max = 0
  for (const v of values) {
    if (v > max) max = v
  }
  return max || 1
}

export function createScatterVisualMap(
  max: number,
  styling: ChartStyling,
  dimension: 0 | 1 | 2 = 2
) {
  return {
    show: true,
    min: 0,
    max: max || 1,
    dimension,
    calculable: true,
    orient: 'vertical' as const,
    right: '0%',
    top: 'center',
    inRange: { color: VALUE_3D_COLOR_RANGE },
    textStyle: { color: styling.textColor },
  }
}

/** ECharts merge keeps omitted visualMap — pass `[]` when off (with replaceMerge). */
export function resolve2DScatterVisualMap(
  enabled: boolean,
  values: number[],
  styling: ChartStyling,
  dimension: 0 | 1 | 2 = 2
) {
  if (!enabled || values.length === 0) return []
  return createScatterVisualMap(maxFromScatterValues(values), styling, dimension)
}
