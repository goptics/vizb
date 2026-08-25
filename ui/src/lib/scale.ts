import type { ScaleAxis, ScaleInput, ScaleType } from '@/types'

export const DEFAULT_LOG_BASE = 10

const SCALE_AXES: ReadonlySet<string> = new Set(['x', 'y', 'z'])

export type ParsedScale = {
  type: ScaleType
  /** Explicit axes; `null` means omitted → use the chart's default value axis. */
  axes: ScaleAxis[] | null
  base: number
  baseX?: number
  baseY?: number
  baseZ?: number
}

const isScaleAxis = (value: unknown): value is ScaleAxis =>
  typeof value === 'string' && SCALE_AXES.has(value)

/** ECharts logBase must be finite and not 0 or 1. */
export function validLogBase(value: unknown): number | undefined {
  if (typeof value !== 'number' || !Number.isFinite(value)) return undefined
  if (value <= 0 || value === 1) return undefined
  return value
}

function parseAxes(axes: unknown): ScaleAxis[] | null {
  if (!Array.isArray(axes)) return null
  return axes.filter(isScaleAxis)
}

/** Normalize Dataset `scale` (string or object) without applying chart defaults. */
export function parseScale(input: ScaleInput | undefined | null): ParsedScale {
  if (input == null || typeof input !== 'object') {
    return { type: input === 'log' ? 'log' : 'linear', axes: null, base: DEFAULT_LOG_BASE }
  }

  return {
    type: input.type === 'log' ? 'log' : 'linear',
    axes: parseAxes(input.axes),
    base: validLogBase(input.base) ?? DEFAULT_LOG_BASE,
    baseX: validLogBase(input.baseX),
    baseY: validLogBase(input.baseY),
    baseZ: validLogBase(input.baseZ),
  }
}

export function axisIsLog(
  parsed: ParsedScale,
  axis: ScaleAxis,
  defaultAxes: readonly ScaleAxis[]
): boolean {
  if (parsed.type !== 'log') return false
  const axes = parsed.axes ?? defaultAxes
  return axes.includes(axis)
}

export function axisLogBase(parsed: ParsedScale, axis: ScaleAxis): number {
  const override = axis === 'x' ? parsed.baseX : axis === 'y' ? parsed.baseY : parsed.baseZ
  return override ?? parsed.base
}

/**
 * When X is log and every category label is a finite number > 0, return those
 * numbers so callers can coerce the category axis to a value/log axis.
 * Non-numeric or non-positive labels: stay category (ignore X in axes).
 */
export function numericLogXValues(labels: string[], xLog: boolean): number[] | null {
  if (!xLog || labels.length === 0) return null
  const nums: number[] = []
  for (const label of labels) {
    const n = Number(label)
    if (!Number.isFinite(n) || n <= 0) return null
    nums.push(n)
  }
  return nums
}

/** Pair numeric log-X categories with Y, nulling non-positive Y when Y is log. */
export function asLogXPairs(
  xNums: number[],
  values: (number | null)[],
  yLog: boolean
): [number, number | null][] {
  return xNums.map((x, i) => {
    const y = values[i] ?? null
    return [x, y === null || (yLog && y <= 0) ? null : y]
  })
}
