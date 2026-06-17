import { describe, it, expect } from 'vitest'
import { boxSizeForAxisCount, create3DGridConfig } from './3d'

const styling = {
  textColor: '#111',
  axisColor: '#ccc',
  opacity: 0.5,
  backgroundColor: undefined,
}

describe('boxSizeForAxisCount', () => {
  it.each([
    [0, 80],
    [1, 80],
    [4, 80],
    [5, 100],
    [9, 100],
    [10, 100],
    [14, 100],
    [15, 200],
    [20, 200],
    [100, 200],
  ])('len %i → %i', (len, expected) => {
    expect(boxSizeForAxisCount(len)).toBe(expected)
  })
})

describe('create3DGridConfig', () => {
  it('boxWidth follows xCount and boxDepth follows yCount independently', () => {
    const grid = create3DGridConfig({
      styling,
      autoRotate: false,
      xCount: 3,
      yCount: 12,
    })
    expect(grid.boxWidth).toBe(80)
    expect(grid.boxDepth).toBe(100)
  })

  it('viewControl distance is sum of boxWidth and boxDepth', () => {
    const grid = create3DGridConfig({
      styling,
      autoRotate: false,
      xCount: 15,
      yCount: 5,
    })
    expect(grid.viewControl.distance).toBe(300)
  })
})
