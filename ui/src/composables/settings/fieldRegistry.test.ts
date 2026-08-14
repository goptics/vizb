import { describe, it, expect } from 'vitest'
import type {
  BarConfig,
  LineConfig,
  ScatterConfig,
  PieConfig,
  HeatmapConfig,
  RadarConfig,
  SankeyConfig,
  ChordConfig,
} from '@/types'
// Side-effect: top-level vi.mock for every settings control SFC fieldRegistry imports.
import '@/test-utils/mockSettingsControls'

const { fieldRegistry, getRenderableFields, partitionRenderableFields } = await import(
  './fieldRegistry'
)

const BOOLEAN_KEYS = [
  'stack',
  'showLabels',
  'smooth',
  'horizontal',
  'threeD',
  'threeDVisualMap',
  'visualMap',
  'threeDRotate',
] as const

describe('fieldRegistry', () => {
  it('exposes the eleven known field controls', () => {
    expect(Object.keys(fieldRegistry).sort()).toEqual(
      [
        'horizontal',
        'threeDRotate',
        'scale',
        'stack',
        'showLabels',
        'smooth',
        'sort',
        'swap',
        'threeD',
        'threeDVisualMap',
        'visualMap',
      ].sort()
    )
  })

  it('boolean fields share BooleanControl and carry toggle copy', () => {
    for (const key of BOOLEAN_KEYS) {
      expect(fieldRegistry[key]!.component).toBe(fieldRegistry.stack.component)
      expect(fieldRegistry[key]!.id).toBeTruthy()
      expect(fieldRegistry[key]!.label).toBeTruthy()
      expect(fieldRegistry[key]!.description).toBeTruthy()
    }
    expect(fieldRegistry.sort.component).not.toBe(fieldRegistry.stack.component)
    expect(fieldRegistry.scale.component).not.toBe(fieldRegistry.stack.component)
    expect(fieldRegistry.swap.component).not.toBe(fieldRegistry.stack.component)
    expect(fieldRegistry.stack).toMatchObject({
      id: 'stack-switch',
      label: 'Stack series',
      description: 'Stack grouped bar or line series into category totals.',
      separator: true,
    })
    expect(fieldRegistry.showLabels).toMatchObject({
      id: 'labels-switch',
      separator: true,
    })
    expect(fieldRegistry.smooth).toMatchObject({
      id: 'smooth-lines-switch',
      separator: true,
    })
    expect(fieldRegistry.horizontal).toMatchObject({
      id: 'horizontal-bars-switch',
      separator: true,
    })
    expect(fieldRegistry.threeD).toMatchObject({
      id: 'three-d-switch',
      label: '3D view',
    })
    expect(fieldRegistry.threeD.separator).toBeUndefined()
    expect(fieldRegistry.threeDVisualMap.separator).toBeUndefined()
    expect(fieldRegistry.threeDRotate.separator).toBeUndefined()
    expect(fieldRegistry.visualMap.separator).toBeUndefined()
  })

  it('scale and threeDRotate apply to bar, line, and scatter', () => {
    expect(fieldRegistry['scale']!.appliesTo).toEqual(['bar', 'line', 'scatter'])
    expect(fieldRegistry['threeDRotate']!.appliesTo).toEqual(['bar', 'line', 'scatter'])
  })

  it('stack applies to 2D bar and line only', () => {
    expect(fieldRegistry['stack']!.appliesTo).toEqual(['bar', 'line'])
    expect(fieldRegistry['stack']!.visible?.({ rendering3D: false, dimension: '2D' })).toBe(true)
    expect(fieldRegistry['stack']!.visible?.({ rendering3D: true })).toBe(false)
    expect(fieldRegistry['stack']!.visible?.({ dimension: '3D' })).toBe(false)
  })

  it('threeDRotate uses rendering3D visibility', () => {
    expect(fieldRegistry['threeDRotate']!.visible).toBeDefined()
  })

  it('smooth applies only to 2D line charts', () => {
    expect(fieldRegistry['smooth']!.appliesTo).toEqual(['line'])
    expect(fieldRegistry['smooth']!.visible?.({ rendering3D: false })).toBe(true)
    expect(fieldRegistry['smooth']!.visible?.({ rendering3D: true })).toBe(false)
  })

  it('sort, showLabels, and swap apply to all eight chart types', () => {
    for (const key of ['sort', 'showLabels', 'swap'] as const) {
      expect(fieldRegistry[key]!.appliesTo).toEqual([
        'bar',
        'line',
        'scatter',
        'pie',
        'heatmap',
        'radar',
        'sankey',
        'chord',
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

  it('copies boolean toggle props onto renderable fields', () => {
    const cfg: BarConfig = { type: 'bar' }
    const fields = getRenderableFields(cfg, { dimension: '2D', rendering3D: false })
    expect(fields.find((f) => f.key === 'stack')).toMatchObject({
      id: 'stack-switch',
      label: 'Stack series',
      description: 'Stack grouped bar or line series into category totals.',
      separator: true,
    })
    expect(fields.find((f) => f.key === 'sort')?.id).toBeUndefined()
  })

  it('returns 6 entries for a 3D line config', () => {
    const cfg: LineConfig = { type: 'line' }
    expect(
      getRenderableFields(cfg, { dimension: '3D', rendering3D: true, hasZAxis: true }).map(
        (f) => f.key
      )
    ).toEqual(['sort', 'scale', 'showLabels', 'threeDVisualMap', 'threeDRotate', 'swap'])
  })

  it('returns 6 entries for a 2D bar config without value-3D active', () => {
    const cfg: BarConfig = { type: 'bar' }
    expect(
      getRenderableFields(cfg, { dimension: '2D', rendering3D: false }).map((f) => f.key)
    ).toEqual(['sort', 'scale', 'stack', 'showLabels', 'horizontal', 'swap'])
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
    ).toEqual(['sort', 'scale', 'showLabels', 'horizontal', 'threeD', 'swap'])
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
    ).toEqual(['sort', 'scale', 'stack', 'showLabels', 'smooth', 'swap'])
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

  it('returns 3 entries for a sankey config (no scale/threeDRotate; dimension is irrelevant)', () => {
    const cfg: SankeyConfig = { type: 'sankey' }
    expect(getRenderableFields(cfg, { dimension: '2D' }).map((f) => f.key)).toEqual([
      'sort',
      'showLabels',
      'swap',
    ])
  })

  it('returns 3 entries for a chord config (no scale/threeDRotate; dimension is irrelevant)', () => {
    const cfg: ChordConfig = { type: 'chord' }
    expect(getRenderableFields(cfg).map((f) => f.key)).toEqual(['sort', 'showLabels', 'swap'])
  })

  it('treats an unknown dimension (no ctx) as "no dimension constraint" — shows all applicable fields', () => {
    const cfg: BarConfig = { type: 'bar' }
    expect(getRenderableFields(cfg).map((f) => f.key)).toEqual([
      'sort',
      'scale',
      'stack',
      'showLabels',
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

  it('returns 5 entries for a 2D scatter config without value-3D active', () => {
    const cfg: ScatterConfig = { type: 'scatter' }
    expect(
      getRenderableFields(cfg, { dimension: '2D', rendering3D: false }).map((f) => f.key)
    ).toEqual(['sort', 'scale', 'showLabels', 'visualMap', 'swap'])
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

  it('omits missing THREE_D_FIELD_KEYS from the threeD partition', () => {
    const { general, threeD } = partitionRenderableFields([
      { key: 'sort', component: fieldRegistry.sort.component },
      { key: 'threeDRotate', component: fieldRegistry.threeDRotate.component },
    ])
    expect(general.map((f) => f.key)).toEqual(['sort'])
    expect(threeD.map((f) => f.key)).toEqual(['threeDRotate'])
  })
})
