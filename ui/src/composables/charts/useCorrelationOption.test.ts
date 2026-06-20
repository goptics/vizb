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
})
