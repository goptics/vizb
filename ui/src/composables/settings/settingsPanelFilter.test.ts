import { describe, it, expect } from 'vitest'
import type { Axis } from '@/types'
import type { RenderableField } from './fieldRegistry'
import { filterScatterDatasetFields } from './settingsPanelFilter'

const field = (key: RenderableField['key']): RenderableField => ({
  key,
  component: {} as RenderableField['component'],
})

const general = [
  field('sort'),
  field('scale'),
  field('showLabels'),
  field('swap'),
] as RenderableField[]

describe('filterScatterDatasetFields', () => {
  const valueAxes: Axis[] = [
    { key: 'x', type: 'value' },
    { key: 'y', type: 'value' },
  ]
  const hybridAxes: Axis[] = [
    { key: 'x', type: 'category' },
    { key: 'y', type: 'category' },
    { key: 'z', type: 'value' },
  ]

  it('keeps swap for scatter value mode', () => {
    expect(filterScatterDatasetFields(general, valueAxes).map((f) => f.key)).toEqual([
      'scale',
      'showLabels',
      'swap',
    ])
  })

  it('drops swap for hybrid mode', () => {
    expect(filterScatterDatasetFields(general, hybridAxes).map((f) => f.key)).toEqual([
      'scale',
      'showLabels',
    ])
  })
})
