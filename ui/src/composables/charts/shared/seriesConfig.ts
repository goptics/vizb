import { fontSize } from './common'

export type SeriesSymbolProps = {
  symbol?: string
  symbolSize?: number
  sampling?: 'lttb'
}

/** Merge CLI --symbol / --symbol-size overrides onto chart defaults. */
export function resolveSeriesSymbol(
  defaults: SeriesSymbolProps,
  symbol?: string,
  symbolSize?: number
): SeriesSymbolProps {
  const out: SeriesSymbolProps = { ...defaults }
  if (symbol) out.symbol = symbol
  if (symbolSize !== undefined) out.symbolSize = symbolSize
  return out
}

export function resolve3DSymbolSize(
  computed: number | undefined,
  override?: number
): number | undefined {
  if (override !== undefined) return override
  return computed
}

export function resolve3DSymbolProps(
  computedSize: number | undefined,
  symbol?: string,
  symbolSize?: number
): { symbol?: string; symbolSize?: number } {
  const out: { symbol?: string; symbolSize?: number } = {}
  if (symbol) out.symbol = symbol
  const size = resolve3DSymbolSize(computedSize, symbolSize)
  if (size !== undefined) out.symbolSize = size
  return out
}

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
