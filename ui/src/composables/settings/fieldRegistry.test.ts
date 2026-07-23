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
vi.mock('../../components/settings/StackControl.vue', () => ({ default: { name: 'StackControl' } }))
vi.mock('../../components/settings/LabelModeControl.vue', () => ({
  default: { name: 'LabelModeControl' },
}))
vi.mock('../../components/settings/SmoothControl.vue', () => ({
  default: { name: 'SmoothControl' },
}))
vi.mock('../../components/settings/HorizontalControl.vue', () => ({
  default: { name: 'HorizontalControl' },
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
vi.mock('../../components/settings/VisualMapControl.vue', () => ({
  default: { name: 'VisualMapControl' },
}))
vi.mock('../../components/settings/SwapControl.vue', () => ({ default: { name: 'SwapControl' } }))

const { fieldRegistry, getControl, getRenderableFields, partitionRenderableFields } =
  await import('./fieldRegistry')

describe('fieldRegistry', () => {
  it('exposes the eleven known field controls', () => {
    expect(Object.keys(fieldRegistry).sort()).toEqual(
      [
        'horizontal',
        'threeDRotate',
        'scale',
        'stack',
        'labelMode',
        'smooth',
        'sort',
        'swap',
        'threeD',
        'threeDVisualMap',
        'visualMap',
      ].sort()
    )
  })

  it('getControl returns a component for a registered key', () => {
    expect(getControl('sort')).toBeDefined()
    expect(getControl('scale')).toBeDefined()
    expect(getControl('stack')).toBeDefined()
    expect(getControl('labelMode')).toBeDefined()
    expect(getControl('horizontal')).toBeDefined()
    expect(getControl('smooth')).toBeDefined()
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

  it('stack applies to 2D bar and line only', () => {
    expect(fieldRegistry['stack']!.appliesTo).toEqual(['bar', 'line'])
    expect(fieldRegistry['stack']!.appliesOn).toEqual(['2D'])
  })

  it('threeDRotate uses rendering3D visibility', () => {
    expect(fieldRegistry['threeDRotate']!.visible).toBeDefined()
  })

  it('smooth applies only to 2D line charts', () => {
    expect(fieldRegistry['smooth']!.appliesTo).toEqual(['line'])
    expect(fieldRegistry['smooth']!.visible?.({ rendering3D: false })).toBe(true)
    expect(fieldRegistry['smooth']!.visible?.({ rendering3D: true })).toBe(false)
  })

  it('fields with no appliesOn have no dimension constraint', () => {
    // Most fields (sort, scale, labelMode, swap) are available on any
    // dimension. Their `appliesOn` is undefined — the registry treats that as
    // "no constraint".
    for (const key of ['sort', 'scale', 'labelMode', 'swap'] as const) {
      expect(fieldRegistry[key]!.appliesOn).toBeUndefined()
    }
  })

  it('sort, labelMode, and swap apply to all six chart types', () => {
    for (const key of ['sort', 'labelMode', 'swap'] as const) {
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
  it('returns 6 entries for a 3D bar config (sort/scale/labelMode/threeDVisualMap/threeDRotate/swap)', () => {
    const cfg: BarConfig = { type: 'bar' }
    const fields = getRenderableFields(cfg, { dimension: '3D', rendering3D: true, hasZAxis: true })
    expect(fields.map((f) => f.key)).toEqual([
      'sort',
      'scale',
      'labelMode',
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
    ).toEqual(['sort', 'scale', 'labelMode', 'threeDVisualMap', 'threeDRotate', 'swap'])
  })

  it('returns 6 entries for a 2D bar config without value-3D active', () => {
    const cfg: BarConfig = { type: 'bar' }
    expect(
      getRenderableFields(cfg, { dimension: '2D', rendering3D: false }).map((f) => f.key)
    ).toEqual(['sort', 'scale', 'stack', 'labelMode', 'horizontal', 'swap'])
  })

  it('returns 7 entries for a 2D bar config with value-3D active', () => {
    const cfg: BarConfig = { type: 'bar', threeD: true }
    expect(
      getRenderableFields(cfg, {
        dimension: '2D',
        rendering3D: true,
        hasThreeDOption: true,
        hasZAxis: false,
      }).map((f) => f.key)
    ).toEqual(['sort', 'scale', 'labelMode', 'threeD', 'threeDVisualMap', 'threeDRotate', 'swap'])
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
    ).toEqual(['sort', 'scale', 'labelMode', 'threeDVisualMap', 'threeDRotate', 'swap'])
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
    ).toEqual(['sort', 'scale', 'labelMode', 'horizontal', 'threeD', 'swap'])
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

  it('returns 6 entries for a 2D line config without value-3D active', () => {
    const cfg: LineConfig = { type: 'line' }
    expect(
      getRenderableFields(cfg, { dimension: '2D', rendering3D: false }).map((f) => f.key)
    ).toEqual(['sort', 'scale', 'stack', 'labelMode', 'smooth', 'swap'])
  })

  it('hides stack in value and mixed transform modes', () => {
    const cfg: BarConfig = { type: 'bar' }
    expect(
      getRenderableFields(cfg, {
        dimension: '2D',
        rendering3D: false,
        chartMode: 'value',
      }).map((f) => f.key)
    ).not.toContain('stack')
    expect(
      getRenderableFields(cfg, {
        dimension: '2D',
        rendering3D: false,
        chartMode: 'mixed',
      }).map((f) => f.key)
    ).not.toContain('stack')
  })

  it('returns 3 entries for a pie config (no scale/threeDRotate; dimension is irrelevant)', () => {
    const cfg: PieConfig = { type: 'pie' }
    expect(getRenderableFields(cfg, { dimension: '2D' }).map((f) => f.key)).toEqual([
      'sort',
      'labelMode',
      'swap',
    ])
  })

  it('returns 3 entries for a heatmap config (no scale/threeDRotate; dimension is irrelevant)', () => {
    const cfg: HeatmapConfig = { type: 'heatmap' }
    expect(getRenderableFields(cfg, { dimension: '2D' }).map((f) => f.key)).toEqual([
      'sort',
      'labelMode',
      'swap',
    ])
  })

  it('returns 3 entries for a radar config (no scale/threeDRotate; dimension is irrelevant)', () => {
    const cfg: RadarConfig = { type: 'radar' }
    expect(getRenderableFields(cfg, { dimension: '2D' }).map((f) => f.key)).toEqual([
      'sort',
      'labelMode',
      'swap',
    ])
  })

  it('treats an unknown dimension (no ctx) as "no dimension constraint" — shows all applicable fields', () => {
    const cfg: BarConfig = { type: 'bar' }
    expect(getRenderableFields(cfg).map((f) => f.key)).toEqual([
      'sort',
      'scale',
      'stack',
      'labelMode',
      'horizontal',
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
    ).toEqual(['sort', 'scale', 'labelMode', 'threeD', 'threeDVisualMap', 'threeDRotate', 'swap'])
  })

  it('returns 6 entries for a 3D scatter config (grouped x+y+z)', () => {
    const cfg: ScatterConfig = { type: 'scatter' }
    expect(
      getRenderableFields(cfg, { dimension: '3D', rendering3D: true, hasZAxis: true }).map(
        (f) => f.key
      )
    ).toEqual(['sort', 'scale', 'labelMode', 'threeDVisualMap', 'threeDRotate', 'swap'])
  })

  it('returns 5 entries for a 2D scatter config without value-3D active', () => {
    const cfg: ScatterConfig = { type: 'scatter' }
    expect(
      getRenderableFields(cfg, { dimension: '2D', rendering3D: false }).map((f) => f.key)
    ).toEqual(['sort', 'scale', 'labelMode', 'visualMap', 'swap'])
  })

  it('hides 2D visualMap on 3D scatter', () => {
    const cfg: ScatterConfig = { type: 'scatter' }
    expect(
      getRenderableFields(cfg, { dimension: '3D', rendering3D: true, hasZAxis: true }).map(
        (f) => f.key
      )
    ).not.toContain('visualMap')
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
    expect(general.map((f) => f.key)).toEqual(['sort', 'scale', 'labelMode', 'swap'])
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
    ).toEqual(['sort', 'scale', 'labelMode', 'threeDVisualMap', 'threeDRotate', 'swap'])
  })
})
