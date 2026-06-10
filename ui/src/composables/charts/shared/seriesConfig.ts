import { fontSize } from './common'

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
