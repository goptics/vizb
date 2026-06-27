import { describe, it, expect } from 'vitest'
import {
  bandFillRatioForCount,
  barSizeFor3DGrid,
  barSizeForContinuous3D,
  boxSizeForAxisCount,
  orthographicSizeFor3DBox,
  MAX_3D_VIEW_DISTANCE,
  groupedViewDistanceFor3DBox,
  viewDistanceFor3DBox,
  create3DGridConfig,
  create3DTooltipFormatter,
  createContinuous3DTooltipFormatter,
  create3DVisualMap,
  createValue3DTooltipFormatter,
  createZLegendConfig,
  resolve3DVisualMap,
  symbolSizeFor3DGrid,
  symbolSizeForContinuous3D,
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

describe('bandFillRatioForCount', () => {
  it('clamps sparse grids to low fill and dense grids to high fill', () => {
    expect(bandFillRatioForCount(1)).toBe(0.52)
    expect(bandFillRatioForCount(2)).toBe(0.52)
    expect(bandFillRatioForCount(40)).toBe(0.96)
    expect(bandFillRatioForCount(100)).toBe(0.96)
  })

  it('ramps monotonically between sparse and dense', () => {
    const mid = bandFillRatioForCount(20)
    expect(mid).toBeGreaterThan(0.52)
    expect(mid).toBeLessThan(0.96)
    expect(bandFillRatioForCount(10)).toBeLessThan(bandFillRatioForCount(30))
  })
})

describe('barSizeFor3DGrid', () => {
  it('uses a larger fill fraction on dense grids than sparse grids', () => {
    const sparse = barSizeFor3DGrid(3, 3, 80, 80)
    const dense = barSizeFor3DGrid(30, 30, 200, 200)
    const sparseFillX = sparse[0]! / (80 / 3)
    const denseFillX = dense[0]! / (200 / 30)
    expect(denseFillX).toBeGreaterThan(sparseFillX)
  })
})

describe('symbolSizeFor3DGrid', () => {
  it('fills a smaller fraction of each cell on sparse grids than dense grids', () => {
    const sparse = symbolSizeFor3DGrid(3, 3, 80, 80)
    const dense = symbolSizeFor3DGrid(30, 30, 80, 80)
    const sparseBand = 80 / 3
    const denseBand = 80 / 30
    expect(sparse / sparseBand).toBeCloseTo(bandFillRatioForCount(3), 5)
    expect(dense / denseBand).toBeCloseTo(bandFillRatioForCount(30), 5)
  })
})

describe('continuous 3D spacing', () => {
  it('shrinks bar footprint as point count grows', () => {
    const few = barSizeForContinuous3D(10, 80, 80)
    const many = barSizeForContinuous3D(1000, 200, 200)
    expect(many[0]!).toBeLessThan(few[0]!)
  })

  it('fills more of each synthetic grid cell as point count grows', () => {
    const few = symbolSizeForContinuous3D(10, 80, 80)
    const many = symbolSizeForContinuous3D(1000, 200, 200)
    const fewBand = 80 / 10
    const manyBand = 200 / 100
    expect(many / manyBand).toBeGreaterThan(few / fewBand)
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
    expect('boxHeight' in grid).toBe(false)
    expect(grid.viewControl.distance).toBe(180)
  })

  it('continuous mode uses a cubic box and scaled orthographic framing', () => {
    const grid = create3DGridConfig({
      styling,
      autoRotate: false,
      orthographic: true,
      xCount: 100,
      yCount: 100,
      mode: 'continuous',
    })
    expect(grid.boxWidth).toBe(200)
    expect(grid.boxDepth).toBe(200)
    expect('boxHeight' in grid && grid.boxHeight).toBe(200)
    if ('orthographicSize' in grid.viewControl) {
      expect(grid.viewControl.projection).toBe('orthographic')
      expect(grid.viewControl.orthographicSize).toBe(orthographicSizeFor3DBox(200, 200, 200))
      expect(grid.viewControl.maxOrthographicSize).toBeGreaterThanOrEqual(400)
    }
  })

  it('continuous perspective viewControl distance is 2× the largest box edge', () => {
    const grid = create3DGridConfig({
      styling,
      autoRotate: false,
      xCount: 15,
      yCount: 5,
      mode: 'continuous',
    })
    expect(grid.viewControl.distance).toBe(viewDistanceFor3DBox(200, 100, 200))
    if ('maxDistance' in grid.viewControl) {
      expect(grid.viewControl.maxDistance).toBe(MAX_3D_VIEW_DISTANCE)
    }
  })

  it('caps viewControl distance at 300 for large grouped grids', () => {
    const grid = create3DGridConfig({
      styling,
      autoRotate: false,
      xCount: 100,
      yCount: 100,
    })
    expect(groupedViewDistanceFor3DBox(200, 200)).toBe(MAX_3D_VIEW_DISTANCE)
    expect(grid.viewControl.distance).toBe(MAX_3D_VIEW_DISTANCE)
  })

  it('value mode sizes footprint from category counts for rectangular data', () => {
    const grid = create3DGridConfig({
      styling,
      autoRotate: false,
      xCount: 3,
      yCount: 30,
      mode: 'value',
    })
    expect(grid.boxWidth).toBe(80)
    expect(grid.boxDepth).toBe(200)
    expect('boxHeight' in grid && grid.boxHeight).toBe(80)
    expect(grid.viewControl.distance).toBe(viewDistanceFor3DBox(80, 200, 80))
  })

  it('value mode uses a cube only when x and y category counts match', () => {
    const grid = create3DGridConfig({
      styling,
      autoRotate: false,
      orthographic: true,
      xCount: 10,
      yCount: 10,
      mode: 'value',
    })
    expect(grid.boxWidth).toBe(100)
    expect(grid.boxDepth).toBe(100)
    expect('boxHeight' in grid && grid.boxHeight).toBe(100)
    expect(grid.viewControl.distance).toBe(viewDistanceFor3DBox(100, 100, 100))
    if ('orthographicSize' in grid.viewControl) {
      expect(grid.viewControl.orthographicSize).toBe(orthographicSizeFor3DBox(100, 100, 100))
    }
  })

  it('fades grid split lines and keeps axisPointer on axisColor', () => {
    const darkGrid = create3DGridConfig({
      styling: {
        textColor: '#e5e7eb',
        axisColor: '#4b5563',
        opacity: 0.15,
        backgroundColor: 'transparent',
      },
      autoRotate: false,
      xCount: 3,
      yCount: 3,
    })
    const lightGrid = create3DGridConfig({
      styling: {
        ...styling,
        axisColor: '#d1d5db',
        opacity: 0.4,
      },
      autoRotate: false,
      xCount: 3,
      yCount: 3,
    })
    expect(darkGrid.splitLine.lineStyle.opacity).toBeCloseTo(0.1125)
    expect(lightGrid.splitLine.lineStyle.opacity).toBeCloseTo(0.3)
    expect(darkGrid.axisPointer.lineStyle).toEqual({
      color: '#4b5563',
      width: 2,
      opacity: 1,
    })
    expect(darkGrid.axisPointer.label).toEqual({
      color: '#e5e7eb',
      textStyle: { color: '#e5e7eb' },
    })
    expect(lightGrid.axisPointer.lineStyle.color).toBe('#d1d5db')
    expect(lightGrid.axisPointer.label.textStyle.color).toBe('#111')
  })
})

describe('createContinuous3DTooltipFormatter', () => {
  it('shows x, y, z coordinates without metric row for 3-tuples', () => {
    const formatter = createContinuous3DTooltipFormatter(false, {
      x: 'i',
      y: 'j',
      z: 'k',
    })
    const html = formatter({ value: [1, 2, 3] })
    expect(html).toContain('<b>i: 1</b>')
    expect(html).toContain('j: 2')
    expect(html).toContain('k: 3')
    expect(html).not.toContain('value:')
  })

  it('appends metric row when a 4th value is present', () => {
    const formatter = createContinuous3DTooltipFormatter(false, {
      x: 'i',
      y: 'j',
      z: 'k',
      metric: 'value',
    })
    const html = formatter({ value: [0, 1, 2, 4.5] })
    expect(html).toContain('k: 2')
    expect(html).toContain('value: <b>4.5</b>')
  })
})

describe('create3DVisualMap', () => {
  it('category mode colors by z height (dimension 2) without symbolSize or colorAlpha', () => {
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
      inRange: { color: expect.any(Array) },
    })
    expect(visualMap.inRange).not.toHaveProperty('symbolSize')
    expect(visualMap.inRange).not.toHaveProperty('colorAlpha')
  })

  it('metric mode uses dimension 3 with symbolSize and colorAlpha', () => {
    const visualMap = create3DVisualMap(6, styling, 3)
    expect(visualMap).toMatchObject({
      dimension: 3,
      min: 0,
      max: 6,
      inRange: {
        symbolSize: [0.5, 25],
        colorAlpha: [0.2, 1],
      },
    })
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

describe('createValue3DTooltipFormatter', () => {
  const xValues = ['East', 'West']
  const yValues = ['A', 'B']
  const seriesData = [
    { value: [0, 0, 10] },
    { value: [1, 0, 30] },
    { value: [0, 1, 20] },
    { value: [1, 1, 40] },
  ]

  const formatter = createValue3DTooltipFormatter({
    xValues,
    yValues,
    seriesData,
    isDark: false,
    xAxisLabel: 'Region',
    yAxisLabel: 'Product',
    valueLabel: 'Revenue',
    seriesColor: '#5470C6',
  })

  it('uses combined x / y header like grouped 3D tooltips', () => {
    const html = formatter({ value: [0, 0, 10] })
    expect(html).toContain('<b>Region: East / Product: A</b>')
    expect(html).not.toMatch(/<b>Region: East<\/b><br\/>Product: A/)
  })

  it('shows metric row with color dot and cell value without grand total', () => {
    const html = formatter({ value: [0, 0, 10] })
    expect(html).toContain('background:#5470C6')
    expect(html).toContain('Revenue: <b>10</b>')
    expect(html).not.toContain('Σ Total')
    expect(html).not.toContain('(Σ100)')
  })

  it('includes row∪column total (no double-count) and per-axis marginal totals', () => {
    const html = formatter({ value: [0, 0, 10] })
    expect(html).toContain('Σ (Region∪Product): <b>60</b>')
    expect(html).toContain('Σ Region(East): <b>30</b>')
    expect(html).toContain('Σ Product(A): <b>40</b>')
    expect(html).not.toContain('Σ (Region,Product,Category)')
  })

  it('lists per-axis marginals before row∪column total', () => {
    const html = formatter({ value: [0, 0, 10] })
    expect(html.indexOf('Σ Region(East)')).toBeLessThan(html.indexOf('Σ (Region∪Product)'))
    expect(html.indexOf('Σ Product(A)')).toBeLessThan(html.indexOf('Σ (Region∪Product)'))
  })

  it('omits spread and donut sections for single-value cells', () => {
    const html = formatter({ value: [1, 1, 40] })
    expect(html).not.toContain('Median:')
    expect(html).not.toContain('<svg')
  })
})

describe('create3DTooltipFormatter', () => {
  const xValues = ['East', 'West']
  const yValues = ['A', 'B']
  const zValues = ['Z1', 'Z2']
  const aggPoints = [
    { xAxis: 'East', yAxis: 'A', zAxis: 'Z1', value: 10 },
    { xAxis: 'East', yAxis: 'A', zAxis: 'Z2', value: 5 },
    { xAxis: 'West', yAxis: 'A', zAxis: 'Z1', value: 20 },
    { xAxis: 'West', yAxis: 'A', zAxis: 'Z2', value: 15 },
    { xAxis: 'East', yAxis: 'B', zAxis: 'Z1', value: 30 },
    { xAxis: 'East', yAxis: 'B', zAxis: 'Z2', value: 10 },
    { xAxis: 'West', yAxis: 'B', zAxis: 'Z1', value: 40 },
    { xAxis: 'West', yAxis: 'B', zAxis: 'Z2', value: 25 },
  ]

  const formatter = create3DTooltipFormatter({
    xValues,
    yValues,
    zValues,
    aggPoints,
    isDark: false,
    xAxisLabel: 'Region',
    yAxisLabel: 'Product',
    zAxisLabel: 'Category',
  })

  it('omits chart-wide grand total (shown on ChartCard badge instead)', () => {
    const html = formatter({ value: [0, 0, 0] })
    expect(html).not.toContain('Σ Total')
    expect(html).not.toContain('Σ (Region,Product,Category): <b>155</b>')
  })

  it('includes row∪column and cell-level x,y,z sum plus per-axis marginals', () => {
    const html = formatter({ value: [0, 0, 0] })
    expect(html).toContain('Σ (Region∪Product): <b>90</b>')
    expect(html).toContain('Σ (Region,Product,Category): <b>15</b>')
    expect(html).toContain('Σ Region(East): <b>55</b>')
    expect(html).toContain('Σ Product(A): <b>50</b>')
  })

  it('uses Σ (x,y,z) for the cell z-sum at the hovered (x,y)', () => {
    const html = formatter({ value: [0, 0, 0] })
    expect(html).toContain('Σ (Region,Product,Category): <b>15</b>')
    expect(html).not.toMatch(/Σ Category: <b>15<\/b>/)
  })

  it('lists per-axis marginals and cell sum before row∪column total', () => {
    const html = formatter({ value: [0, 0, 0] })
    const xIdx = html.indexOf('Σ Region(East)')
    const yIdx = html.indexOf('Σ Product(A)')
    const xyzIdx = html.indexOf('Σ (Region,Product,Category):')
    const xyIdx = html.indexOf('Σ (Region∪Product):')
    expect(xIdx).toBeLessThan(xyIdx)
    expect(yIdx).toBeLessThan(xyIdx)
    expect(xyzIdx).toBeLessThan(xyIdx)
  })

  it('includes row∪column and cell x,y,z sum for single-z cells', () => {
    const singleZFormatter = create3DTooltipFormatter({
      xValues: ['East'],
      yValues: ['A'],
      zValues: ['Z1'],
      aggPoints: [{ xAxis: 'East', yAxis: 'A', zAxis: 'Z1', value: 10 }],
      isDark: false,
      xAxisLabel: 'Region',
      yAxisLabel: 'Product',
      zAxisLabel: 'Category',
    })
    const html = singleZFormatter({ value: [0, 0, 0] })
    expect(html).toContain('Σ (Region∪Product): <b>10</b>')
    expect(html).toContain('Σ (Region,Product,Category): <b>10</b>')
    expect(html).not.toMatch(/Σ Category: <b>10<\/b>/)
  })
})

describe('resolve3DVisualMap', () => {
  const categorySeries = [{ data: [{ value: [0, 0, 42] }] }]
  const metricSeries = [{ data: [{ value: [0, 0, 1, 4.5] }] }]

  it('returns category config on dimension 2 when no metric column', () => {
    expect(resolve3DVisualMap(true, categorySeries, styling)).toMatchObject({
      show: true,
      dimension: 2,
      max: 42,
    })
  })

  it('returns metric config on dimension 3 when a 4th value is present', () => {
    expect(resolve3DVisualMap(true, metricSeries, styling)).toMatchObject({
      show: true,
      dimension: 3,
      max: 4.5,
      inRange: { symbolSize: [0.5, 25], colorAlpha: [0.2, 1] },
    })
  })

  it('returns empty array when disabled so ECharts can replace-merge it away', () => {
    expect(resolve3DVisualMap(false, metricSeries, styling)).toEqual([])
  })
})
