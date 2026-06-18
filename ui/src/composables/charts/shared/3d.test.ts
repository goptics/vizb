import { describe, it, expect } from 'vitest'
import {
  boxSizeForAxisCount,
  create3DGridConfig,
  create3DVisualMap,
  createZLegendConfig,
  resolve3DVisualMap,
} from './3d'
import { resetColor } from '@/lib/utils'

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

describe('create3DVisualMap', () => {
  it('positions visualMap vertically at the right-center with metric dimension', () => {
    const visualMap = create3DVisualMap(42, styling)
    expect(visualMap).toMatchObject({
      show: true,
      min: 0,
      max: 42,
      dimension: 2,
      calculable: true,
      orient: 'vertical',
      right: '0%',
      top: 'center',
    })
    expect(visualMap.inRange.color.length).toBeGreaterThan(1)
  })
})

describe('createZLegendConfig', () => {
  it('pins palette colors on legend data so visualMap does not change swatches', () => {
    resetColor()
    const legend = createZLegendConfig(['alpha', 'beta'], styling, { alpha: true, beta: false })
    expect(legend.data).toEqual([
      { name: 'alpha', itemStyle: { color: '#5470C6' } },
      { name: 'beta', itemStyle: { color: '#3BA272' } },
    ])
    expect(legend.selected).toEqual({ alpha: true, beta: false })
  })
})

describe('resolve3DVisualMap', () => {
  const series = [{ data: [{ value: [0, 0, 42] }] }]

  it('returns config when enabled', () => {
    expect(resolve3DVisualMap(true, series, styling)).toMatchObject({ show: true, max: 42 })
  })

  it('returns empty array when disabled so ECharts can replace-merge it away', () => {
    expect(resolve3DVisualMap(false, series, styling)).toEqual([])
  })
})
