import { describe, it, expect } from 'vitest'
import {
  boxSizeForAxisCount,
  create3DGridConfig,
  create3DTooltipFormatter,
  create3DVisualMap,
  createValue3DTooltipFormatter,
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
  const series = [{ data: [{ value: [0, 0, 42] }] }]

  it('returns config when enabled', () => {
    expect(resolve3DVisualMap(true, series, styling)).toMatchObject({ show: true, max: 42 })
  })

  it('returns empty array when disabled so ECharts can replace-merge it away', () => {
    expect(resolve3DVisualMap(false, series, styling)).toEqual([])
  })
})
