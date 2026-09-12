import { describe, expect, it } from 'vitest'
import { DEFAULT_FONT_SIZE, resolveFontSize } from './fontSize'

describe('resolveFontSize', () => {
  it('defaults all keys to 14', () => {
    expect(resolveFontSize(undefined)).toEqual({
      series: DEFAULT_FONT_SIZE,
      legend: DEFAULT_FONT_SIZE,
      label: DEFAULT_FONT_SIZE,
    })
  })

  it('applies a number to all three keys', () => {
    expect(resolveFontSize(14)).toEqual({ series: 14, legend: 14, label: 14 })
    expect(resolveFontSize(12.5)).toEqual({ series: 12.5, legend: 12.5, label: 12.5 })
  })

  it('fills missing object keys with 14', () => {
    expect(resolveFontSize({ series: 16, legend: 10 })).toEqual({
      series: 16,
      legend: 10,
      label: DEFAULT_FONT_SIZE,
    })
  })

  it('ignores invalid numbers', () => {
    expect(resolveFontSize(0)).toEqual({
      series: DEFAULT_FONT_SIZE,
      legend: DEFAULT_FONT_SIZE,
      label: DEFAULT_FONT_SIZE,
    })
    expect(resolveFontSize({ series: -1, legend: Number.NaN, label: 14 })).toEqual({
      series: DEFAULT_FONT_SIZE,
      legend: DEFAULT_FONT_SIZE,
      label: 14,
    })
    expect(resolveFontSize(Number.POSITIVE_INFINITY)).toEqual({
      series: DEFAULT_FONT_SIZE,
      legend: DEFAULT_FONT_SIZE,
      label: DEFAULT_FONT_SIZE,
    })
  })
})
