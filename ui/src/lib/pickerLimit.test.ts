import { describe, expect, it } from 'vitest'
import { limitPickerOptions } from './pickerLimit'

const options = Array.from({ length: 1000 }, (_, index) => ({
  value: index.toString(),
  label: `Dataset ${index.toString().padStart(4, '0')}`,
}))

describe('limited dataset picker', () => {
  it('renders at most 100 entries and keeps the active entry visible', () => {
    const visible = limitPickerOptions(options, options[750], 100)
    expect(visible).toHaveLength(100)
    expect(visible).toContain(options[750])
  })

  it('searches the complete catalog before applying the limit', () => {
    const matches = options.filter((option) => option.label.includes('Dataset 0999'))
    const visible = limitPickerOptions(matches, options[0], 100)
    expect(matches).toEqual([options[999]])
    expect(visible).toEqual([options[999], options[0]])
  })

  it('never exceeds the limit when the active entry does not match a full result page', () => {
    const matches = options.filter((option) => option.label.includes('Dataset 0'))
    const visible = limitPickerOptions(matches, options[999], 100)
    expect(visible).toHaveLength(100)
    expect(visible.at(-1)).toBe(options[999])
  })
})
