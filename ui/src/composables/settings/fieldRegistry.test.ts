import { describe, it, expect } from 'vitest'
import type { BarConfig, PieConfig } from '../../types'
import { fieldRegistry, getControl, getRenderableFields } from './fieldRegistry'

describe('fieldRegistry', () => {
  it('exposes the five known field controls', () => {
    expect(Object.keys(fieldRegistry).sort()).toEqual(
      ['autoRotate', 'scale', 'showLabels', 'sort', 'swap'].sort()
    )
  })

  it('getControl returns a component for a registered key', () => {
    expect(getControl('sort')).toBeDefined()
    expect(getControl('scale')).toBeDefined()
    expect(getControl('showLabels')).toBeDefined()
    expect(getControl('autoRotate')).toBeDefined()
    expect(getControl('swap')).toBeDefined()
  })

  it('getControl returns undefined for an unknown key', () => {
    expect(getControl('unknown')).toBeUndefined()
    expect(getControl('type')).toBeUndefined()
  })

  it('getRenderableFields returns one entry per registered key on a bar config (5)', () => {
    const cfg: BarConfig = {
      type: 'bar',
      sort: { enabled: false, order: 'asc' },
      scale: 'linear',
      showLabels: false,
      autoRotate: false,
      swap: '',
    }
    const fields = getRenderableFields(cfg)
    expect(fields.map((f) => f.key)).toEqual(['sort', 'scale', 'showLabels', 'autoRotate', 'swap'])
    for (const f of fields) expect(f.component).toBeDefined()
  })

  it('getRenderableFields returns three entries for a pie config (no scale/autoRotate)', () => {
    const cfg: PieConfig = {
      type: 'pie',
      sort: { enabled: false, order: 'asc' },
      showLabels: false,
      swap: '',
    }
    const fields = getRenderableFields(cfg)
    expect(fields.map((f) => f.key)).toEqual(['sort', 'showLabels', 'swap'])
  })

  it('skips the type discriminator and unknown keys', () => {
    const cfg = {
      type: 'bar',
      sort: { enabled: false, order: 'asc' as const },
      mystery: 42,
    } as unknown as BarConfig
    const fields = getRenderableFields(cfg)
    expect(fields.map((f) => f.key)).toEqual(['sort'])
  })
})
