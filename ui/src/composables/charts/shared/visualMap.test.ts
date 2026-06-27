import { describe, it, expect } from 'vitest'
import { getChartStyling } from './chartConfig'
import { resolve2DScatterVisualMap } from './visualMap'

const styling = getChartStyling(false)

describe('resolve2DScatterVisualMap', () => {
  it('matches 3D visualMap layout (vertical, right edge)', () => {
    const visualMap = resolve2DScatterVisualMap(true, [1, 5, 3], styling, 2)
    expect(Array.isArray(visualMap)).toBe(false)
    expect(visualMap).toMatchObject({
      show: true,
      min: 0,
      max: 5,
      dimension: 2,
      orient: 'vertical',
      right: '0%',
      top: 'center',
    })
    if (!Array.isArray(visualMap)) {
      expect(visualMap.inRange).not.toHaveProperty('symbolSize')
    }
  })

  it('returns empty array when disabled', () => {
    expect(resolve2DScatterVisualMap(false, [1, 5], styling)).toEqual([])
  })

  it('returns empty array when enabled but no values', () => {
    expect(resolve2DScatterVisualMap(true, [], styling)).toEqual([])
  })
})
