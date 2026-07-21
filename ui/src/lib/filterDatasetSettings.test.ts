import { describe, it, expect } from 'vitest'
import { filterDatasetSettings } from './filterDatasetSettings'
import type { Dataset } from '../types'

const dataset: Dataset = {
  name: 'Test',
  settings: [{ type: 'bar' }, { type: 'line' }, { type: 'pie' }],
  data: [],
}

describe('filterDatasetSettings', () => {
  it('returns dataset unchanged when allowed is empty', () => {
    expect(filterDatasetSettings(dataset, [])).toEqual(dataset)
    expect(filterDatasetSettings(dataset, undefined)).toEqual(dataset)
  })

  it('keeps only bundled chart types', () => {
    const filtered = filterDatasetSettings(dataset, ['bar'])
    expect(filtered.settings).toEqual([{ type: 'bar' }])
  })

  it('preserves original settings order', () => {
    const filtered = filterDatasetSettings(dataset, ['pie', 'bar'])
    expect(filtered.settings?.map((s) => s.type)).toEqual(['bar', 'pie'])
  })
})
