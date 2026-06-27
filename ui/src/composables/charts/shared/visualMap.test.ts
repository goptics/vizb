import { describe, it, expect } from 'vitest'
import { getChartStyling } from './chartConfig'
import { createScatterVisualMap, resolve2DScatterVisualMap } from './visualMap'

const styling = getChartStyling(false)

describe('createScatterVisualMap', () => {
  it('returns vertical gradient config on dimension 2', () => {
    const visualMap = createScatterVisualMap(42, styling, 2)
    expect(visualMap).toMatchObject({
      show: true,
      min: 0,
      max: 42,
      dimension: 2,
      orient: 'vertical',
      right: '0%',
    })
    expect(visualMap.inRange).not.toHaveProperty('symbolSize')
  })
})

describe('resolve2DScatterVisualMap', () => {
  it('returns config when enabled with values', () => {
    expect(resolve2DScatterVisualMap(true, [1, 5, 3], styling)).toMatchObject({
      show: true,
      max: 5,
    })
  })

  it('returns empty array when disabled', () => {
    expect(resolve2DScatterVisualMap(false, [1, 5], styling)).toEqual([])
  })

  it('returns empty array when enabled but no values', () => {
    expect(resolve2DScatterVisualMap(true, [], styling)).toEqual([])
  })
})
