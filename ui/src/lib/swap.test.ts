import { describe, it, expect } from 'vitest'
import { arrangementHasChartZ } from './swap'

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
