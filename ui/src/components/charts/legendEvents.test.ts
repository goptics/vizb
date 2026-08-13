import { describe, it, expect, vi } from 'vitest'
import {
  isLegendSelectChangedEvent,
  createLegendSelectChangedForwarder,
  type LegendSelectChangedEvent,
} from './legendEvents'

describe('isLegendSelectChangedEvent', () => {
  it('accepts a record of booleans', () => {
    expect(isLegendSelectChangedEvent({ selected: { z1: false, z2: true } })).toBe(true)
    expect(isLegendSelectChangedEvent({ selected: {} })).toBe(true)
  })

  it('rejects non-record, non-boolean, or missing selected', () => {
    expect(isLegendSelectChangedEvent(undefined)).toBe(false)
    expect(isLegendSelectChangedEvent(null)).toBe(false)
    expect(isLegendSelectChangedEvent('selected')).toBe(false)
    expect(isLegendSelectChangedEvent({})).toBe(false)
    expect(isLegendSelectChangedEvent({ selected: [true] })).toBe(false)
    expect(isLegendSelectChangedEvent({ selected: { z1: 'yes' } })).toBe(false)
  })
})

describe('createLegendSelectChangedForwarder', () => {
  it('forwards only valid legend-select-changed events', () => {
    const forward = vi.fn()
    const handle = createLegendSelectChangedForwarder(forward)

    const good: LegendSelectChangedEvent = { selected: { z1: false, z2: true } }
    handle(good)
    expect(forward).toHaveBeenCalledTimes(1)
    expect(forward).toHaveBeenCalledWith(good)

    // dblclick-style payloads with non-boolean selected are dropped.
    handle({ selected: { z1: 'toggle' } })
    handle(null)
    handle({})
    expect(forward).toHaveBeenCalledTimes(1)
  })
})
