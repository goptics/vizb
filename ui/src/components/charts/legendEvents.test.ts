import { describe, it, expect, vi } from 'vitest'
import { createLegendSelectChangedForwarder, type LegendSelectChangedEvent } from './legendEvents'

describe('createLegendSelectChangedForwarder', () => {
  it('forwards only valid legend-select-changed events', () => {
    const forward = vi.fn()
    const handle = createLegendSelectChangedForwarder(forward)

    const good: LegendSelectChangedEvent = { selected: { z1: false, z2: true } }
    handle(good)
    handle({ selected: {} })
    expect(forward).toHaveBeenCalledTimes(2)
    expect(forward).toHaveBeenNthCalledWith(1, good)
    expect(forward).toHaveBeenNthCalledWith(2, { selected: {} })

    // dblclick-style payloads with non-boolean selected are dropped.
    handle({ selected: { z1: 'toggle' } })
    handle({ selected: [true] })
    handle(null)
    handle(undefined)
    handle('selected')
    handle({})
    expect(forward).toHaveBeenCalledTimes(2)
  })
})
