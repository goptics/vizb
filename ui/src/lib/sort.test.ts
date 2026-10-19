import { describe, it, expect } from 'vitest'
import { resolveSort } from './sort'

describe('resolveSort', () => {
  const order = (o: 'asc' | 'desc' | 'best', better?: 'high' | 'low') =>
    resolveSort({ enabled: true, order: o }, better).order

  it('keeps asc and desc regardless of direction', () => {
    expect(order('asc', 'high')).toBe('asc')
    expect(order('desc', 'low')).toBe('desc')
  })

  it('resolves best from the chart direction, desc when unknown', () => {
    expect(order('best', 'low')).toBe('asc')
    expect(order('best', 'high')).toBe('desc')
    expect(order('best')).toBe('desc')
  })

  it('keeps enabled', () => {
    expect(resolveSort({ enabled: false, order: 'desc' })).toEqual({
      enabled: false,
      order: 'desc',
    })
  })
})
