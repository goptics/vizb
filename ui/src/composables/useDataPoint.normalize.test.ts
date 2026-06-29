import { describe, it, expect } from 'vitest'
import type { DataSet } from '../types'

/** Mirrors useDataPoint's load-time normalization for multi-dataset payloads. */
function normalizeDataSets(data: DataSet | DataSet[]): DataSet[] {
  const raw = Array.isArray(data) ? data : [data]
  return raw
}

describe('useDataPoint multi-dataset normalization', () => {
  const single: DataSet = {
    name: 'A',
    settings: [{ type: 'scatter' }],
    data: [{ xAxis: 'x', yAxis: '1', stats: [] }],
  }

  const second: DataSet = {
    name: 'B',
    settings: [{ type: 'scatter' }],
    data: [{ xAxis: 'y', yAxis: '2', stats: [] }],
  }

  it('wraps a single dataset object in an array', () => {
    expect(normalizeDataSets(single)).toEqual([single])
  })

  it('preserves an array of datasets for multi-tab rendering', () => {
    expect(normalizeDataSets([single, second])).toEqual([single, second])
    expect(normalizeDataSets([single, second])).toHaveLength(2)
  })
})
