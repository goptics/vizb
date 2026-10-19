import { describe, it, expect } from 'vitest'
import { effectiveSortOrder, resolveSort } from './sort'

describe('effectiveSortOrder', () => {
  it('keeps asc and desc regardless of direction', () => {
    expect(effectiveSortOrder('asc', 'high')).toBe('asc')
    expect(effectiveSortOrder('desc', 'low')).toBe('desc')
  })

  it('resolves best from the chart direction, desc when unknown', () => {
    expect(effectiveSortOrder('best', 'low')).toBe('asc')
    expect(effectiveSortOrder('best', 'high')).toBe('desc')
    expect(effectiveSortOrder('best', undefined)).toBe('desc')
  })
})

describe('resolveSort', () => {
  it('resolves the order and keeps enabled', () => {
    expect(resolveSort({ enabled: true, order: 'best' }, 'low')).toEqual({
      enabled: true,
      order: 'asc',
    })
    expect(resolveSort({ enabled: false, order: 'desc' })).toEqual({
      enabled: false,
      order: 'desc',
    })
  })
})
