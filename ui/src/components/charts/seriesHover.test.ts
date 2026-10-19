import { describe, it, expect, vi, afterEach } from 'vitest'
import { createSeriesHover } from './seriesHover'

describe('createSeriesHover', () => {
  afterEach(() => vi.useRealTimers())

  const setup = (tooltip: unknown = [{ trigger: 'axis' }]) => {
    vi.useFakeTimers()
    const calls: Record<string, unknown>[] = []
    const chart = {
      dispatchAction: (p: Record<string, unknown>) => void calls.push(p),
      setOption: (o: Record<string, unknown>) => void calls.push({ setOption: o }),
      getOption: () => ({ tooltip }),
    }
    return { calls, hover: createSeriesHover(() => chart) }
  }
  const block = { componentType: 'series', seriesIndex: 2, dataIndex: 1 }

  it('highlights the whole series and hides the group tooltip on block hover', () => {
    const { calls, hover } = setup()
    hover.onMouseover(block)
    expect(calls).toEqual([
      { type: 'downplay' },
      { type: 'highlight', seriesIndex: 2 },
      { type: 'hideTip' },
      { setOption: { tooltip: { show: false } } },
    ])
  })

  it('restores the column highlight and tooltip after leaving', () => {
    const { calls, hover } = setup()
    hover.onMouseover(block)
    calls.length = 0
    hover.onMouseout(block)
    expect(calls).toEqual([]) // deferred
    vi.runAllTimers()
    expect(calls).toEqual([
      { type: 'downplay' },
      { type: 'highlight', dataIndex: 1 },
      { setOption: { tooltip: { show: true } } },
    ])
  })

  it('treats a redraw mouseout + mouseover on the same block as staying hovered', () => {
    const { calls, hover } = setup()
    hover.onMouseover(block)
    calls.length = 0
    hover.onMouseout(block)
    hover.onMouseover(block)
    vi.runAllTimers()
    expect(calls).toEqual([])
  })

  it('re-highlights when moving to a block of another series', () => {
    const { calls, hover } = setup()
    hover.onMouseover(block)
    calls.length = 0
    hover.onMouseout(block)
    hover.onMouseover({ componentType: 'series', seriesIndex: 0, dataIndex: 1 })
    vi.runAllTimers()
    expect(calls).toEqual([{ type: 'downplay' }, { type: 'highlight', seriesIndex: 0 }])
  })

  it('keeps item tooltips (value mode) and does not restore what it never hid', () => {
    const { calls, hover } = setup({ trigger: 'item' })
    hover.onMouseover({ componentType: 'series', seriesIndex: 0, dataIndex: 0 })
    hover.onMouseout({ componentType: 'series', seriesIndex: 0 })
    vi.runAllTimers()
    expect(calls).toEqual([
      { type: 'downplay' },
      { type: 'highlight', seriesIndex: 0 },
      { type: 'downplay' },
    ])
  })

  it('treats a missing option as no axis tooltip', () => {
    const { calls, hover } = setup(null)
    hover.onMouseover({ componentType: 'series', seriesIndex: 0 })
    expect(calls).toHaveLength(2)
  })

  it('ignores non-series events and a missing chart', () => {
    const { calls, hover } = setup()
    hover.onMouseover({ componentType: 'legend' })
    hover.onMouseout({ componentType: 'series' })
    vi.runAllTimers()
    expect(calls).toEqual([])
    const none = createSeriesHover(() => null)
    expect(() => none.onMouseover(block)).not.toThrow()
    expect(() => none.onMouseout(block)).not.toThrow()
  })

  it('skips the deferred leave if the chart is gone by then', () => {
    let chart: {
      dispatchAction: () => void
      setOption: () => void
      getOption: () => object
    } | null = { dispatchAction: vi.fn(), setOption: vi.fn(), getOption: () => ({}) }
    vi.useFakeTimers()
    const hover = createSeriesHover(() => chart)
    hover.onMouseout(block)
    chart = null
    expect(() => vi.runAllTimers()).not.toThrow()
  })
})
