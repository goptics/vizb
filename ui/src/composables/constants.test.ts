import { describe, it, expect } from 'vitest'
import { ALL_CHART_TYPES } from './constants'

describe('ALL_CHART_TYPES', () => {
  it('lists the closed set of 2D chart types', () => {
    expect(ALL_CHART_TYPES).toEqual(['bar', 'line', 'scatter', 'pie', 'heatmap', 'radar', 'sankey'])
  })
})
