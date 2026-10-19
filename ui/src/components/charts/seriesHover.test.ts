import { describe, it, expect } from 'vitest'
import { createSeriesHover } from './seriesHover'

describe('createSeriesHover', () => {
  const setup = () => {
    const calls: Record<string, unknown>[] = []
    const chart = { dispatchAction: (p: Record<string, unknown>) => void calls.push(p) }
    return { calls, hover: createSeriesHover(() => chart) }
  }

  it('highlights the whole series of the hovered block', () => {
    const { calls, hover } = setup()
    hover.onMouseover({ componentType: 'series', seriesIndex: 2, dataIndex: 1 })
    expect(calls).toEqual([{ type: 'downplay' }, { type: 'highlight', seriesIndex: 2 }])
  })

  it('hands the column back to the axis highlight on leave', () => {
    const { calls, hover } = setup()
    hover.onMouseout({ componentType: 'series', seriesIndex: 2, dataIndex: 1 })
    expect(calls).toEqual([{ type: 'downplay' }, { type: 'highlight', dataIndex: 1 }])
  })

  it('only downplays when the leave event has no data index', () => {
    const { calls, hover } = setup()
    hover.onMouseout({ componentType: 'series', seriesIndex: 0 })
    expect(calls).toEqual([{ type: 'downplay' }])
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
