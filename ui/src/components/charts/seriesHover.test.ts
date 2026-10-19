import { describe, it, expect, vi, afterEach } from 'vitest'
import { createSeriesHover, LEAVE_DELAY_MS } from './seriesHover'

describe('createSeriesHover', () => {
  afterEach(() => vi.useRealTimers())

  const setup = (hasChart = true) => {
    vi.useFakeTimers()
    const calls: Record<string, unknown>[] = []
    const blockHover: boolean[] = []
    const chart = { dispatchAction: (p: Record<string, unknown>) => void calls.push(p) }
    const hover = createSeriesHover(
      () => (hasChart ? chart : null),
      (on) => blockHover.push(on)
    )
    return { calls, blockHover, hover }
  }
  const block = { componentType: 'series', seriesIndex: 2, dataIndex: 1 }

  it('highlights the whole series and flags block hover', () => {
    const { calls, blockHover, hover } = setup()
    hover.onMouseover(block)
    expect(calls).toEqual([{ type: 'downplay' }, { type: 'highlight', seriesIndex: 2 }])
    expect(blockHover).toEqual([true])
  })

  it('restores the column highlight and clears block hover after the leave delay', () => {
    const { calls, blockHover, hover } = setup()
    hover.onMouseover(block)
    calls.length = 0
    hover.onMouseout(block)
    vi.advanceTimersByTime(LEAVE_DELAY_MS - 1)
    expect(calls).toEqual([])
    vi.advanceTimersByTime(1)
    expect(calls).toEqual([{ type: 'downplay' }, { type: 'highlight', dataIndex: 1 }])
    expect(blockHover).toEqual([true, false])
  })

  it('keeps block hover while crossing a gap to another bar (no tooltip flash)', () => {
    const { calls, blockHover, hover } = setup()
    hover.onMouseover({ componentType: 'series', seriesIndex: 0, dataIndex: 1 })
    hover.onMouseout({ componentType: 'series', seriesIndex: 0, dataIndex: 1 })
    vi.advanceTimersByTime(LEAVE_DELAY_MS / 2)
    hover.onMouseover({ componentType: 'series', seriesIndex: 1, dataIndex: 1 })
    vi.runAllTimers()
    expect(blockHover).toEqual([true, true])
    expect(calls).toEqual([
      { type: 'downplay' },
      { type: 'highlight', seriesIndex: 0 },
      { type: 'downplay' },
      { type: 'highlight', seriesIndex: 1 },
    ])
  })

  it('does not re-dispatch for a redraw mouseout + mouseover on the same block', () => {
    const { calls, hover } = setup()
    hover.onMouseover(block)
    calls.length = 0
    hover.onMouseout(block)
    hover.onMouseover(block)
    vi.runAllTimers()
    expect(calls).toEqual([])
  })

  it('only downplays on leave when the event has no data index', () => {
    const { calls, hover } = setup()
    hover.onMouseout({ componentType: 'series', seriesIndex: 0 })
    vi.runAllTimers()
    expect(calls).toEqual([{ type: 'downplay' }])
  })

  it('leaving the chart cancels the pending column restore and clears highlight', () => {
    const { calls, blockHover, hover } = setup()
    hover.onMouseover(block)
    hover.onMouseout(block)
    calls.length = 0
    hover.onGlobalout()
    vi.runAllTimers()
    expect(calls).toEqual([{ type: 'downplay' }])
    expect(blockHover.at(-1)).toBe(false)
  })

  it('ignores non-series events and a missing chart', () => {
    const { calls, hover } = setup()
    hover.onMouseover({ componentType: 'legend' })
    hover.onMouseout({ componentType: 'series' })
    vi.runAllTimers()
    expect(calls).toEqual([])
    const none = setup(false)
    expect(() => {
      none.hover.onMouseover(block)
      none.hover.onMouseout(block)
      none.hover.onGlobalout()
    }).not.toThrow()
  })

  it('skips the deferred restore if the chart is gone by then', () => {
    vi.useFakeTimers()
    let chart: { dispatchAction: () => void } | null = { dispatchAction: vi.fn() }
    const hover = createSeriesHover(
      () => chart,
      () => {}
    )
    hover.onMouseout(block)
    chart = null
    expect(() => vi.runAllTimers()).not.toThrow()
  })
})
