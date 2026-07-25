import { describe, it, expect } from 'vitest'
import { buildCorrelationOption } from './useCorrelationOption'
import { LARGE_X_THRESHOLD } from './shared'

describe('buildCorrelationOption', () => {
  const labels = ['A', 'B', 'C']
  const matrix = [
    [1, 0.5, 0.2],
    [0.5, 1, 0.8],
    [0.2, 0.8, 1],
  ]

  it('shows x-axis series labels', () => {
    const option = buildCorrelationOption(labels, matrix, false)
    const xAxis = option.xAxis as { axisLabel?: { show?: boolean } }
    expect(xAxis.axisLabel?.show).not.toBe(false)
  })

  it('attaches dataZoom when series count exceeds threshold', () => {
    const many = Array.from({ length: LARGE_X_THRESHOLD + 1 }, (_, i) => `s${i}`)
    const bigMatrix = many.map(() => many.map(() => 0.5))
    const option = buildCorrelationOption(many, bigMatrix, false)
    expect(option.dataZoom).toBeDefined()
    expect((option.dataZoom as unknown[]).length).toBeGreaterThan(0)
  })

  it('uses larger grid bottom when dataZoom is present', () => {
    const many = Array.from({ length: LARGE_X_THRESHOLD + 1 }, (_, i) => `s${i}`)
    const bigMatrix = many.map(() => many.map(() => 0.5))
    const large = buildCorrelationOption(many, bigMatrix, false)
    const small = buildCorrelationOption(labels, matrix, false)
    expect((large.grid as { bottom: number }).bottom).toBeGreaterThan(
      (small.grid as { bottom: number }).bottom
    )
  })

  it('formats tooltip and cell labels', () => {
    const option = buildCorrelationOption(labels, matrix, false, 'corr', 'spearman')
    const tooltip = option.tooltip as { formatter: (p: unknown) => string }
    expect(tooltip.formatter({ data: [1, 0, 0.5] })).toContain('ρ = 0.500')
    expect(tooltip.formatter({ data: [99, 99, 0.1] })).toContain('—')
    const series = option.series as { label: { formatter: (p: unknown) => string } }[]
    expect(series[0]!.label.formatter({ data: [0, 0, 0.555] })).toBe('0.56')
  })

  it('uses dcor sequential palette and dark neutrals', () => {
    const dark = buildCorrelationOption(labels, matrix, true, 'c', 'dcor')
    const light = buildCorrelationOption(labels, matrix, false, 'c', 'dcor')
    const darkVm = dark.visualMap as { min: number; inRange: { color: string[] } }
    const lightVm = light.visualMap as { min: number; inRange: { color: string[] } }
    expect(darkVm.min).toBe(0)
    expect(lightVm.min).toBe(0)
    expect(darkVm.inRange.color[0]).toBe('#1e293b')
    expect(lightVm.inRange.color[0]).toBe('#f9fafb')
  })

  it('uses kendall/pearson prefixes and skips non-finite cells', () => {
    const m = [
      [1, Number.NaN],
      [Number.POSITIVE_INFINITY, 0.2],
    ]
    const option = buildCorrelationOption(['A', 'B'], m, false, 'c', 'kendall')
    const series = option.series as { data: [number, number, number][] }[]
    expect(series[0]!.data).toEqual([
      [0, 0, 1],
      [1, 1, 0.2],
    ])
    const tooltip = option.tooltip as { formatter: (p: unknown) => string }
    expect(tooltip.formatter({ data: [0, 0, 1] })).toContain('τ = 1.000')
  })

  it('rotates labels when total length exceeds threshold on small matrices', () => {
    const long = [Array(60).fill('x').join(''), Array(50).fill('y').join('')]
    const option = buildCorrelationOption(
      long,
      [
        [1, 0.1],
        [0.1, 1],
      ],
      false
    )
    expect((option.xAxis as { axisLabel?: { rotate?: number } }).axisLabel?.rotate).toBe(30)
  })
})
