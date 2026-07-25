import { describe, expect, it } from 'vitest'
import { formatPercentageLabel } from './labels'

describe('formatPercentageLabel', () => {
  it.each([
    [1, 3, '33.33%'],
    [1, 8, '12.5%'],
    [-1, 4, '-25%'],
    [0, 4, '0%'],
  ])('formats %s / %s', (value, total, expected) => {
    expect(formatPercentageLabel(value, total)).toBe(expected)
  })

  it.each([
    [1, 0],
    [NaN, 1],
    [Infinity, 1],
    [1, NaN],
    [1, Infinity],
  ])('omits invalid %s / %s', (value, total) => {
    expect(formatPercentageLabel(value, total)).toBe('')
  })
})
