import { describe, it, expect } from 'vitest'
import { filterDataSetSettings } from './filterDataSetSettings'
import type { DataSet } from '../types'

const dataset: DataSet = {
  name: 'Test',
  settings: [{ type: 'bar' }, { type: 'line' }, { type: 'pie' }],
  data: [],
}

describe('filterDataSetSettings', () => {
  it('returns dataset unchanged when allowed is empty', () => {
    expect(filterDataSetSettings(dataset, [])).toEqual(dataset)
    expect(filterDataSetSettings(dataset, undefined)).toEqual(dataset)
  })

  it('keeps only bundled chart types', () => {
    const filtered = filterDataSetSettings(dataset, ['bar'])
    expect(filtered.settings).toEqual([{ type: 'bar' }])
  })

  it('preserves original settings order', () => {
    const filtered = filterDataSetSettings(dataset, ['pie', 'bar'])
    expect(filtered.settings?.map((s) => s.type)).toEqual(['bar', 'pie'])
  })
})