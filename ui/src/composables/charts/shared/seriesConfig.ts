import { fontSize } from './common'
import type { ChartStyling } from './chartConfig'
import { createLabelConfig } from './chartConfig'

export function createPieLabelConfig(
  showLabels: boolean,
  styling: { textColor: string },
  customFormatter?: (params: any) => string
): any {
  return {
    show: showLabels,
    formatter: customFormatter,
    fontSize,
    color: styling.textColor,
  }
}

export function createPieSeriesConfig(
  name: string,
  data: any[],
  showLabels: boolean,
  styling: { textColor: string },
  customFormatter?: (params: any) => string,
  radius: [string, string] = ['40%', '70%'],
  center: [string, string] = ['50%', '50%']
): any {
  return {
    name,
    type: 'pie',
    radius,
    center,
    data,
    label: createPieLabelConfig(showLabels, styling, customFormatter),
  }
}

// Converts a nullable value into an ECharts data item with label config.
// Returns null to produce a gap (log-scale zero/negative handling).
export function makeDataItem(
  val: number | null,
  showLabels: boolean,
  styling: ChartStyling
): { value: number; label: ReturnType<typeof createLabelConfig> } | null {
  return val === null ? null : { value: val, label: createLabelConfig(showLabels, styling) }
}
