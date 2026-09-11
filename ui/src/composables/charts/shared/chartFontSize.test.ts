import { describe, expect, it } from 'vitest'
import { appearanceFontSize, chartFontSize } from './chartFontSize'
import { DEFAULT_FONT_SIZE } from '@/lib/fontSize'

describe('chartFontSize', () => {
  it('defaults to 12 when unset', () => {
    appearanceFontSize.value = undefined
    expect(chartFontSize()).toEqual({
      series: DEFAULT_FONT_SIZE,
      legend: DEFAULT_FONT_SIZE,
      label: DEFAULT_FONT_SIZE,
    })
  })

  it('resolves a baked number and a partial object', () => {
    appearanceFontSize.value = 14
    expect(chartFontSize().legend).toBe(14)
    appearanceFontSize.value = { series: 16 }
    expect(chartFontSize()).toEqual({
      series: 16,
      legend: DEFAULT_FONT_SIZE,
      label: DEFAULT_FONT_SIZE,
    })
  })
})
