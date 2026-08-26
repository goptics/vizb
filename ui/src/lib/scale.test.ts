import { describe, it, expect } from 'vitest'
import {
  DEFAULT_LOG_BASE,
  applyScaleType,
  asLogXPairs,
  axisIsLog,
  axisLogBase,
  numericLogXValues,
  parseScale,
  scaleLogInfoText,
  validLogBase,
} from './scale'

describe('parseScale', () => {
  it('treats omitted, null, and linear as linear with default base', () => {
    expect(parseScale(undefined)).toEqual({ type: 'linear', axes: null, base: DEFAULT_LOG_BASE })
    expect(parseScale(null)).toEqual({ type: 'linear', axes: null, base: DEFAULT_LOG_BASE })
    expect(parseScale('linear')).toEqual({ type: 'linear', axes: null, base: DEFAULT_LOG_BASE })
  })

  it('parses the log string as log with omitted axes', () => {
    expect(parseScale('log')).toEqual({ type: 'log', axes: null, base: DEFAULT_LOG_BASE })
  })

  it('ignores non-object non-string input', () => {
    expect(parseScale(1 as never)).toEqual({ type: 'linear', axes: null, base: DEFAULT_LOG_BASE })
  })

  it('object type log without axes uses chart defaults later', () => {
    expect(parseScale({ type: 'log' })).toMatchObject({ type: 'log', axes: null, base: 10 })
  })

  it('object type linear never logs even with axes', () => {
    const parsed = parseScale({ type: 'linear', axes: ['x'] })
    expect(parsed.type).toBe('linear')
    expect(axisIsLog(parsed, 'x', ['x'])).toBe(false)
  })

  it('omitted type stays linear even with axes', () => {
    const parsed = parseScale({ axes: ['x'] })
    expect(parsed.type).toBe('linear')
    expect(parsed.axes).toEqual(['x'])
  })

  it('unknown type without axes is linear', () => {
    expect(parseScale({ type: 'foo' as never })).toMatchObject({ type: 'linear' })
  })

  it('filters invalid axes and keeps explicit empty', () => {
    expect(parseScale({ type: 'log', axes: ['q' as never, 'x', 1 as never] }).axes).toEqual(['x'])
    expect(parseScale({ type: 'log', axes: ['q' as never] }).axes).toEqual([])
    expect(parseScale({ type: 'log', axes: 'x' as never }).axes).toBeNull()
  })

  it('defaults invalid base to 10 and keeps per-axis overrides', () => {
    expect(parseScale({ type: 'log', base: 0 }).base).toBe(10)
    expect(parseScale({ type: 'log', base: 1 }).base).toBe(10)
    expect(parseScale({ type: 'log', base: -2 }).base).toBe(10)
    expect(parseScale({ type: 'log', base: Number.NaN }).base).toBe(10)
    expect(parseScale({ type: 'log', base: 2 }).base).toBe(2)
    const parsed = parseScale({ type: 'log', base: 2, baseX: 4, baseY: 1, baseZ: 8 })
    expect(axisLogBase(parsed, 'x')).toBe(4)
    expect(axisLogBase(parsed, 'y')).toBe(2)
    expect(axisLogBase(parsed, 'z')).toBe(8)
  })
})

describe('axisIsLog', () => {
  it('string log uses the chart default value axis only', () => {
    const parsed = parseScale('log')
    expect(axisIsLog(parsed, 'y', ['y'])).toBe(true)
    expect(axisIsLog(parsed, 'x', ['y'])).toBe(false)
    expect(axisIsLog(parsed, 'x', ['x'])).toBe(true)
    expect(axisIsLog(parsed, 'z', ['z'])).toBe(true)
    expect(axisIsLog(parsed, 'x', ['x', 'y', 'z'])).toBe(true)
    expect(axisIsLog(parsed, 'y', ['x', 'y', 'z'])).toBe(true)
    expect(axisIsLog(parsed, 'z', ['x', 'y', 'z'])).toBe(true)
  })

  it('explicit axes override the default', () => {
    const parsed = parseScale({ type: 'log', axes: ['x'] })
    expect(axisIsLog(parsed, 'x', ['y'])).toBe(true)
    expect(axisIsLog(parsed, 'y', ['y'])).toBe(false)
  })

  it('explicit empty axes logs nothing', () => {
    const parsed = parseScale({ type: 'log', axes: [] })
    expect(axisIsLog(parsed, 'y', ['y'])).toBe(false)
  })
})

describe('validLogBase / numericLogXValues / asLogXPairs', () => {
  it('maps object scale type for the Linear / Logarithmic tab', () => {
    expect(parseScale(undefined).type).toBe('linear')
    expect(parseScale('log').type).toBe('log')
    expect(parseScale({ type: 'log', axes: ['x'] }).type).toBe('log')
    expect(parseScale({ axes: ['x'] }).type).toBe('linear')
    expect(parseScale({ type: 'linear' }).type).toBe('linear')
  })

  it('rejects invalid log bases', () => {
    expect(validLogBase(10)).toBe(10)
    expect(validLogBase(2)).toBe(2)
    expect(validLogBase(0)).toBeUndefined()
    expect(validLogBase(1)).toBeUndefined()
    expect(validLogBase('10')).toBeUndefined()
  })

  it('coerces numeric-step labels only when x is log and every label is > 0', () => {
    expect(numericLogXValues(['1', '2', '10'], true)).toEqual([1, 2, 10])
    expect(numericLogXValues(['1', 'Jan'], true)).toBeNull()
    expect(numericLogXValues(['0', '1'], true)).toBeNull()
    expect(numericLogXValues(['-1', '2'], true)).toBeNull()
    expect(numericLogXValues(['1', '2'], false)).toBeNull()
    expect(numericLogXValues([], true)).toBeNull()
  })

  it('pairs log-X numbers with Y and nulls non-positive Y when Y is log', () => {
    expect(asLogXPairs([1, 10], [2, 4], false)).toEqual([
      [1, 2],
      [10, 4],
    ])
    expect(asLogXPairs([1, 10], [0, 4], true)).toEqual([
      [1, null],
      [10, 4],
    ])
    expect(asLogXPairs([1], [null], true)).toEqual([[1, null]])
  })
})

describe('scaleLogInfoText', () => {
  it('is empty for linear', () => {
    expect(scaleLogInfoText(undefined)).toBeUndefined()
    expect(scaleLogInfoText('linear')).toBeUndefined()
    expect(scaleLogInfoText({ type: 'linear', axes: ['x'] })).toBeUndefined()
  })

  it('names only the default value axis for string log', () => {
    expect(scaleLogInfoText('log')).toBe('Y log (base 10)')
    expect(scaleLogInfoText('log', ['x'])).toBe('X log (base 10)')
    expect(scaleLogInfoText('log', ['z'])).toBe('Z log (base 10)')
    expect(scaleLogInfoText({ type: 'log' })).toBe('Y log (base 10)')
  })

  it('names linear axes when the object lists some axes but not others', () => {
    expect(scaleLogInfoText({ type: 'log', axes: ['x'], base: 10 })).toBe(
      'X log (base 10) · Y linear'
    )
    expect(scaleLogInfoText({ type: 'log', axes: ['x'], base: 10 }, ['z'])).toBe(
      'X log (base 10) · Y linear · Z linear'
    )
  })

  it('uses per-axis bases when both axes are log', () => {
    expect(scaleLogInfoText({ type: 'log', axes: ['x', 'y'], baseX: 5, baseY: 10 })).toBe(
      'X log (base 5) · Y log (base 10)'
    )
  })

  it('treats an explicit empty axes list as all-linear chart axes', () => {
    expect(scaleLogInfoText({ type: 'log', axes: [] })).toBe('X linear · Y linear')
  })
})

describe('applyScaleType', () => {
  it('applyScaleType merges type into an object spec and returns strings as-is', () => {
    const spec = { type: 'log' as const, axes: ['x'] as ['x'], base: 10, baseX: 5 }
    expect(applyScaleType(spec, 'log')).toEqual(spec)
    expect(applyScaleType(spec, 'linear')).toEqual({ ...spec, type: 'linear' })
    expect(applyScaleType({ type: 'linear', axes: ['x'] }, 'log')).toEqual({
      type: 'log',
      axes: ['x'],
    })
    expect(applyScaleType('linear', 'log')).toBe('log')
    expect(applyScaleType('log', 'linear')).toBe('linear')
    expect(applyScaleType(undefined, 'log')).toBe('log')
  })
})
