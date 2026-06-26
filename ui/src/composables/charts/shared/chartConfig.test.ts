import { describe, it, expect } from 'vitest'
import {
  createGridConfig,
  createHeatmapLayoutConfig,
  heatmapDataZoomXBottom,
  HEATMAP_DATAZOOM_X_HEIGHT,
  HEATMAP_VISUAL_MAP_BAND,
  HEATMAP_VISUAL_MAP_BOTTOM,
  HEATMAP_X_TICK_BAND,
  HEATMAP_Y_LABEL_LEFT,
  HEATMAP_Y_ZOOM_INSET,
  formatRadarItemTooltip,
  hasRotatedXLabels,
  renderTooltipLegendColumns,
  TOOLTIP_LEGEND_MAX_ROWS_PER_COL,
  getTooltipTheme,
} from './chartConfig'

const indicators = ['A', 'B', 'C']

describe('createGridConfig', () => {
  it('reserves fixed px bottom only when dataZoom is present', () => {
    expect(createGridConfig(1, true).bottom).toBe(100)
    expect(createGridConfig(1, true).containLabel).toBe(false)
  })

  it('uses containLabel for no-dataZoom layout (axis title space is automatic)', () => {
    expect(createGridConfig(1, false).bottom).toBe(28)
    expect(createGridConfig(1, false).containLabel).toBe(true)
  })

  it('keeps dataZoom bottom larger than the no-zoom tier', () => {
    expect(createGridConfig(1, true).bottom).toBeGreaterThan(createGridConfig(1, false).bottom)
  })
})

describe('createHeatmapLayoutConfig', () => {
  it('reserves visualMap + tick band only when dataZoom is absent', () => {
    const layout = createHeatmapLayoutConfig({ compact: true })
    expect(layout.visualMapBottom).toBe(HEATMAP_VISUAL_MAP_BOTTOM)
    expect(layout.dataZoomXBottom).toBeUndefined()
    expect(layout.grid.bottom).toBe(
      HEATMAP_VISUAL_MAP_BOTTOM + HEATMAP_VISUAL_MAP_BAND + HEATMAP_X_TICK_BAND
    )
    expect(layout.grid.containLabel).toBe(true)
    expect(layout.grid.left).toBe(8)
    expect(layout.grid.right).toBe(8)
  })

  it('stacks dataZoom above visualMap and enlarges bottom when x dataZoom is present', () => {
    const layout = createHeatmapLayoutConfig({ hasXDataZoom: true, hasYDataZoom: true })
    expect(layout.dataZoomXBottom).toBe(heatmapDataZoomXBottom())
    expect(layout.dataZoomXBottom).toBeGreaterThan(
      HEATMAP_VISUAL_MAP_BOTTOM + HEATMAP_VISUAL_MAP_BAND
    )
    expect(layout.grid.bottom).toBe(
      heatmapDataZoomXBottom() + HEATMAP_DATAZOOM_X_HEIGHT + HEATMAP_X_TICK_BAND
    )
    expect(layout.grid.containLabel).toBe(false)
    expect(layout.grid.bottom).toBeGreaterThan(
      createHeatmapLayoutConfig({ compact: true }).grid.bottom
    )
  })

  it('reserves fixed left/right for y-axis dataZoom slider', () => {
    const layout = createHeatmapLayoutConfig({ hasYDataZoom: true })
    expect(layout.grid.left).toBe(HEATMAP_Y_LABEL_LEFT)
    expect(layout.grid.right).toBe(HEATMAP_Y_ZOOM_INSET)
  })
})

describe('hasRotatedXLabels', () => {
  it('returns false for large axes (dataZoom handles navigation)', () => {
    expect(hasRotatedXLabels(['a'.repeat(60)], true)).toBe(false)
  })

  it('returns true when total label length exceeds threshold on small axes', () => {
    expect(
      hasRotatedXLabels([Array(60).fill('x').join(''), Array(50).fill('y').join('')], false)
    ).toBe(true)
  })
})

describe('getTooltipTheme', () => {
  it('enables enterable tooltips with selectable text', () => {
    const theme = getTooltipTheme(false)
    expect(theme.enterable).toBe(true)
    expect(theme.extraCssText).toContain('user-select:text')
  })
})

describe('renderTooltipLegendColumns', () => {
  it('returns empty string for no rows', () => {
    expect(renderTooltipLegendColumns([])).toBe('')
  })

  it('joins up to max rows in a single column', () => {
    const rows = Array.from({ length: TOOLTIP_LEGEND_MAX_ROWS_PER_COL }, (_, i) => `row${i}`)
    const html = renderTooltipLegendColumns(rows)
    expect(html).toBe(rows.join('<br/>'))
    expect(html).not.toContain('display:grid')
  })

  it('flows into balanced columns when count exceeds threshold', () => {
    const rows = Array.from({ length: 11 }, (_, i) => `row${i}`)
    const html = renderTooltipLegendColumns(rows)
    expect(html).toContain('display:grid')
    expect(html).toContain('grid-auto-flow:column')
    expect(html).toContain('grid-template-rows:repeat(6,auto)')
  })

  it('never exceeds max rows per column for large lists', () => {
    const rows = Array.from({ length: 25 }, (_, i) => `row${i}`)
    const html = renderTooltipLegendColumns(rows)
    expect(html).toContain('grid-template-rows:repeat(9,auto)')
  })
})

describe('formatRadarItemTooltip', () => {
  it('returns empty string when params.data is missing', () => {
    expect(formatRadarItemTooltip({}, indicators, false)).toBe('')
  })

  it('single spoke: rows only, no Σ / spread / donut', () => {
    const html = formatRadarItemTooltip({ data: { name: 'Series', value: [10] } }, ['A'], false)
    expect(html).toContain('<b>Series</b>')
    expect(html).toContain('A: <b>10</b>')
    expect(html).not.toContain('Σ')
    expect(html).not.toContain('Median')
    expect(html).not.toContain('<svg')
  })

  it('multi-spoke: includes Σ, spread stats, and donut', () => {
    const html = formatRadarItemTooltip(
      { data: { name: 'Series', value: [10, 20, 30] } },
      indicators,
      false
    )
    expect(html).toContain('Σ Series: <b>60</b>')
    expect(html).toContain('Median:')
    expect(html).toContain('IQR:')
    expect(html).toContain('CV:')
    expect(html).toContain('<svg')
  })

  it('uses seriesName / data.name header when both differ (X+Y+Z)', () => {
    const html = formatRadarItemTooltip(
      {
        seriesName: 'Pool1',
        data: { name: 'alloc', value: [1, 2] },
      },
      ['Y1', 'Y2'],
      false
    )
    expect(html).toContain('<b>Pool1 / alloc</b>')
  })

  it('uses multi-column grid for many spokes', () => {
    const indicators = Array.from({ length: 11 }, (_, i) => `S${i}`)
    const values = indicators.map((_, i) => i + 1)
    const html = formatRadarItemTooltip(
      { data: { name: 'Series', value: values } },
      indicators,
      false
    )
    expect(html).toContain('display:grid')
    expect(html).toContain('grid-auto-flow:column')
  })
})
