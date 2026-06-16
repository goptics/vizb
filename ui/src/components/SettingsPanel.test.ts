import { describe, it, expect } from 'vitest'
import type { BarConfig, LineConfig, PieConfig, HeatmapConfig, RadarConfig } from '../types'
import { getRenderableFields } from '../composables/settings/fieldRegistry'

// SettingsPanel.vue is declarative: it walks `Object.keys(activeConfig)` via
// `getRenderableFields` and renders the registered control for each match. The
// rendering itself is a one-line `<component :is="...">` loop — there is no
// chart-type branching to test. So the panel's contract is fully exercised by
// asserting which field keys make it through `getRenderableFields` for each
// chart type. If the matrix below regresses, the panel renders the wrong set
// of controls.

describe('SettingsPanel field selection', () => {
  it('renders 5 controls for a bar config (sort/scale/showLabels/autoRotate/swap)', () => {
    const cfg: BarConfig = {
      type: 'bar',
      sort: { enabled: false, order: 'asc' },
      scale: 'linear',
      showLabels: false,
      autoRotate: false,
      swap: '',
    }
    expect(getRenderableFields(cfg).map((f) => f.key)).toEqual([
      'sort',
      'scale',
      'showLabels',
      'autoRotate',
      'swap',
    ])
  })

  it('renders 5 controls for a line config (sort/scale/showLabels/autoRotate/swap)', () => {
    const cfg: LineConfig = {
      type: 'line',
      sort: { enabled: false, order: 'asc' },
      scale: 'linear',
      showLabels: false,
      autoRotate: false,
      swap: '',
    }
    expect(getRenderableFields(cfg).map((f) => f.key)).toEqual([
      'sort',
      'scale',
      'showLabels',
      'autoRotate',
      'swap',
    ])
  })

  it('renders 3 controls for a pie config (no scale/autoRotate)', () => {
    const cfg: PieConfig = {
      type: 'pie',
      sort: { enabled: false, order: 'asc' },
      showLabels: false,
      swap: '',
    }
    expect(getRenderableFields(cfg).map((f) => f.key)).toEqual(['sort', 'showLabels', 'swap'])
  })

  it('renders 3 controls for a heatmap config (no scale/autoRotate)', () => {
    const cfg: HeatmapConfig = {
      type: 'heatmap',
      sort: { enabled: false, order: 'asc' },
      showLabels: false,
      swap: '',
    }
    expect(getRenderableFields(cfg).map((f) => f.key)).toEqual(['sort', 'showLabels', 'swap'])
  })

  it('renders 3 controls for a radar config (no scale/autoRotate)', () => {
    const cfg: RadarConfig = {
      type: 'radar',
      sort: { enabled: false, order: 'asc' },
      showLabels: false,
      swap: '',
    }
    expect(getRenderableFields(cfg).map((f) => f.key)).toEqual(['sort', 'showLabels', 'swap'])
  })

  it('omits keys whose value is absent from the config (no defaults injected)', () => {
    const cfg: BarConfig = { type: 'bar', sort: { enabled: false, order: 'asc' } }
    expect(getRenderableFields(cfg).map((f) => f.key)).toEqual(['sort'])
  })
})
