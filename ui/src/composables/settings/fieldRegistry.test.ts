import { describe, it, expect, vi } from 'vitest'
import type {
  BarConfig,
  LineConfig,
  ScatterConfig,
  PieConfig,
  HeatmapConfig,
  RadarConfig,
} from '@/types'

// fieldRegistry imports the .vue control components directly. The vitest config
// intentionally excludes the Vue plugin (pure-function tests only, per project
// convention), so we stub the .vue files with placeholder objects. The registry's
// shape + appliesTo matrix + getRenderableFields logic are what this suite
// actually exercises; the .vue bodies are exercised by the runtime / browser.
vi.mock('../../components/settings/SortControl.vue', () => ({ default: { name: 'SortControl' } }))
vi.mock('../../components/settings/ScaleControl.vue', () => ({ default: { name: 'ScaleControl' } }))
vi.mock('../../components/settings/ShowLabelsControl.vue', () => ({
  default: { name: 'ShowLabelsControl' },
}))
vi.mock('../../components/settings/ThreeDRotateControl.vue', () => ({
  default: { name: 'ThreeDRotateControl' },
}))
vi.mock('../../components/settings/ThreeDControl.vue', () => ({
  default: { name: 'ThreeDControl' },
}))
vi.mock('../../components/settings/ThreeDVisualMapControl.vue', () => ({
  default: { name: 'ThreeDVisualMapControl' },
}))
vi.mock('../../components/settings/SwapControl.vue', () => ({ default: { name: 'SwapControl' } }))

const { fieldRegistry, getControl, getRenderableFields, partitionRenderableFields } =
  await import('./fieldRegistry')

describe('fieldRegistry', () => {
  it('exposes the seven known field controls', () => {
    expect(Object.keys(fieldRegistry).sort()).toEqual(
      ['threeDRotate', 'scale', 'showLabels', 'sort', 'swap', 'threeD', 'threeDVisualMap'].sort()
    )
  })

  it('getControl returns a component for a registered key', () => {
    expect(getControl('sort')).toBeDefined()
    expect(getControl('scale')).toBeDefined()
    expect(getControl('showLabels')).toBeDefined()
    expect(getControl('threeDRotate')).toBeDefined()
    expect(getControl('swap')).toBeDefined()
  })

  it('getControl returns undefined for an unknown key', () => {
    expect(getControl('unknown')).toBeUndefined()
    expect(getControl('type')).toBeUndefined()
  })

  it('scale and threeDRotate apply to bar, line, and scatter', () => {
    expect(fieldRegistry['scale']!.appliesTo).toEqual(['bar', 'line', 'scatter'])
    expect(fieldRegistry['threeDRotate']!.appliesTo).toEqual(['bar', 'line', 'scatter'])
  })

  it('threeDRotate uses rendering3D visibility', () => {
    expect(fieldRegistry['threeDRotate']!.visible).toBeDefined()
  })

  it('fields with no appliesOn have no dimension constraint', () => {
    // Most fields (sort, scale, showLabels, swap) are available on any
    // dimension. Their `appliesOn` is undefined — the registry treats that as
    // "no constraint".
    for (const key of ['sort', 'scale', 'showLabels', 'swap'] as const) {
      expect(fieldRegistry[key]!.appliesOn).toBeUndefined()
    }
  })

  it('sort, showLabels, and swap apply to all six chart types', () => {
    for (const key of ['sort', 'showLabels', 'swap'] as const) {
      expect(fieldRegistry[key]!.appliesTo).toEqual([
        'bar',
        'line',
        'scatter',
        'pie',
        'heatmap',
        'radar',
      ])
    }
  })
})

describe('getRenderableFields', () => {
  it('returns 6 entries for a 3D bar config (sort/scale/showLabels/threeDVisualMap/threeDRotate/swap)', () => {
    const cfg: BarConfig = { type: 'bar' }
    const fields = getRenderableFields(cfg, { dimension: '3D', rendering3D: true, hasZAxis: true })
    expect(fields.map((f) => f.key)).toEqual([
      'sort',
      'scale',
      'showLabels',
      'threeDVisualMap',
      'threeDRotate',
      'swap',
    ])
    for (const f of fields) expect(f.component).toBeDefined()
  })

  it('returns 6 entries for a 3D line config', () => {
    const cfg: LineConfig = { type: 'line' }
    expect(
      getRenderableFields(cfg, { dimension: '3D', rendering3D: true, hasZAxis: true }).map(
        (f) => f.key
      )
    ).toEqual(['sort', 'scale', 'showLabels', 'threeDVisualMap', 'threeDRotate', 'swap'])
  })

  it('returns 4 entries for a 2D bar config without value-3D active', () => {
    const cfg: BarConfig = { type: 'bar' }
    expect(
      getRenderableFields(cfg, { dimension: '2D', rendering3D: false }).map((f) => f.key)
    ).toEqual(['sort', 'scale', 'showLabels', 'swap'])
  })

  it('returns 6 entries for a 2D bar config with value-3D active', () => {
    const cfg: BarConfig = { type: 'bar', threeD: true }
    expect(
      getRenderableFields(cfg, {
        dimension: '2D',
        rendering3D: true,
        hasThreeDOption: true,
        hasZAxis: false,
      }).map((f) => f.key)
    ).toEqual(['sort', 'scale', 'showLabels', 'threeD', 'threeDVisualMap', 'threeDRotate', 'swap'])
  })

  it('hides threeD when z is on chart axes in the active swap (xyz)', () => {
    const cfg: BarConfig = { type: 'bar', threeD: true }
    expect(
      getRenderableFields(cfg, {
        dimension: '3D',
        rendering3D: true,
        hasThreeDOption: true,
        hasZAxis: true,
      }).map((f) => f.key)
    ).toEqual(['sort', 'scale', 'showLabels', 'threeDVisualMap', 'threeDRotate', 'swap'])
  })

  it('shows threeD toggle without baked threeD when engine is available (xyn swap)', () => {
    const cfg: BarConfig = { type: 'bar' }
    expect(
      getRenderableFields(cfg, {
        dimension: '3D',
        rendering3D: false,
        hasThreeDOption: true,
        hasZAxis: false,
      }).map((f) => f.key)
    ).toEqual(['sort', 'scale', 'showLabels', 'threeD', 'swap'])
  })

  it('hides rotate/visualMap on flat 2D xyn chart until value 3D is enabled', () => {
    const cfg: BarConfig = { type: 'bar' }
    expect(
      getRenderableFields(cfg, {
        dimension: '3D',
        rendering3D: false,
        hasThreeDOption: true,
        hasZAxis: false,
      }).map((f) => f.key)
    ).not.toContain('threeDRotate')
    expect(
      getRenderableFields(cfg, {
        dimension: '3D',
        rendering3D: false,
        hasThreeDOption: true,
        hasZAxis: false,
      }).map((f) => f.key)
    ).not.toContain('threeDVisualMap')
  })

  it('returns 4 entries for a 2D line config without value-3D active', () => {
    const cfg: LineConfig = { type: 'line' }
    expect(
      getRenderableFields(cfg, { dimension: '2D', rendering3D: false }).map((f) => f.key)
    ).toEqual(['sort', 'scale', 'showLabels', 'swap'])
  })

  it('returns 3 entries for a pie config (no scale/threeDRotate; dimension is irrelevant)', () => {
    const cfg: PieConfig = { type: 'pie' }
    expect(getRenderableFields(cfg, { dimension: '2D' }).map((f) => f.key)).toEqual([
      'sort',
      'showLabels',
      'swap',
    ])
  })

  it('returns 3 entries for a heatmap config (no scale/threeDRotate; dimension is irrelevant)', () => {
    const cfg: HeatmapConfig = { type: 'heatmap' }
    expect(getRenderableFields(cfg, { dimension: '2D' }).map((f) => f.key)).toEqual([
      'sort',
      'showLabels',
      'swap',
    ])
  })

  it('returns 3 entries for a radar config (no scale/threeDRotate; dimension is irrelevant)', () => {
    const cfg: RadarConfig = { type: 'radar' }
    expect(getRenderableFields(cfg, { dimension: '2D' }).map((f) => f.key)).toEqual([
      'sort',
      'showLabels',
      'swap',
    ])
  })

  it('treats an unknown dimension (no ctx) as "no dimension constraint" — shows all applicable fields', () => {
    const cfg: BarConfig = { type: 'bar' }
    expect(getRenderableFields(cfg).map((f) => f.key)).toEqual([
      'sort',
      'scale',
      'showLabels',
      'threeDVisualMap',
      'threeDRotate',
      'swap',
    ])
  })

  it('returns scale and 3D fields for a scatter config with value-3D active', () => {
    const cfg: ScatterConfig = { type: 'scatter', threeD: true }
    expect(
      getRenderableFields(cfg, {
        dimension: '2D',
        rendering3D: true,
        hasThreeDOption: true,
        hasZAxis: false,
      }).map((f) => f.key)
    ).toEqual(['sort', 'scale', 'showLabels', 'threeD', 'threeDVisualMap', 'threeDRotate', 'swap'])
  })

  it('returns 6 entries for a 3D scatter config (grouped x+y+z)', () => {
    const cfg: ScatterConfig = { type: 'scatter' }
    expect(
      getRenderableFields(cfg, { dimension: '3D', rendering3D: true, hasZAxis: true }).map(
        (f) => f.key
      )
    ).toEqual(['sort', 'scale', 'showLabels', 'threeDVisualMap', 'threeDRotate', 'swap'])
  })

  it('returns 4 entries for a 2D scatter config without value-3D active', () => {
    const cfg: ScatterConfig = { type: 'scatter' }
    expect(
      getRenderableFields(cfg, { dimension: '2D', rendering3D: false }).map((f) => f.key)
    ).toEqual(['sort', 'scale', 'showLabels', 'swap'])
  })

  it('partitions 3D fields into a dedicated section', () => {
    const cfg: BarConfig = { type: 'bar', threeD: true }
    const fields = getRenderableFields(cfg, {
      dimension: '2D',
      rendering3D: true,
      hasThreeDOption: true,
      hasZAxis: false,
    })
    const { general, threeD } = partitionRenderableFields(fields)
    expect(general.map((f) => f.key)).toEqual(['sort', 'scale', 'showLabels', 'swap'])
    expect(threeD.map((f) => f.key)).toEqual(['threeD', 'threeDVisualMap', 'threeDRotate'])
  })

  it('renders all applicable fields even when most keys are absent from the config', () => {
    const cfg = {
      type: 'bar',
      sort: { enabled: false, order: 'asc' as const },
    } as unknown as BarConfig
    expect(
      getRenderableFields(cfg, { dimension: '3D', rendering3D: true, hasZAxis: true }).map(
        (f) => f.key
      )
    ).toEqual(['sort', 'scale', 'showLabels', 'threeDVisualMap', 'threeDRotate', 'swap'])
  })
})
