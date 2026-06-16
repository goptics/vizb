import { describe, it, expect } from 'vitest'
import type { BarConfig, LineConfig, PieConfig, HeatmapConfig, RadarConfig } from '../../types'
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

  it('scale and autoRotate are restricted to bar/line', () => {
    expect(fieldRegistry['scale']!.appliesTo).toEqual(['bar', 'line'])
    expect(fieldRegistry['autoRotate']!.appliesTo).toEqual(['bar', 'line'])
  })

  it('sort, showLabels, and swap apply to all five chart types', () => {
    for (const key of ['sort', 'showLabels', 'swap'] as const) {
      expect(fieldRegistry[key]!.appliesTo).toEqual(['bar', 'line', 'pie', 'heatmap', 'radar'])
    }
  })
})

describe('getRenderableFields', () => {
  it('returns 5 entries for a bar config (sort/scale/showLabels/autoRotate/swap)', () => {
    const cfg: BarConfig = { type: 'bar' }
    const fields = getRenderableFields(cfg)
    expect(fields.map((f) => f.key)).toEqual(['sort', 'scale', 'showLabels', 'autoRotate', 'swap'])
    for (const f of fields) expect(f.component).toBeDefined()
  })

  it('returns 5 entries for a line config', () => {
    const cfg: LineConfig = { type: 'line' }
    expect(getRenderableFields(cfg).map((f) => f.key)).toEqual([
      'sort',
      'scale',
      'showLabels',
      'autoRotate',
      'swap',
    ])
  })

  it('returns 3 entries for a pie config (no scale/autoRotate)', () => {
    const cfg: PieConfig = { type: 'pie' }
    expect(getRenderableFields(cfg).map((f) => f.key)).toEqual(['sort', 'showLabels', 'swap'])
  })

  it('returns 3 entries for a heatmap config (no scale/autoRotate)', () => {
    const cfg: HeatmapConfig = { type: 'heatmap' }
    expect(getRenderableFields(cfg).map((f) => f.key)).toEqual(['sort', 'showLabels', 'swap'])
  })

  it('returns 3 entries for a radar config (no scale/autoRotate)', () => {
    const cfg: RadarConfig = { type: 'radar' }
    expect(getRenderableFields(cfg).map((f) => f.key)).toEqual(['sort', 'showLabels', 'swap'])
  })

  it('renders all applicable fields even when most keys are absent from the config', () => {
    // The panel must show every available field, not just the keys the user
    // populated. The control components display in their default state when
    // the field is absent.
    const cfg = { type: 'bar', sort: { enabled: false, order: 'asc' as const } } as unknown as BarConfig
    expect(getRenderableFields(cfg).map((f) => f.key)).toEqual([
      'sort',
      'scale',
      'showLabels',
      'autoRotate',
      'swap',
    ])
  })
})
