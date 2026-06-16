import { describe, it, expect, vi } from 'vitest'
import type { BarConfig, LineConfig, PieConfig, HeatmapConfig, RadarConfig } from '../../types'

// fieldRegistry imports the .vue control components directly. The vitest config
// intentionally excludes the Vue plugin (pure-function tests only, per project
// convention), so we stub the .vue files with placeholder objects. The registry's
// shape + appliesTo matrix + getRenderableFields logic are what this suite
// actually exercises; the .vue bodies are exercised by the runtime / browser.
vi.mock('../../components/settings/SortControl.vue', () => ({ default: { name: 'SortControl' } }))
vi.mock('../../components/settings/ScaleControl.vue', () => ({ default: { name: 'ScaleControl' } }))
vi.mock('../../components/settings/ShowLabelsControl.vue', () => ({ default: { name: 'ShowLabelsControl' } }))
vi.mock('../../components/settings/AutoRotateControl.vue', () => ({ default: { name: 'AutoRotateControl' } }))
vi.mock('../../components/settings/SwapControl.vue', () => ({ default: { name: 'SwapControl' } }))

const { fieldRegistry, getControl, getRenderableFields } = await import('./fieldRegistry')

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

  it('autoRotate is constrained to 3D data', () => {
    // autoRotate writes grid3D.viewControl.autoRotate — only meaningful when
    // the chart actually renders as 3D (i.e. z axis is present in the data).
    expect(fieldRegistry['autoRotate']!.appliesOn).toEqual(['3D'])
  })

  it('fields with no appliesOn have no dimension constraint', () => {
    // Most fields (sort, scale, showLabels, swap) are available on any
    // dimension. Their `appliesOn` is undefined — the registry treats that as
    // "no constraint".
    for (const key of ['sort', 'scale', 'showLabels', 'swap'] as const) {
      expect(fieldRegistry[key]!.appliesOn).toBeUndefined()
    }
  })

  it('sort, showLabels, and swap apply to all five chart types', () => {
    for (const key of ['sort', 'showLabels', 'swap'] as const) {
      expect(fieldRegistry[key]!.appliesTo).toEqual(['bar', 'line', 'pie', 'heatmap', 'radar'])
    }
  })
})

describe('getRenderableFields', () => {
  it('returns 5 entries for a 3D bar config (sort/scale/showLabels/autoRotate/swap)', () => {
    const cfg: BarConfig = { type: 'bar' }
    const fields = getRenderableFields(cfg, { dimension: '3D' })
    expect(fields.map((f) => f.key)).toEqual(['sort', 'scale', 'showLabels', 'autoRotate', 'swap'])
    for (const f of fields) expect(f.component).toBeDefined()
  })

  it('returns 5 entries for a 3D line config', () => {
    const cfg: LineConfig = { type: 'line' }
    expect(getRenderableFields(cfg, { dimension: '3D' }).map((f) => f.key)).toEqual([
      'sort',
      'scale',
      'showLabels',
      'autoRotate',
      'swap',
    ])
  })

  it('returns 4 entries for a 2D bar config (no autoRotate — 3D-only)', () => {
    const cfg: BarConfig = { type: 'bar' }
    expect(getRenderableFields(cfg, { dimension: '2D' }).map((f) => f.key)).toEqual([
      'sort',
      'scale',
      'showLabels',
      'swap',
    ])
  })

  it('returns 4 entries for a 2D line config (no autoRotate)', () => {
    const cfg: LineConfig = { type: 'line' }
    expect(getRenderableFields(cfg, { dimension: '2D' }).map((f) => f.key)).toEqual([
      'sort',
      'scale',
      'showLabels',
      'swap',
    ])
  })

  it('returns 3 entries for a pie config (no scale/autoRotate; dimension is irrelevant)', () => {
    const cfg: PieConfig = { type: 'pie' }
    expect(getRenderableFields(cfg, { dimension: '2D' }).map((f) => f.key)).toEqual([
      'sort',
      'showLabels',
      'swap',
    ])
  })

  it('returns 3 entries for a heatmap config (no scale/autoRotate; dimension is irrelevant)', () => {
    const cfg: HeatmapConfig = { type: 'heatmap' }
    expect(getRenderableFields(cfg, { dimension: '2D' }).map((f) => f.key)).toEqual([
      'sort',
      'showLabels',
      'swap',
    ])
  })

  it('returns 3 entries for a radar config (no scale/autoRotate; dimension is irrelevant)', () => {
    const cfg: RadarConfig = { type: 'radar' }
    expect(getRenderableFields(cfg, { dimension: '2D' }).map((f) => f.key)).toEqual([
      'sort',
      'showLabels',
      'swap',
    ])
  })

  it('treats an unknown dimension (no ctx) as "no dimension constraint" — shows all applicable fields', () => {
    // The panel passes `activeDataDimension` which is `undefined` until the
    // dataset loads. The dimension constraint must be skipped in that case so
    // the panel still shows every field by default.
    const cfg: BarConfig = { type: 'bar' }
    expect(getRenderableFields(cfg).map((f) => f.key)).toEqual([
      'sort',
      'scale',
      'showLabels',
      'autoRotate',
      'swap',
    ])
  })

  it('renders all applicable fields even when most keys are absent from the config', () => {
    // The panel must show every available field, not just the keys the user
    // populated. The control components display in their default state when
    // the field is absent.
    const cfg = { type: 'bar', sort: { enabled: false, order: 'asc' as const } } as unknown as BarConfig
    expect(getRenderableFields(cfg, { dimension: '3D' }).map((f) => f.key)).toEqual([
      'sort',
      'scale',
      'showLabels',
      'autoRotate',
      'swap',
    ])
  })
})
