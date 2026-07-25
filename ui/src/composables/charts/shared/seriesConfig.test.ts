import { describe, it, expect } from 'vitest'
import {
  resolveSeriesSymbol,
  resolve3DSymbolSize,
  resolve3DSymbolProps,
  createPieLabelConfig,
  createPieSeriesConfig,
} from './seriesConfig'
import { fontSize } from './common'

describe('resolveSeriesSymbol', () => {
  it('returns defaults when overrides absent', () => {
    expect(resolveSeriesSymbol({ symbol: 'circle', symbolSize: 8 })).toEqual({
      symbol: 'circle',
      symbolSize: 8,
    })
  })

  it('applies symbol and symbolSize overrides', () => {
    expect(
      resolveSeriesSymbol({ symbol: 'circle', symbolSize: 8, sampling: 'lttb' }, 'diamond', 12)
    ).toEqual({
      symbol: 'diamond',
      symbolSize: 12,
      sampling: 'lttb',
    })
  })

  it('ignores empty symbol string but accepts symbolSize 0', () => {
    expect(resolveSeriesSymbol({ symbol: 'circle', symbolSize: 8 }, '', 0)).toEqual({
      symbol: 'circle',
      symbolSize: 0,
    })
  })
})

describe('resolve3DSymbolSize', () => {
  it('prefers override when provided', () => {
    expect(resolve3DSymbolSize(10, 4)).toBe(4)
    expect(resolve3DSymbolSize(undefined, 4)).toBe(4)
  })

  it('falls back to computed', () => {
    expect(resolve3DSymbolSize(10)).toBe(10)
    expect(resolve3DSymbolSize(undefined)).toBeUndefined()
  })
})

describe('resolve3DSymbolProps', () => {
  it('returns empty object when nothing set', () => {
    expect(resolve3DSymbolProps(undefined)).toEqual({})
  })

  it('includes symbol and computed size', () => {
    expect(resolve3DSymbolProps(9, 'triangle')).toEqual({
      symbol: 'triangle',
      symbolSize: 9,
    })
  })

  it('override size wins over computed', () => {
    expect(resolve3DSymbolProps(9, 'rect', 3)).toEqual({
      symbol: 'rect',
      symbolSize: 3,
    })
  })

  it('omits symbol when falsy and size when both undefined', () => {
    expect(resolve3DSymbolProps(undefined, '')).toEqual({})
    expect(resolve3DSymbolProps(undefined, undefined, undefined)).toEqual({})
  })
})

describe('createPieLabelConfig / createPieSeriesConfig', () => {
  const styling = { textColor: '#abc' }

  it('builds label config with optional formatter', () => {
    const fmt = (p: { name: string }) => p.name
    expect(createPieLabelConfig(true, styling, fmt)).toEqual({
      show: true,
      formatter: fmt,
      fontSize,
      color: '#abc',
    })
    expect(createPieLabelConfig(false, styling).show).toBe(false)
  })

  it('builds pie series with default radius/center', () => {
    const data = [{ name: 'a', value: 1 }]
    expect(createPieSeriesConfig('share', data, true, styling)).toMatchObject({
      name: 'share',
      type: 'pie',
      radius: ['40%', '70%'],
      center: ['50%', '50%'],
      data,
      label: { show: true, color: '#abc', fontSize },
    })
  })

  it('accepts custom radius, center, and formatter', () => {
    const fmt = () => 'x'
    const series = createPieSeriesConfig(
      'n',
      [],
      false,
      styling,
      fmt,
      ['0%', '50%'],
      ['40%', '60%']
    )
    expect(series.radius).toEqual(['0%', '50%'])
    expect(series.center).toEqual(['40%', '60%'])
    expect(series.label.formatter).toBe(fmt)
    expect(series.label.show).toBe(false)
  })
})
