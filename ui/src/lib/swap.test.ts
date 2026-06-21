import { describe, it, expect } from 'vitest'
import type { DataPoint } from '../types'
import { arrangementHasChartZ, swapOptionKeys } from './swap'

const dp = (partial: Partial<DataPoint>): DataPoint => ({
  name: '',
  xAxis: '',
  yAxis: '',
  zAxis: '',
  stats: [],
  ...partial,
})

describe('swapOptionKeys', () => {
  it('value mode 2-col offers xy and yx only', () => {
    const data = [dp({ xAxis: '1', yAxis: '2' })]
    expect(swapOptionKeys(data, true)).toEqual(['xy', 'yx'])
  })

  it('value mode 3-col never includes n', () => {
    const data = [dp({ xAxis: '1', yAxis: '2', zAxis: '3' })]
    const keys = swapOptionKeys(data, true)
    expect(keys.every((k) => !k.includes('n'))).toBe(true)
    expect(keys).toContain('xyz')
    expect(keys).toContain('yxz')
  })

  it('grouped mode with name offers n-containing arrangements', () => {
    const data = [dp({ name: 'bench', xAxis: '1', yAxis: '2', zAxis: '3' })]
    const keys = swapOptionKeys(data, false)
    expect(keys.some((k) => k.includes('n'))).toBe(true)
  })

  it('2D grouped data without 3D engine omits z arrangements', () => {
    const data = [dp({ name: 'bench', xAxis: '1', yAxis: '2' })]
    const keys = swapOptionKeys(data, false, false)
    expect(keys.every((k) => !k.includes('z'))).toBe(true)
    expect(keys).toContain('nxy')
    expect(keys).not.toContain('xyz')
  })

  it('2D grouped data with baked --3d offers z arrangements', () => {
    const data = [dp({ name: 'bench', xAxis: '1', yAxis: '2' })]
    const keys = swapOptionKeys(data, false, true)
    expect(keys.some((k) => k.includes('z'))).toBe(true)
    expect(keys).toContain('xyz')
  })
})

describe('arrangementHasChartZ', () => {
  it('is true when z maps to chart zAxis (xyz)', () => {
    expect(arrangementHasChartZ('xyz')).toBe(true)
  })

  it('is false when z is folded to name (xyn, nxy)', () => {
    expect(arrangementHasChartZ('xyn')).toBe(false)
    expect(arrangementHasChartZ('nxy')).toBe(false)
  })

  it('is true for four-axis permutations that keep z on chart axes', () => {
    expect(arrangementHasChartZ('nxyz')).toBe(true)
    expect(arrangementHasChartZ('xyzn')).toBe(true)
  })
})
