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

describe('maxFromScatterValues', () => {
  it('returns 1 when all values are <= 0', async () => {
    const { maxFromScatterValues, resolve2DScatterVisualMap } = await import('./visualMap')
    expect(maxFromScatterValues([])).toBe(1)
    expect(maxFromScatterValues([-1, 0])).toBe(1)
    expect(maxFromScatterValues([3, 7, 2])).toBe(7)
    // dimension default branch
    const vm = resolve2DScatterVisualMap(true, [2], styling)
    expect(vm).toMatchObject({ dimension: 2, max: 2 })
  })
})
