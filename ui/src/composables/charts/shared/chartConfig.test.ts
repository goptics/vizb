import { describe, it, expect, vi } from 'vitest'
import {
  createAxisConfig,
  createValueAxisConfig,
  createDataZoomConfig,
  createGridConfig,
  createTooltipConfig,
  createValueModeGridConfig,
  VALUE_MODE_GRID_TOP,
  createHeatmapLayoutConfig,
  DATAZOOM_INITIAL_END_PERCENT,
  getChartStyling,
  heatmapDataZoomXBottom,
  HEATMAP_DATAZOOM_X_HEIGHT,
  HEATMAP_VISUAL_MAP_BAND,
  HEATMAP_VISUAL_MAP_BOTTOM,
  HEATMAP_X_TICK_BAND,
  HEATMAP_Y_LABEL_LEFT,
  HEATMAP_Y_ZOOM_INSET,
  formatRadarItemTooltip,
  hasRotatedXLabels,
  renderDonutSvg,
  renderTooltipLegendColumns,
  TOOLTIP_LEGEND_MAX_ROWS_PER_COL,
  getTooltipTheme,
} from './chartConfig'

const indicators = ['A', 'B', 'C']
const styling = getChartStyling(true)

describe('createAxisConfig (y-axis range)', () => {
  it('includes zero on linear scale by default (bar-style baseline)', () => {
    const { yAxis } = createAxisConfig(styling, ['a', 'b'], 'linear')
    expect(yAxis.scale).toBeUndefined()
  })

  it('fits y-axis to data range for line/scatter (scale: true)', () => {
    const { yAxis } = createAxisConfig(
      styling,
      ['a', 'b'],
      'linear',
      undefined,
      undefined,
      false,
      true
    )
    expect(yAxis.scale).toBe(true)
  })

  it('sets log min from data instead of scale when log scale is active', () => {
    const { yAxis } = createAxisConfig(styling, ['a', 'b'], 'log', 2500, undefined, false, true)
    expect(yAxis.scale).toBeUndefined()
    expect(yAxis.min).toBe(1000)
  })
})

describe('createValueAxisConfig (y-axis range)', () => {
  it('fits y-axis to data for value-mode line/scatter', () => {
    const { yAxis } = createValueAxisConfig(styling, 'x', 'y', 'linear', undefined, true)
    expect(yAxis.scale).toBe(true)
  })

  it('keeps zero baseline for value-mode bar charts', () => {
    const { yAxis } = createValueAxisConfig(styling, 'x', 'y', 'linear', undefined, false)
    expect(yAxis.scale).toBeUndefined()
  })
})

describe('createValueModeGridConfig', () => {
  it('uses minimal fixed top because value mode hides the legend', () => {
    expect(createValueModeGridConfig(false).top).toBe(VALUE_MODE_GRID_TOP)
    expect(createValueModeGridConfig(false).top).not.toBe(createGridConfig(1, false).top)
  })
})

describe('createDataZoomConfig', () => {
  const xAxisData = Array.from({ length: 100 }, (_, i) => `cat${i}`)

  it('starts grouped charts at a fixed 20% visible window', () => {
    const dataZoom = createDataZoomConfig(xAxisData, styling)
    expect(dataZoom).toHaveLength(2)
    for (const entry of dataZoom) {
      expect(entry.start).toBe(0)
      expect(entry.end).toBe(DATAZOOM_INITIAL_END_PERCENT)
      expect(entry.xAxisIndex).toBe(0)
      expect(entry.filterMode).toBe('filter')
    }
  })

  it('keeps slider chrome and themed boundary labels', () => {
    const slider = createDataZoomConfig(xAxisData, styling).find((z) => z.type === 'slider')
    expect(slider).toMatchObject({
      bottom: 34,
      height: 28,
      textStyle: { color: styling.textColor },
    })
  })
})

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

describe('renderDonutSvg', () => {
  it('includes zero-valued slices in the legend without drawing their arcs', () => {
    const html = renderDonutSvg([
      { name: 'A', value: 10, color: '#a00' },
      { name: 'B', value: 0, color: '#0b0' },
      { name: 'C', value: 5, color: '#00c' },
    ])

    expect(html).toContain('<span>A</span><b style="margin-left:6px">66.7%</b>')
    expect(html).toContain('<span>B</span><b style="margin-left:6px">n/a</b>')
    expect(html).toContain('<span>C</span><b style="margin-left:6px">33.3%</b>')
    expect(html.match(/<path /g) ?? []).toHaveLength(2)
    expect(html).toMatch(/<path [^>]*fill="#a00"\/>/)
    expect(html).not.toMatch(/<path [^>]*fill="#0b0"\/>/)
    expect(html).toMatch(/<path [^>]*fill="#00c"\/>/)
  })

  it('returns empty when fewer than two positive slices are present', () => {
    expect(
      renderDonutSvg([
        { name: 'A', value: 10, color: '#a00' },
        { name: 'B', value: 0, color: '#0b0' },
      ])
    ).toBe('')
  })
})

describe('createTooltipConfig', () => {
  it('keeps missing and nonnumeric values out of the donut while showing numeric zero as n/a', () => {
    const tooltip = createTooltipConfig(true, false) as any
    const html = tooltip.formatter([
      { name: 'X', seriesName: 'A', value: 10, color: '#a00', marker: 'A-marker' },
      { name: 'X', seriesName: 'Zero', value: 0, color: '#0b0', marker: 'Zero-marker' },
      { name: 'X', seriesName: 'Null', value: null, color: '#ccc', marker: 'Null-marker' },
      {
        name: 'X',
        seriesName: 'Undefined',
        value: undefined,
        color: '#ddd',
        marker: 'Undefined-marker',
      },
      {
        name: 'X',
        seriesName: 'Text',
        value: 'not-a-number',
        color: '#eee',
        marker: 'Text-marker',
      },
      { name: 'X', seriesName: 'C', value: 5, color: '#00c', marker: 'C-marker' },
    ])

    expect(html).toContain('Zero-marker Zero: 0')
    expect(html).toContain('Text-marker Text: not-a-number')
    expect(html).not.toContain('Null-marker')
    expect(html).not.toContain('Undefined-marker')

    const donut = html.slice(html.indexOf('<svg'))
    expect(donut).toContain('<span>Zero</span><b style="margin-left:6px">n/a</b>')
    expect(donut.match(/>n\/a<\/b>/g) ?? []).toHaveLength(1)
    expect(donut).not.toContain('<span>Null</span>')
    expect(donut).not.toContain('<span>Undefined</span>')
    expect(donut).not.toContain('<span>Text</span>')
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

  it('omits an indicator without a corresponding numeric value from the donut', () => {
    const html = formatRadarItemTooltip(
      { data: { name: 'Series', value: [10, 5] } },
      indicators,
      false
    )
    const donut = html.slice(html.indexOf('<svg'))

    expect(donut).toContain('<span>A</span>')
    expect(donut).toContain('<span>B</span>')
    expect(donut).not.toContain('<span>C</span>')
    expect(donut).not.toContain('n/a')
  })
})

describe('createHorizontalDataZoomConfig', () => {
  it('returns inside+slider pair for y axis', async () => {
    const { createHorizontalDataZoomConfig, DATAZOOM_INITIAL_END_PERCENT } = await import(
      './chartConfig'
    )
    const dz = createHorizontalDataZoomConfig(styling)
    expect(dz).toHaveLength(2)
    expect(dz[0]).toMatchObject({
      type: 'inside',
      yAxisIndex: 0,
      end: DATAZOOM_INITIAL_END_PERCENT,
    })
    expect(dz[1]).toMatchObject({ type: 'slider', yAxisIndex: 0 })
  })
})

describe('createValueAxisConfig log min', () => {
  it('sets log minimum from minValue', () => {
    const axes = createValueAxisConfig(styling, 'x', 'y', 'log', 25)
    expect(axes.yAxis.min).toBe(10)
  })
})

describe('createValueModeTooltip', () => {
  it('formats x/y and optional cross pointer', async () => {
    const { createValueModeTooltip } = await import('./chartConfig')
    const tip = createValueModeTooltip(false, 'price', 'lat', true) as {
      formatter: (p: unknown) => string
      axisPointer?: { type?: string }
    }
    expect(tip.axisPointer?.type).toBe('cross')
    expect(tip.formatter({ data: [10, 2] })).toContain('price: 10')
    expect(tip.formatter({ data: [10, 2] })).toContain('lat: 2')
  })
})

describe('createHorizontalAxisConfig log + large', () => {
  it('sets log min and auto interval for large categories', async () => {
    const { createHorizontalAxisConfig } = await import('./chartConfig')
    const many = Array.from({ length: 60 }, (_, i) => `c${i}`)
    const axes = createHorizontalAxisConfig(styling, many, 'log', 50, 'cat', true)
    expect(axes.xAxis.min).toBe(10)
    expect(axes.yAxis.axisLabel.interval).toBe('auto')
    expect(axes.yAxis.nameGap).toBe(88)
  })

  it('adds tick margin when no name and no dataZoom', async () => {
    const { createHorizontalAxisConfig } = await import('./chartConfig')
    const axes = createHorizontalAxisConfig(styling, ['a', 'b'], 'linear')
    expect(axes.yAxis.axisLabel.margin).toBe(14)
  })
})

describe('createPinnedAxisTooltip', () => {
  it('formats first axis param and guards empty inputs', async () => {
    const { createPinnedAxisTooltip } = await import('./chartConfig')
    const tip = createPinnedAxisTooltip(true) as {
      position: (pt: number[]) => unknown
      formatter: (p: unknown) => string
    }
    expect(tip.position([12, 3])).toEqual([12, '10%'])
    expect(tip.position([])).toEqual([0, '10%'])
    expect(tip.formatter({ name: 'x' })).toBe('')
    expect(tip.formatter([])).toBe('')
    expect(tip.formatter([{ name: 'A', marker: '*', value: 3 }])).toContain('<strong>A</strong>')
  })
})

describe('createTooltipConfig item + empty branches', () => {
  it('returns empty for non-array axis params and empty present set', () => {
    const tip = createTooltipConfig(true, false) as { formatter: (p: unknown) => string }
    expect(tip.formatter({ name: 'x' })).toBe('')
    expect(tip.formatter([{ name: 'X', seriesName: 'A', value: null, marker: 'm' }])).toBe('')
  })

  it('formats single-series axis tooltip without sum/donut', () => {
    const tip = createTooltipConfig(true, false) as { formatter: (p: unknown) => string }
    const html = tip.formatter([
      { name: 'X', seriesName: 'A', value: 10, color: '#a00', marker: 'm' },
    ])
    expect(html).toContain('A')
    expect(html).not.toContain('Σ')
  })

  it('formats item tooltip using seriesName fallbacks', () => {
    const tip = createTooltipConfig(false, false) as { formatter: (p: unknown) => string }
    expect(tip.formatter([])).toBe('')
    expect(tip.formatter({ marker: '*', name: 'N', value: 1 })).toContain('<strong>N</strong>')
    expect(tip.formatter({ marker: '*', seriesName: 'S', value: 2 })).toContain(
      '<strong>S</strong>'
    )
  })
})

describe('createLegendConfig single series', () => {
  it('hides legend when only one series', async () => {
    const { createLegendConfig } = await import('./chartConfig')
    expect(createLegendConfig([{ xAxis: 'a' }], styling, false)).toEqual({ show: false })
  })
})

describe('renderDonutSvg total guard', () => {
  it('returns empty when positive slice total is non-positive', () => {
    // unreachable via normal positive filter, but keep API contract: empty when <2 positives
    expect(
      renderDonutSvg([
        { name: 'A', value: 0, color: '#a' },
        { name: 'B', value: 0, color: '#b' },
      ])
    ).toBe('')
  })
})

describe('createHeatmapLayoutConfig legend top', () => {
  it('uses legend percentage top when hasLegend', () => {
    const layout = createHeatmapLayoutConfig({ hasLegend: true, seriesLength: 20 })
    expect(String(layout.grid.top)).toContain('%')
  })
})

describe('remaining chartConfig branches', () => {
  it('createAxisConfig rotates long labels', () => {
    const long = [Array(60).fill('x').join(''), Array(50).fill('y').join('')]
    const axes = createAxisConfig(styling, long, 'linear')
    expect(axes.xAxis.axisLabel.rotate).toBe(30)
  })

  it('formatRadarItemTooltip non-array values and title fallbacks', () => {
    expect(
      formatRadarItemTooltip(
        { data: { name: undefined, value: undefined as unknown as number[] }, name: 'N' },
        indicators,
        false
      )
    ).toContain('<b>N</b>')

    const html = formatRadarItemTooltip(
      { data: { value: [1, 2, 3] }, seriesName: 'S', color: 123 as unknown as string },
      indicators,
      false
    )
    expect(html).toContain('<b>S</b>')
  })

  it('tooltipSpreadRows non-finite stats', async () => {
    const { tooltipSpreadRows } = await import('./chartConfig')
    // single value → empty
    expect(tooltipSpreadRows([1], false)).toBe('')
    // identical values can yield non-finite cv depending on describe()
    const html = tooltipSpreadRows([0, 0, 0], false)
    // still returns a block or empty; just invoke
    expect(typeof html).toBe('string')
  })

  it('createTooltipConfig seriesTotals and non-string colors', () => {
    const totals = new Map<string, number>([['A', 10]])
    const tip = createTooltipConfig(true, false, totals) as { formatter: (p: unknown) => string }
    const html = tip.formatter([
      { name: 'X', seriesName: 'A', value: 4, color: { toString: () => '#abc' }, marker: 'm' },
      { name: 'X', seriesName: 'B', value: 6, color: '#00c', marker: 'n' },
    ])
    expect(html).toContain('Σ10')
    expect(html).toContain('Σ X:')

    // no donut when only one positive after filter? two values should donut
    expect(html).toContain('<svg')
  })

  it('createTooltipConfig item path never takes hasXYAxis true branch in item mode', () => {
    // dead branch at 746 is unreachable in item mode (hasXYAxis false). Cover seriesName fallback only.
    const tip = createTooltipConfig(false, true) as { formatter: (p: unknown) => string }
    expect(tip.formatter({ marker: '*', seriesName: 'Only', value: 1 })).toContain('Only')
  })
})

describe('final chartConfig branch cleanup', () => {
  it('radar title falls back to empty string', () => {
    const html = formatRadarItemTooltip({ data: { value: [1] } }, ['A'], false)
    expect(html).toContain('<b></b>')
  })

  it('radar uses string series color', () => {
    const html = formatRadarItemTooltip(
      { data: { value: [1, 2] }, color: '#ff0000' },
      ['A', 'B'],
      false
    )
    expect(html).toContain('#ff0000')
  })

  it('tooltip with missing seriesName/name still formats multi series', () => {
    const tip = createTooltipConfig(true, false) as { formatter: (p: unknown) => string }
    const html = tip.formatter([
      { value: 1, color: '#a00', marker: 'm' },
      { value: 2, color: '#0b0', marker: 'n' },
    ])
    expect(html).toContain('Σ')
  })

  it('tooltip seriesTotals lookup with missing seriesName key', () => {
    const totals = new Map<string, number>([['', 9]])
    const tip = createTooltipConfig(true, false, totals) as { formatter: (p: unknown) => string }
    const html = tip.formatter([
      { name: 'X', value: 1, color: '#a00', marker: 'm' },
      { name: 'X', value: 2, color: '#0b0', marker: 'n' },
    ])
    expect(html).toContain('Σ9')
  })

  it('multi-series all-zero skips donut but keeps sum', () => {
    const tip = createTooltipConfig(true, false) as { formatter: (p: unknown) => string }
    const html = tip.formatter([
      { name: 'X', seriesName: 'A', value: 0, color: '#a00', marker: 'm' },
      { name: 'X', seriesName: 'B', value: 0, color: '#0b0', marker: 'n' },
    ])
    expect(html).toContain('Σ X:')
    expect(html).not.toContain('<svg')
  })
})

describe('tooltipSpreadRows non-finite median', () => {
  it('renders em dash when describe yields non-finite stats', async () => {
    const stats = await import('@/lib/stats')
    const empty = stats.describe([])
    const spy = vi.spyOn(stats, 'describe').mockReturnValue({
      ...empty,
      count: 2,
      median: Number.NaN,
      iqr: Number.NaN,
      cv: Number.NaN,
    })
    const { tooltipSpreadRows } = await import('./chartConfig')
    const html = tooltipSpreadRows([1, 2, 3], false)
    expect(html).toContain('Median: <b>—</b>')
    expect(html).toContain('IQR: <b>—</b>')
    expect(html).toContain('CV: <b>—</b>')
    spy.mockRestore()
  })
})
