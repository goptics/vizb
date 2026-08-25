import { describe, it, expect } from 'vitest'
import {
  DEFAULT_LOG_AXES,
  DEFAULT_LOG_BASE,
  axisIsLog,
  axisLogBase,
  numericLogXValues,
  parseScale,
  scaleTabValue,
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

  it('omitted type with axes implies log', () => {
    const parsed = parseScale({ axes: ['x'] })
    expect(parsed.type).toBe('log')
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
    expect(axisIsLog(parsed, 'y', DEFAULT_LOG_AXES.grouped2d)).toBe(true)
    expect(axisIsLog(parsed, 'x', DEFAULT_LOG_AXES.grouped2d)).toBe(false)
    expect(axisIsLog(parsed, 'x', DEFAULT_LOG_AXES.horizontalBar)).toBe(true)
    expect(axisIsLog(parsed, 'z', DEFAULT_LOG_AXES.grouped3d)).toBe(true)
    expect(axisIsLog(parsed, 'x', DEFAULT_LOG_AXES.continuous3d)).toBe(true)
    expect(axisIsLog(parsed, 'y', DEFAULT_LOG_AXES.continuous3d)).toBe(true)
    expect(axisIsLog(parsed, 'z', DEFAULT_LOG_AXES.continuous3d)).toBe(true)
  })

  it('explicit axes override the default', () => {
    const parsed = parseScale({ type: 'log', axes: ['x'] })
    expect(axisIsLog(parsed, 'x', DEFAULT_LOG_AXES.grouped2d)).toBe(true)
    expect(axisIsLog(parsed, 'y', DEFAULT_LOG_AXES.grouped2d)).toBe(false)
  })

  it('explicit empty axes logs nothing', () => {
    const parsed = parseScale({ type: 'log', axes: [] })
    expect(axisIsLog(parsed, 'y', DEFAULT_LOG_AXES.grouped2d)).toBe(false)
  })
})

describe('scaleTabValue / validLogBase / numericLogXValues', () => {
  it('maps object scale to the Linear / Logarithmic tab', () => {
    expect(scaleTabValue(undefined)).toBe('linear')
    expect(scaleTabValue('log')).toBe('log')
    expect(scaleTabValue({ type: 'log', axes: ['x'] })).toBe('log')
    expect(scaleTabValue({ axes: ['x'] })).toBe('log')
    expect(scaleTabValue({ type: 'linear' })).toBe('linear')
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
})
