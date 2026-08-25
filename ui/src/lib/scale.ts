import type { ScaleAxis, ScaleInput, ScaleType } from '@/types'

export const DEFAULT_LOG_BASE = 10

export const DEFAULT_LOG_AXES = {
  grouped2d: ['y'] as const satisfies readonly ScaleAxis[],
  horizontalBar: ['x'] as const satisfies readonly ScaleAxis[],
  value2d: ['y'] as const satisfies readonly ScaleAxis[],
  mixed2d: ['y'] as const satisfies readonly ScaleAxis[],
  grouped3d: ['z'] as const satisfies readonly ScaleAxis[],
  value3d: ['z'] as const satisfies readonly ScaleAxis[],
  mixed3d: ['y', 'z'] as const satisfies readonly ScaleAxis[],
  continuous3d: ['x', 'y', 'z'] as const satisfies readonly ScaleAxis[],
}

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
  if (input == null || input === 'linear') {
    return { type: 'linear', axes: null, base: DEFAULT_LOG_BASE }
  }
  if (input === 'log') {
    return { type: 'log', axes: null, base: DEFAULT_LOG_BASE }
  }
  if (typeof input !== 'object') {
    return { type: 'linear', axes: null, base: DEFAULT_LOG_BASE }
  }

  const axes = parseAxes(input.axes)
  const type: ScaleType =
    input.type === 'log' || input.type === 'linear'
      ? input.type
      : axes !== null && axes.length > 0
        ? 'log'
        : 'linear'

  return {
    type,
    axes,
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

/** Linear / Logarithmic tab value. Object scale is display-only; the toggle still writes a string. */
export function scaleTabValue(scale: ScaleInput | undefined | null): ScaleType {
  return parseScale(scale).type
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
