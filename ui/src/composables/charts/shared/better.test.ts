import { describe, it, expect } from 'vitest'
import { applyBestMarker, bestMask, betterHint, BEST_MARK } from './better'

describe('betterHint', () => {
  it('describes each direction and is empty when unknown', () => {
    expect(betterHint('low')).toBe('Lower is better')
    expect(betterHint('high')).toBe('Higher is better')
    expect(betterHint(undefined)).toBe('')
  })
})

describe('bestMask', () => {
  it('marks the lowest per category across rows', () => {
    expect(
      bestMask(
        [
          [10, 1],
          [7, 2],
          [12, 3],
        ],
        'low',
        [true, true, true]
      )
    ).toEqual([
      [false, true],
      [true, false],
      [false, false],
    ])
  })

  it('marks the highest and keeps ties', () => {
    expect(bestMask([[5], [5], [1]], 'high', [true, true, true])).toEqual([[true], [true], [false]])
  })

  it('ignores hidden rows and nulls', () => {
    expect(bestMask([[1], [3], [2]], 'low', [false, true, true])).toEqual([
      [false],
      [false],
      [true],
    ])
    expect(bestMask([[null], [3]], 'low', [true, true])).toEqual([[false], [false]])
  })

  it('compares categories when only one row is visible', () => {
    expect(bestMask([[4, 2, 9]], 'low', [true])).toEqual([[false, true, false]])
    expect(
      bestMask(
        [
          [4, 2],
          [1, 1],
        ],
        'high',
        [true, false]
      )
    ).toEqual([
      [true, false],
      [false, false],
    ])
  })

  it('marks nothing without competition', () => {
    expect(bestMask([[4]], 'low', [true])).toEqual([[false]])
    expect(bestMask([], 'low', [])).toEqual([])
  })
})

describe('applyBestMarker', () => {
  const series = () => [
    {
      name: 'a',
      data: [10, 1] as unknown[],
      label: {
        formatter: (p: { value?: number | (number | null)[] | null }) => `v${p.value}`,
      },
    },
    { name: 'b', data: [7, 2] as unknown[] },
  ]
  const opts = { isDark: false, showLabels: false }

  it('styles best points with a star label', () => {
    const s = series()
    applyBestMarker(s, 'low', {}, opts)
    expect(s[0]!.data[0]).toBe(10)
    const best = s[1]!.data[0] as {
      value: number
      label: { formatter: (p: object) => string }
    }
    expect(best.value).toBe(7)
    expect(best.label.formatter({ value: 7 })).toBe(BEST_MARK)
    expect((best as unknown as { itemStyle: { borderColor: string } }).itemStyle.borderColor).toBe(
      '#d97706'
    )
    // series b has no value formatter: hover label falls back to the raw value
    const hoverB = best as unknown as { emphasis: { label: { formatter: (p: object) => string } } }
    expect(hoverB.emphasis.label.formatter({ value: 7 })).toBe(`${BEST_MARK} 7`)
    const withBase = s[0]!.data[1] as { emphasis: { label: { formatter: (p: object) => string } } }
    expect(withBase.emphasis.label.formatter({ value: 1 })).toBe(`${BEST_MARK} v1`)
  })

  it('prefixes the value label when labels are on', () => {
    const s = series()
    applyBestMarker(s, 'low', {}, { ...opts, showLabels: true })
    const best = s[0]!.data[1] as {
      label: { formatter: (p: object) => string }
    }
    expect(best.label.formatter({ value: 1 })).toBe(`${BEST_MARK} v1`)
  })

  it('recomputes from legend selection', () => {
    const s = series()
    // b hidden: a is the only visible series, so its best category wins.
    applyBestMarker(s, 'low', { b: false }, opts)
    expect(typeof s[0]!.data[0]).toBe('number')
    expect(typeof s[0]!.data[1]).toBe('object')
    expect(s[1]!.data).toEqual([7, 2])
  })

  it('uses a light border in dark mode and treats unnamed series as visible', () => {
    const s = [{ data: [3, 1] as unknown[] }]
    applyBestMarker(s, 'low', {}, { isDark: true, showLabels: true })
    const best = s[0]!.data[1] as {
      itemStyle: { borderColor: string }
      label: { formatter: (p: object) => string }
    }
    expect(best.itemStyle.borderColor).toBe('#fbbf24')
    expect(best.label.formatter({ value: 1 })).toBe(BEST_MARK)
  })

  it('skips large series', () => {
    const s = [{ data: Array.from({ length: 2001 }, (_, i) => i) as unknown[] }]
    applyBestMarker(s, 'low', {}, opts)
    expect(s[0]!.data[0]).toBe(0)
  })

  it('is a no-op without direction or with tuple data', () => {
    const s = series()
    applyBestMarker(s, undefined, {}, opts)
    expect(s[1]!.data[0]).toBe(7)
    const tuples = [{ data: [[1, 2]] as unknown[] }, { data: [[1, 3]] as unknown[] }]
    applyBestMarker(tuples, 'low', {}, opts)
    expect(tuples[0]!.data[0]).toEqual([1, 2])
  })
})
