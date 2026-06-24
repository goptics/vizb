import { describe, it, expect } from 'vitest'
import {
  bandFillRatioForCount,
  barSizeFor3DGrid,
  barSizeForContinuous3D,
  boxSizeForAxisCount,
  VALUE_MODE_3D_BOX_SIZE,
  VALUE_MODE_3D_VIEW_DISTANCE,
  orthographicSizeFor3DBox,
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
    expect(bandFillRatioForCount(1)).toBe(0.45)
    expect(bandFillRatioForCount(2)).toBe(0.45)
    expect(bandFillRatioForCount(40)).toBe(0.92)
    expect(bandFillRatioForCount(100)).toBe(0.92)
  })

  it('ramps monotonically between sparse and dense', () => {
    const mid = bandFillRatioForCount(20)
    expect(mid).toBeGreaterThan(0.45)
    expect(mid).toBeLessThan(0.92)
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
      expect(grid.viewControl.maxDistance).toBeGreaterThanOrEqual(400)
    }
  })

  it('value mode uses fixed box size and view distance regardless of category count', () => {
    const grid = create3DGridConfig({
      styling,
      autoRotate: false,
      xCount: 3,
      yCount: 30,
      mode: 'value',
    })
    expect(grid.boxWidth).toBe(VALUE_MODE_3D_BOX_SIZE)
    expect(grid.boxDepth).toBe(VALUE_MODE_3D_BOX_SIZE)
    expect('boxHeight' in grid && grid.boxHeight).toBe(VALUE_MODE_3D_BOX_SIZE)
    expect(grid.viewControl.distance).toBe(VALUE_MODE_3D_VIEW_DISTANCE)
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

  it('includes combined x+y total and per-axis marginal totals', () => {
    const html = formatter({ value: [0, 0, 10] })
    expect(html).toContain('Σ (Region+Product): <b>70</b>')
    expect(html).toContain('Σ Region(East): <b>30</b>')
    expect(html).toContain('Σ Product(A): <b>40</b>')
    expect(html).not.toContain('Σ (Region+Product+')
  })

  it('lists per-axis marginals before combined x+y total', () => {
    const html = formatter({ value: [0, 0, 10] })
    expect(html.indexOf('Σ Region(East)')).toBeLessThan(html.indexOf('Σ (Region+Product)'))
    expect(html.indexOf('Σ Product(A)')).toBeLessThan(html.indexOf('Σ (Region+Product)'))
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
  })

  it('includes combined x+y and x+y+z totals plus per-axis marginals', () => {
    const html = formatter({ value: [0, 0, 0] })
    expect(html).toContain('Σ (Region+Product): <b>105</b>')
    expect(html).toContain('Σ (Region+Product+Category): <b>120</b>')
    expect(html).toContain('Σ Region(East): <b>55</b>')
    expect(html).toContain('Σ Product(A): <b>50</b>')
  })

  it('includes cell-level z sum when multiple z groups share the cell', () => {
    const html = formatter({ value: [0, 0, 0] })
    expect(html).toContain('Σ Category: <b>15</b>')
  })

  it('lists per-axis marginals and z sum before combined totals', () => {
    const html = formatter({ value: [0, 0, 0] })
    const xIdx = html.indexOf('Σ Region(East)')
    const yIdx = html.indexOf('Σ Product(A)')
    const zIdx = html.indexOf('Σ Category:')
    const xyIdx = html.indexOf('Σ (Region+Product):')
    const xyzIdx = html.indexOf('Σ (Region+Product+Category):')
    expect(xIdx).toBeLessThan(xyIdx)
    expect(yIdx).toBeLessThan(xyIdx)
    expect(zIdx).toBeLessThan(xyIdx)
    expect(xyIdx).toBeLessThan(xyzIdx)
  })

  it('includes x+y+z total for single-z cells', () => {
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
    expect(html).toContain('Σ (Region+Product): <b>20</b>')
    expect(html).toContain('Σ (Region+Product+Category): <b>30</b>')
    expect(html).not.toMatch(/Σ Category: <b>10<\/b><br\/>/)
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
