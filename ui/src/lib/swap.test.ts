import { describe, it, expect } from 'vitest'
import type { DataPoint, Axis } from '../types'
import {
  arrangementHasChartZ,
  swapOptionKeys,
  presentAxisKeys,
  presentAxisString,
  swapAxisLabels,
  sourceFieldForChartAxis,
  translateAxisKey,
} from './swap'

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

describe('presentAxisKeys / presentAxisString', () => {
  it('returns empty for missing data', () => {
    expect(presentAxisKeys(undefined)).toEqual([])
    expect(presentAxisKeys([])).toEqual([])
    expect(presentAxisString(undefined)).toBe('')
  })

  it('lists present axes in nxyz order', () => {
    const data = [dp({ name: 'n', xAxis: '1', yAxis: '2' })]
    expect(presentAxisKeys(data)).toEqual(['n', 'x', 'y'])
    expect(presentAxisString(data)).toBe('nxy')
  })
})

describe('swapOptionKeys empty and axes z override', () => {
  it('returns empty without data', () => {
    expect(swapOptionKeys(undefined)).toEqual([])
    expect(swapOptionKeys([])).toEqual([])
  })

  it('uses axes z flag when present even if data lacks z', () => {
    // grouped: present n/x/y (k=3); axes declare z so pool keeps z without 3D engine
    const data = [dp({ name: 'bench', xAxis: '1', yAxis: '2' })]
    const axes: Axis[] = [{ key: 'z' }]
    const keys = swapOptionKeys(data, false, false, axes)
    expect(keys.some((k) => k.includes('z'))).toBe(true)
    // without axes z and without 3D engine, z is stripped
    expect(swapOptionKeys(data, false, false).every((k) => !k.includes('z'))).toBe(true)
  })
})

describe('swapAxisLabels / sourceFieldForChartAxis', () => {
  it('no-ops without labels or on length mismatch', () => {
    expect(swapAxisLabels('xy', 'yx', undefined)).toBeUndefined()
    expect(swapAxisLabels('xy', 'xyz', { x: 'X', y: 'Y' })).toEqual({ x: 'X', y: 'Y' })
  })

  it('permutes labels with the arrangement', () => {
    expect(swapAxisLabels('xy', 'yx', { x: 'price', y: 'lat' })).toEqual({
      x: 'lat',
      y: 'price',
    })
  })

  it('sourceFieldForChartAxis maps target chart axis to identity field', () => {
    const id = translateAxisKey('xyz')
    const tgt = translateAxisKey('xzy')
    expect(sourceFieldForChartAxis(id, tgt, 'yAxis')).toBe('zAxis')
    expect(sourceFieldForChartAxis(id, tgt, 'xAxis')).toBe('xAxis')
    expect(sourceFieldForChartAxis(id, ['xAxis'], 'zAxis')).toBeUndefined()
  })
})
