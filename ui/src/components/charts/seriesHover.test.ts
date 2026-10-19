import { describe, it, expect } from 'vitest'
import { createSeriesHover } from './seriesHover'

describe('createSeriesHover', () => {
  const setup = (tooltip: unknown = [{ trigger: 'axis' }]) => {
    const calls: Record<string, unknown>[] = []
    const chart = {
      dispatchAction: (p: Record<string, unknown>) => void calls.push(p),
      setOption: (o: Record<string, unknown>) => void calls.push({ setOption: o }),
      getOption: () => ({ tooltip }),
    }
    return { calls, hover: createSeriesHover(() => chart) }
  }

  it('highlights the whole series and hides the group tooltip on block hover', () => {
    const { calls, hover } = setup()
    hover.onMouseover({ componentType: 'series', seriesIndex: 2, dataIndex: 1 })
    expect(calls).toEqual([
      { type: 'downplay' },
      { type: 'highlight', seriesIndex: 2 },
      { type: 'hideTip' },
      { setOption: { tooltip: { show: false } } },
    ])
  })

  it('restores the column highlight and tooltip on leave', () => {
    const { calls, hover } = setup()
    hover.onMouseover({ componentType: 'series', seriesIndex: 2, dataIndex: 1 })
    calls.length = 0
    hover.onMouseout({ componentType: 'series', seriesIndex: 2, dataIndex: 1 })
    expect(calls).toEqual([
      { type: 'downplay' },
      { type: 'highlight', dataIndex: 1 },
      { setOption: { tooltip: { show: true } } },
    ])
  })

  it('keeps item tooltips (value mode) and does not restore what it never hid', () => {
    const { calls, hover } = setup({ trigger: 'item' })
    hover.onMouseover({ componentType: 'series', seriesIndex: 0, dataIndex: 0 })
    hover.onMouseout({ componentType: 'series', seriesIndex: 0 })
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
    expect(calls).toEqual([])
    const none = createSeriesHover(() => null)
    expect(() => none.onMouseover({ componentType: 'series', seriesIndex: 0 })).not.toThrow()
    expect(() => none.onMouseout({ componentType: 'series', seriesIndex: 0 })).not.toThrow()
  })
})
