export const DEFAULT_FONT_SIZE = 12

export type FontSizeSpec = {
  series?: number
  legend?: number
  label?: number
}

export type ResolvedFontSize = {
  series: number
  legend: number
  label: number
}

function validSize(n: unknown): n is number {
  return typeof n === 'number' && Number.isFinite(n) && n > 0
}

export function resolveFontSize(input?: number | FontSizeSpec | null): ResolvedFontSize {
  if (typeof input === 'number' && validSize(input)) {
    return { series: input, legend: input, label: input }
  }
  const spec = input && typeof input === 'object' ? input : {}
  return {
    series: validSize(spec.series) ? spec.series : DEFAULT_FONT_SIZE,
    legend: validSize(spec.legend) ? spec.legend : DEFAULT_FONT_SIZE,
    label: validSize(spec.label) ? spec.label : DEFAULT_FONT_SIZE,
  }
}
