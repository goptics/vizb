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
})
