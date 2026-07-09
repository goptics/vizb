import { describe, expect, it } from 'vitest'
import {
  CATEGORY_KEYS,
  DESCRIPTIVE_COLUMNS,
  columnsFromKeys,
  defaultSelectedKeys,
  keysForMath,
  nextSelectedKeys,
  resetSelectedKeys,
  selectedColumnSummary,
  sortKeyForVisibleColumns,
} from './descriptiveColumns'

describe('descriptiveColumns', () => {
  it('defaults to every descriptive column when math is absent or empty', () => {
    const allKeys = DESCRIPTIVE_COLUMNS.map((col) => col.key)

    expect(defaultSelectedKeys()).toEqual(allKeys)
    expect(defaultSelectedKeys([])).toEqual(allKeys)
  })

  it('maps stat math categories to ordered column keys', () => {
    expect(keysForMath(['counts', 'center'])).toEqual([
      ...CATEGORY_KEYS.counts,
      ...CATEGORY_KEYS.center,
    ])
  })

  it('ignores correlations when seeding descriptive columns with other categories', () => {
    expect(keysForMath(['correlations', 'spread'])).toEqual(CATEGORY_KEYS.spread)
  })

  it('keeps correlation-only math from enabling descriptive columns', () => {
    expect(keysForMath(['correlations'])).toEqual([])
    expect(defaultSelectedKeys(['correlations'])).toEqual([])
  })

  it('keeps canonical column order when resolving selected keys', () => {
    expect(columnsFromKeys(['median', 'count', 'stdDev']).map((col) => col.key)).toEqual([
      'count',
      'median',
      'stdDev',
    ])
  })

  it('prevents a combobox update from clearing the last selected column', () => {
    expect(nextSelectedKeys(['median'], [])).toEqual(['median'])
    expect(nextSelectedKeys(['median'], ['not-a-column'])).toEqual(['median'])
  })

  it('normalizes combobox updates to known keys in display order', () => {
    expect(nextSelectedKeys(['median'], ['stdDev', 'count', 'invalid'])).toEqual([
      'count',
      'stdDev',
    ])
  })

  it('resets to stat defaults, falling back to all columns for invalid defaults', () => {
    expect(resetSelectedKeys(['median', 'count'])).toEqual(['count', 'median'])
    expect(resetSelectedKeys(['not-a-column'])).toEqual(DESCRIPTIVE_COLUMNS.map((col) => col.key))
  })

  it('summarizes the trigger label for all, short, and longer selections', () => {
    expect(selectedColumnSummary(DESCRIPTIVE_COLUMNS.map((col) => col.key))).toBe('All columns')
    expect(selectedColumnSummary(['median'])).toBe('Median')
    expect(selectedColumnSummary(['mean', 'median'])).toBe('Mean, Median')
    expect(selectedColumnSummary(['count', 'mean', 'median'])).toBe('3 columns')
  })

  it('clears descriptive sort when its column is hidden, but keeps name sort', () => {
    expect(sortKeyForVisibleColumns('median', ['count', 'mean'])).toBeNull()
    expect(sortKeyForVisibleColumns('median', ['count', 'median'])).toBe('median')
    expect(sortKeyForVisibleColumns('name', ['count'])).toBe('name')
    expect(sortKeyForVisibleColumns(null, ['count'])).toBeNull()
  })
})
