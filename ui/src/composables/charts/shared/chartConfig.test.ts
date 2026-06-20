import { describe, it, expect } from 'vitest'
import { createGridConfig, formatRadarItemTooltip, hasRotatedXLabels } from './chartConfig'

const indicators = ['A', 'B', 'C']

describe('createGridConfig', () => {
  it('reserves fixed px bottom only when dataZoom is present', () => {
    expect(createGridConfig(1, true).bottom).toBe(100)
    expect(createGridConfig(1, true).containLabel).toBe(false)
  })

  it('uses containLabel for no-dataZoom layout (axis title space is automatic)', () => {
    expect(createGridConfig(1, false).bottom).toBe(28)
    expect(createGridConfig(1, false).containLabel).toBe(true)
  })

  it('keeps dataZoom bottom larger than the no-zoom tier', () => {
    expect(createGridConfig(1, true).bottom).toBeGreaterThan(createGridConfig(1, false).bottom)
  })
})

describe('hasRotatedXLabels', () => {
  it('returns false for large axes (dataZoom handles navigation)', () => {
    expect(hasRotatedXLabels(['a'.repeat(60)], true)).toBe(false)
  })

  it('returns true when total label length exceeds threshold on small axes', () => {
    expect(
      hasRotatedXLabels([Array(60).fill('x').join(''), Array(50).fill('y').join('')], false)
    ).toBe(true)
  })
})

describe('formatRadarItemTooltip', () => {
  it('returns empty string when params.data is missing', () => {
    expect(formatRadarItemTooltip({}, indicators, false)).toBe('')
  })

  it('single spoke: rows only, no Σ / spread / donut', () => {
    const html = formatRadarItemTooltip({ data: { name: 'Series', value: [10] } }, ['A'], false)
    expect(html).toContain('<b>Series</b>')
    expect(html).toContain('A: <b>10</b>')
    expect(html).not.toContain('Σ')
    expect(html).not.toContain('Median')
    expect(html).not.toContain('<svg')
  })

  it('multi-spoke: includes Σ, spread stats, and donut', () => {
    const html = formatRadarItemTooltip(
      { data: { name: 'Series', value: [10, 20, 30] } },
      indicators,
      false
    )
    expect(html).toContain('Σ Series: <b>60</b>')
    expect(html).toContain('Median:')
    expect(html).toContain('IQR:')
    expect(html).toContain('CV:')
    expect(html).toContain('<svg')
  })

  it('uses seriesName / data.name header when both differ (X+Y+Z)', () => {
    const html = formatRadarItemTooltip(
      {
        seriesName: 'Pool1',
        data: { name: 'alloc', value: [1, 2] },
      },
      ['Y1', 'Y2'],
      false
    )
    expect(html).toContain('<b>Pool1 / alloc</b>')
  })
})
