import { describe, it, expect } from 'vitest'
import { shouldUseTabPicker, CHART_PICKER_TAB_THRESHOLD } from './pickerRule'

describe('shouldUseTabPicker', () => {
  it('threshold constant is 3', () => {
    expect(CHART_PICKER_TAB_THRESHOLD).toBe(3)
  })

  it('returns true for 1, 2, and 3 chart types (use tabs)', () => {
    expect(shouldUseTabPicker(1)).toBe(true)
    expect(shouldUseTabPicker(2)).toBe(true)
    expect(shouldUseTabPicker(3)).toBe(true)
  })

  it('returns false for 4, 5, and 6 chart types (use combobox)', () => {
    expect(shouldUseTabPicker(4)).toBe(false)
    expect(shouldUseTabPicker(5)).toBe(false)
    expect(shouldUseTabPicker(6)).toBe(false)
  })
})
