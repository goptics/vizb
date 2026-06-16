import { describe, it, expect } from 'vitest'
import { axisKeyConcat } from './swap'
import type { Axis } from '../types'

describe('axisKeyConcat', () => {
  it('returns an empty string for an empty axis list', () => {
    expect(axisKeyConcat([])).toBe('')
    expect(axisKeyConcat(undefined)).toBe('')
  })

  it('concatenates axis keys in order with name → n', () => {
    const axes: Axis[] = [{ key: 'x' }, { key: 'y' }, { key: 'name' }]
    expect(axisKeyConcat(axes)).toBe('xyn')
  })

  it('uses the first char of single-letter axis keys', () => {
    expect(axisKeyConcat([{ key: 'x' }])).toBe('x')
    expect(axisKeyConcat([{ key: 'y' }])).toBe('y')
    expect(axisKeyConcat([{ key: 'z' }])).toBe('z')
    expect(axisKeyConcat([{ key: 'name' }])).toBe('n')
  })

  it('preserves axis order', () => {
    const a: Axis[] = [{ key: 'name' }, { key: 'x' }]
    const b: Axis[] = [{ key: 'x' }, { key: 'name' }]
    expect(axisKeyConcat(a)).toBe('nx')
    expect(axisKeyConcat(b)).toBe('xn')
  })

  it('handles a z-axis (3D) present in the axis list', () => {
    expect(axisKeyConcat([{ key: 'x' }, { key: 'y' }, { key: 'z' }, { key: 'name' }])).toBe('xyzn')
  })
})
